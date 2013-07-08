// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this
// source code under the terms of the MIT License found in the LICENSE file.

%{
package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"text/scanner"
)

var gMutex sync.Mutex

var gDecls *Decls
var gLex   *lexer

var gLID   int
var gExpr  Expr
var gError *ParseError
%}

%union {
	str   string
	num   float64
	bin   bool
	expr  Expr
	exprs []Expr
	loop  *Loop
}

%token EQ	// "=="
%token NEQ	// "!="
%token MATCH	// "=~"
%token PROD_SEQ	// "<-"
%token PROD_PAR	// "<~"
%token LTE	// "<="
%token GTE	// ">="
%token CAT	// "++"
%token AND	// "&&"
%token OR	// "||"
%token TRUE	// "true"
%token FALSE	// "false"

%token <num> NUMBER
%token <str> IDENT
%token <str> STRING

%type <expr>	primary_expression
%type <expr>	postfix_expression
%type <expr>	unary_expression
%type <expr>	multiplicative_expression
%type <expr>	additive_expression
%type <expr>	relational_expression
%type <expr>	equality_expression
%type <expr>	expression
%type <exprs>	expression_list
%type <exprs>	expression_list_or_empty
%type <expr>	object_field
%type <exprs>	object_field_list
%type <bin>	generator
%type <loop>	generator_list

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
		addr := gDecls.Declare("", String($1), ScalarType(0))
		$$ = ExprLoad(strconv.Quote($1), addr)
		gDecls.SetType($$, ScalarType(0))
	}
    | NUMBER
	{
		addr := gDecls.Declare("", Number($1), ScalarType(0))
		$$ = ExprLoad(fmt.Sprintf("%v", $1), addr)
		gDecls.SetType($$, ScalarType(0))
	}
    | TRUE
	{
		addr := gDecls.Declare("", Bool(true), ScalarType(0))
		$$ = ExprLoad("true", addr)
		gDecls.SetType($$, ScalarType(0))
	}
    | FALSE
	{
		addr := gDecls.Declare("", Bool(false), ScalarType(0))
		$$ = ExprLoad("false", addr)
		gDecls.SetType($$, ScalarType(0))
	}
    | IDENT
	{
		addr := gDecls.UseIdent($1)
		$$ = ExprLoad($1, addr)
		gDecls.SetType($$, TypeOfIdent($1))
	}
    | '{' object_field_list '}'
	{
		ot := make(ObjectType, len($2))
		for i, f := range $2 {
			ot[i].Type = TypeOfExpr(f.Id)
			ot[i].Name = f.Name
		}
		$$ = ExprObject($2)
		gDecls.SetType($$, ot)
	}
    | '[' expression_list ']'
	{
		eids := make([]int64, 0)
		for _, e := range $2 {
			eids = append(eids, e.Id)
		}
		$$ = ExprList($2)
		gDecls.SameType(eids)
		gDecls.SetType($$, ListType{TypeOfExpr(eids[0])})
	}
    | '[' expression '|' generator_list ']'
	{
		gDecls.Strict(false)
		resType := ListType{TypeOfExpr($2.Id)}
		resAddr := gDecls.Declare("", nil, resType)
		loop := $4.Return($2, resAddr)
		$$ = ExprComp(loop, resAddr)
		gDecls.SetType($$, resType)
	}
    | '(' expression ')'
	{
		$$ = $2
	}
    ;

generator_list:
      IDENT generator expression
	{
		gDecls.Strict(true)
		varAddr := gDecls.Declare($1, nil, TypeOfElem($3.Id))
		$$ = ForEach(gLID, varAddr, $3, $2)
		gLID++
	}
    | generator_list ',' expression
	{
		$$ = $1.Select($3)
	}
    | generator_list ',' IDENT generator expression
	{
		varAddr := gDecls.Declare($3, nil, TypeOfElem($5.Id))
		$$ = $1.Nest(gLID, varAddr, $5, $4)
		gLID++
	}
    ;

generator:
      PROD_SEQ { $$ = false }
    | PROD_PAR { $$ = true }
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
		$$ = Expr{$3.Id, $1, $3.Code}
	}
    ;

