%{
package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/scanner"
)

type ParseError struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Error  string `json:"error"`
}

func NewError(line, column int, msg string, args ...interface{}) *ParseError {
	return &ParseError{Line: line, Column: column, Error: fmt.Sprintf(msg, args...)}
}

var gMutex sync.Mutex

var gDecls *Decls
var gMem   *Mem
var gLex   *lexer

var gExpr  Expr
var gError *ParseError
%}

%union {
	str  string
	num  float64
	expr Expr
	args []Expr
}

%token EQ    // "=="
%token NEQ   // "!="
%token MATCH // "=~"
%token PROD  // "<-"
%token LTE   // "<="
%token GTE   // ">="
%token CAT   // "++"
%token AND   // "&&"
%token OR    // "||"
%token TRUE  // "true"
%token FALSE // "false"

%token <num> NUMBER
%token <str> IDENT
%token <str> STRING

%type <expr> primary_expression
%type <expr> postfix_expression
%type <expr> unary_expression
%type <expr> multiplicative_expression
%type <expr> additive_expression
%type <expr> relational_expression
%type <expr> equality_expression
%type <expr> expression
%type <args> expression_list
%type <expr> object_field
%type <args> object_field_list

%start program

%%

program:
      expression
	{
		gExpr = $1
	}
    ;

primary_expression:
      STRING
	{
		$$ = ExprConst(String($1))
	}
    | NUMBER
	{
		$$ = ExprConst(Number($1))
	}
    | TRUE
	{
		$$ = ExprConst(Bool(true))
	}
    | FALSE
	{
		$$ = ExprConst(Bool(false))
	}
    | IDENT
	{
		gDecls.Use($1)
		$$ = ExprLoad($1)
	}
    | '{' object_field_list '}'
	{
		$$ = ExprObject($2)
	}
    | '[' expression_list ']'
	{
		$$ = ExprList($2)
	}
    | '[' expression_list '|' IDENT PROD expression ']'
	{
		gDecls.Declare($4)
		if len($2) == 1 {
			$$ = ExprLoop($4, $6, $2[0])
		} else {
			$$ = ExprLoop($4, $6, ExprObject($2))
		}
	}
/*
    | '[' expression_list '|' IDENT PROD expression ',' expression ']'
	{
		$$ = ExprLoop($4, $6, ExprSelect($8).Return($2))
	}
*/
    | '(' expression ')'
	{
		$$ = $2
	}
    ;

object_field_list:
      object_field
	{
		$$ = []Expr{$1}
	}
    | object_field_list ',' object_field
	{
		$$ = append($1, $3)
	}
    ;

object_field:
      expression
	{
		$$ = $1
	}
    | IDENT ':' expression
	{
		$$ = Expr{$1, $3.Eval}
	}
    ;

expression_list:
      expression
	{
		$$ = []Expr{$1}
	}
    | expression_list ',' expression
	{
		$$ = append($1, $3)
	}
    ;

postfix_expression:
      primary_expression
	{
		$$ = $1
	}
    | postfix_expression '.' IDENT
	{
		$$ = $1.Field($3)
	}
    | postfix_expression '(' ')'
	{
		expr, err := $1.Call(nil)
		if err == nil {
			gDecls.Reset($1.Name)
			$$ = expr
		} else {
			parseError(err.Error())
		}
	}
    | postfix_expression '(' expression_list ')'
	{
		expr, err := $1.Call($3)
		if err == nil {
			gDecls.Reset($1.Name)
			$$ = expr
		} else {
			parseError(err.Error())
		}
	}
    ;

unary_expression:
      postfix_expression
	{
		$$ = $1
	}
    | '!' postfix_expression
	{
		$$ = $2.Not()
	}
    | '-' postfix_expression
	{
		$$ = $2.Neg()
	}
    | '+' postfix_expression
	{
		$$ = $2.Pos()
	}
    ;

multiplicative_expression:
      unary_expression
	{
		$$ = $1
	}
    | multiplicative_expression '*' unary_expression
	{
		$$ = $1.Mul($3)
	}
    | multiplicative_expression '/' unary_expression
	{
		$$ = $1.Div($3)
	}
    ;

