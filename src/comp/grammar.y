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

var gStore Store
var gLex   *lexer

var gMem   *Mem
var gLoad  string
var gComp  Comp
var gHead  []string
var gError *ParseError
%}

%union {
	str  string
	num  float64
	expr Expr
	args []Expr
}

%token TK_EQ    // "=="
%token TK_NEQ   // "!="
%token TK_MATCH // "=~"
%token TK_PROD  // "<-"
%token TK_LTEQ  // "<="
%token TK_GTEQ  // ">="
%token TK_CAT   // "++"
%token TK_AND   // "&&"
%token TK_OR    // "||"

%token <num> NUMBER
%token <str> IDENT
%token <str> STRING

%type <str>  identifier
%type <str>  generator
%type <expr> primary_expression
%type <expr> postfix_expression
%type <expr> unary_expression
%type <expr> multiplicative_expression
%type <expr> additive_expression
%type <expr> relational_expression
%type <expr> equality_expression
%type <expr> expression
%type <args> expression_list

%%

comprehension:
	{
		gLoad = ""
		gComp = nil
		gHead = nil
	}
    | '[' expression_list '|' generator ']'
	{
		gLoad = $4
		gComp = Return(Reflect, $2)
		gHead = ExprHead($2)
	}
    | '[' expression_list '|' generator ',' expression ']'
	{
		gLoad = $4
		gComp = Return(Select(Reflect, $6), $2)
		gHead = ExprHead($2)
	}
    ;

generator:
      IDENT TK_PROD IDENT
	{
		if !gStore.IsDef($3) {
			parseError("unknown dataset %v", $3)
		} else {
			gStore.Declare(gMem, $1, $3)
			$$ = $3
		}
	}
    ;

primary_expression:
      STRING
	{
		$$ = ExprValue($1)
	}
    | NUMBER
	{
		$$ = ExprValue($1)
	}
    | '(' expression ')'
	{
		$$ = $2
	}
    ;

identifier:
      IDENT
	{
		$$ = $1
	}
    | identifier '.' IDENT
	{
		$$ = fmt.Sprintf("%v.%v", $1, $3)
	}
    ;

postfix_expression:
      primary_expression
	{
		$$ = $1
	}
    | identifier
	{
		$$ = ExprAttr($1, gMem.AttrPos($1))
	}
    | identifier '(' ')'
	{
		expr, err := ExprFunc($1, nil)
		if err == nil {
			$$ = expr
		} else {
			parseError(err.Error())
		}
	}
    | identifier '(' expression_list ')'
	{
		expr, err := ExprFunc($1, $3)
		if err == nil {
			$$ = expr
		} else {
			parseError(err.Error())
		}
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
    | additive_expression TK_CAT multiplicative_expression
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
    | relational_expression TK_LTEQ additive_expression
	{
		$$ = $1.LTE($3)
	}
    | relational_expression TK_GTEQ additive_expression
	{
		$$ = $1.GTE($3)
	}
    ;

equality_expression:
      relational_expression
	{
		$$ = $1
	}
    | equality_expression TK_EQ relational_expression
	{
		$$ = $1.Eq($3)
	}
    | equality_expression TK_NEQ relational_expression
	{
		$$ = $1.NotEq($3)
	}
    | equality_expression TK_MATCH STRING
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
    | expression TK_AND equality_expression
	{
		$$ = $1.And($3)
	}
    | expression TK_OR equality_expression
	{
		$$ = $1.Or($3)
	}
    ;

%%

func parseError(s string, v ...interface{}) {
	gError = NewError(gLex.scan.Pos().Line, gLex.scan.Pos().Column, s, v...)
}

type lexer struct {
	scan scanner.Scanner
}

func (l *lexer) Lex(yylval *comp_SymType) int {
	tok := l.scan.Scan()
	switch tok {
	case scanner.Ident:
		yylval.str = l.scan.TokenText()
		return IDENT
	case scanner.Int, scanner.Float:
		yylval.num, _ = strconv.ParseFloat(l.scan.TokenText(), 64)
		return NUMBER
	case scanner.String, scanner.RawString:
		yylval.str = l.scan.TokenText()
		yylval.str = yylval.str[1:len(yylval.str)-1]
		return STRING
	case '<':
		if l.scan.Peek() == '-' {
			l.scan.Next()
			return TK_PROD
		} else if l.scan.Peek() == '=' {
			l.scan.Next()
			return TK_LTEQ
		}
		return '<'
	case '>':
		if l.scan.Peek() == '=' {
			l.scan.Next()
			return TK_GTEQ
		}
		return '>'
	case '=':
		if l.scan.Peek() == '=' {
			l.scan.Next()
			return TK_EQ
		} else if l.scan.Peek() == '~' {
			l.scan.Next()
			return TK_MATCH
		}
		return '='
	case '+':
		if l.scan.Peek() == '+' {
			l.scan.Next()
			return TK_CAT
		}
		return '+'
	case '!':
		if l.scan.Peek() == '=' {
			l.scan.Next()
			return TK_NEQ
		}
		return '!'
	case '&':
		if l.scan.Peek() == '&' {
			l.scan.Next()
			return TK_AND
		}
		return '&'
	case '|':
		if l.scan.Peek() == '|' {
			l.scan.Next()
			return TK_OR
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

func Parse(query string, store Store) (*Mem, string, Comp, []string, *ParseError) {
	gMutex.Lock()
	defer gMutex.Unlock()

	gStore = store
	gLex = &lexer{}

	gMem = NewMem()
	gLoad = ""
	gComp = nil
	gHead = nil
	gError = nil

	reader := strings.NewReader(query)
	gLex.scan.Init(reader)
	comp_Parse(gLex)

	if gError == nil {
		bad := gMem.BadAttrs()
		if len(bad) > 0 {
			gError = NewError(0, 0, "unknown identifier(s): %v", bad)
		}
	}

	return gMem, gLoad, gComp, gHead, gError
}