expression_list_or_empty:
	{
		$$ = nil
	}
    | expression_list
	{
		$$ = $1
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
		pos := gDecls.UseField($1.Id, $3)
		$$ = $1.Field($3, pos)
		gDecls.SetType($$, TypeOfField{$1.Id, $3})
	}
    | postfix_expression '[' STRING ']'
	{
		pos := gDecls.UseField($1.Id, $3)
		$$ = $1.Field($3, pos)
		gDecls.SetType($$, TypeOfField{$1.Id, $3})
	}
    | postfix_expression '(' expression_list_or_empty ')'
	{
		eids := make([]int64, len($3))
		for i, e := range $3 {
			eids[i] = e.Id
		}
		fn := gDecls.UseFunc($1.Name, eids)
		if fn > -1 {
			$$ = ExprCall(fn, $3)
			gDecls.SetType($$, TypeOfFunc($1.Name))
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
		$$ = $2.Unary(OpNot(), "!")
		gDecls.SetType($$, ScalarType(0))
	}
    | '-' postfix_expression
	{
		$$ = $2.Unary(OpNeg(), "-")
		gDecls.SetType($$, ScalarType(0))
	}
    | '+' postfix_expression
	{
		$$ = $2.Unary(OpPos(), "+")
		gDecls.SetType($$, ScalarType(0))
	}
    ;

multiplicative_expression:
      unary_expression
	{
		$$ = $1
	}
    | multiplicative_expression '*' unary_expression
	{
		$$ = $1.Binary($3, OpMul(), "*")
		gDecls.SetType($$, ScalarType(0))
	}
    | multiplicative_expression '/' unary_expression
	{
		$$ = $1.Binary($3, OpDiv(), "/")
		gDecls.SetType($$, ScalarType(0))
	}
    ;

additive_expression:
      multiplicative_expression
	{
		$$ = $1
	}
    | additive_expression '+' multiplicative_expression
	{
		$$ = $1.Binary($3, OpAdd(), "+")
		gDecls.SetType($$, ScalarType(0))
	}
    | additive_expression '-' multiplicative_expression
	{
		$$ = $1.Binary($3, OpSub(), "-")
		gDecls.SetType($$, ScalarType(0))
	}
    | additive_expression CAT multiplicative_expression
	{
		$$ = $1.Binary($3, OpCat(), "++")
		gDecls.SetType($$, ScalarType(0))
	}
    ;

relational_expression:
      additive_expression
	{
		$$ = $1
	}
    | relational_expression '<' additive_expression
	{
		$$ = $1.Binary($3, OpLT(), "<")
		gDecls.SetType($$, ScalarType(0))
	}
    | relational_expression '>' additive_expression
	{
		$$ = $1.Binary($3, OpGT(), ">")
		gDecls.SetType($$, ScalarType(0))
	}
    | relational_expression LTE additive_expression
	{
		$$ = $1.Binary($3, OpLTE(), "<=")
		gDecls.SetType($$, ScalarType(0))
	}
    | relational_expression GTE additive_expression
	{
		$$ = $1.Binary($3, OpGTE(), ">=")
		gDecls.SetType($$, ScalarType(0))
	}
    ;

equality_expression:
      relational_expression
	{
		$$ = $1
	}
    | equality_expression EQ relational_expression
	{
		$$ = $1.Binary($3, OpEq(), "==")
		gDecls.SetType($$, ScalarType(0))
	}
    | equality_expression NEQ relational_expression
	{
		$$ = $1.Binary($3, OpNEq(), "!=")
		gDecls.SetType($$, ScalarType(0))
	}
    | equality_expression MATCH STRING
	{
		re, err := gDecls.RegExp($3)
		if err == nil {
			$$ = $1.Match($3, re)
			gDecls.SetType($$, ScalarType(0))
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
		$$ = $1.Binary($3, OpAnd(), "&&")
		gDecls.SetType($$, ScalarType(0))
	}
    | expression OR equality_expression
	{
		$$ = $1.Binary($3, OpOr(), "||")
		gDecls.SetType($$, ScalarType(0))
	}
    ;

%%

type ParseError struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Error  string `json:"error"`
}

func NewError(line, column int, msg string, args ...interface{}) *ParseError {
	return &ParseError{Line: line, Column: column, Error: fmt.Sprintf(msg, args...)}
}

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
			return PROD_SEQ
		} else if l.scan.Peek() == '~' {
			l.scan.Next()
			return PROD_PAR
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

func Compile(expr string, decls *Decls) (*Program, Type, *ParseError) {
	gMutex.Lock()
	defer gMutex.Unlock()

	gLID = 0
	gDecls = decls
	gLex = &lexer{}

	gError = nil
	gExpr = BadExpr

	reader := strings.NewReader(expr)
	gLex.scan.Init(reader)
	comp_Parse(gLex)

	var prog *Program
	if gError == nil {
		resType, errors := gDecls.Verify(gExpr.Id)
		if len(errors) > 0 {
			gError = NewError(0, 0, "%v", errors[0])
		} else {
			code := gExpr.Code()
			loops := make([]*iterator, gLID)
			prog = &Program{code, gDecls.values, gDecls.regexps, gDecls.funcs, loops}
		}

		return prog, resType, gError
	}

	return nil, nil, gError
}