additive_expression:
      multiplicative_expression
	{
		$$ = $1
	}
    | additive_expression '+' multiplicative_expression
	{
		$$ = $1.Add($3)
	}
    | additive_expression '-' multiplicative_expression
	{
		$$ = $1.Sub($3)
	}
    | additive_expression CAT multiplicative_expression
	{
		$$ = $1.Cat($3)
	}
    ;

relational_expression:
      additive_expression
	{
		$$ = $1
	}
    | relational_expression '<' additive_expression
	{
		$$ = $1.LT($3)
	}
    | relational_expression '>' additive_expression
	{
		$$ = $1.GT($3)
	}
    | relational_expression LTE additive_expression
	{
		$$ = $1.LTE($3)
	}
    | relational_expression GTE additive_expression
	{
		$$ = $1.GTE($3)
	}
    ;

equality_expression:
      relational_expression
	{
		$$ = $1
	}
    | equality_expression EQ relational_expression
	{
		$$ = $1.Eq($3)
	}
    | equality_expression NEQ relational_expression
	{
		$$ = $1.NotEq($3)
	}
    | equality_expression MATCH STRING
	{
		re, err := gMem.RegExp($3)
		if err == nil {
			$$ = $1.Match(re)
		} else {
			parseError("%v", err)
		}
	}
    ;

expression:
      equality_expression
	{
		$$ = $1
	}
    | expression AND equality_expression
	{
		$$ = $1.And($3)
	}
    | expression OR equality_expression
	{
		$$ = $1.Or($3)
	}
    ;

%%

type lexer struct {
	scan scanner.Scanner
}

func (l *lexer) Lex(yylval *comp_SymType) int {
	tok := l.scan.Scan()
	switch tok {
	case scanner.Ident:
		ident := l.scan.TokenText()
		if ident == "true" {
			return TRUE
		} else if ident == "false" {
			return FALSE
		}

		yylval.str = ident
		return IDENT
	case scanner.Int, scanner.Float:
		yylval.num, _ = strconv.ParseFloat(l.scan.TokenText(), 64)
		return NUMBER
	case scanner.String, scanner.RawString:
		yylval.str = l.scan.TokenText()
		yylval.str = yylval.str[1 : len(yylval.str)-1]
		return STRING
	case '<':
		if l.scan.Peek() == '-' {
			l.scan.Next()
			return PROD
		} else if l.scan.Peek() == '=' {
			l.scan.Next()
			return LTE
		}
		return '<'
	case '>':
		if l.scan.Peek() == '=' {
			l.scan.Next()
			return GTE
		}
		return '>'
	case '=':
		if l.scan.Peek() == '=' {
			l.scan.Next()
			return EQ
		} else if l.scan.Peek() == '~' {
			l.scan.Next()
			return MATCH
		}
		return '='
	case '+':
		if l.scan.Peek() == '+' {
			l.scan.Next()
			return CAT
		}
		return '+'
	case '!':
		if l.scan.Peek() == '=' {
			l.scan.Next()
			return NEQ
		}
		return '!'
	case '&':
		if l.scan.Peek() == '&' {
			l.scan.Next()
			return AND
		}
		return '&'
	case '|':
		if l.scan.Peek() == '|' {
			l.scan.Next()
			return OR
		}
		return '|'
	default:
		return int(tok)
	}

	return 0
}

func (l *lexer) Error(s string) {
	parseError(s)
}

func parseError(s string, v ...interface{}) {
	gError = NewError(gLex.scan.Pos().Line, gLex.scan.Pos().Column, s, v...)
}

func Compile(expr string, mem *Mem) (Expr, *ParseError) {
	gMutex.Lock()
	defer gMutex.Unlock()

	gMem = mem
	gDecls = mem.Decls()
	gLex = &lexer{}

	gError = nil
	gExpr = BadExpr

	reader := strings.NewReader(expr)
	gLex.scan.Init(reader)
	comp_Parse(gLex)

	if gError == nil {
		if bad := gDecls.Unknown(); len(bad) > 0 {
			gError = NewError(0, 0, "unknown identifier(s): %v", bad)
		}
	}

	return gExpr, gError
}
