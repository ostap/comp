%{
package main

import (
	"text/scanner"
	"strings"
	"strconv"
	"fmt"
	"math"
)

type ParseError struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Error  string `json:"error"`
}

func NewError(line, column int, msg string, args ...interface{}) *ParseError {
	return &ParseError{Line: line, Column: column, Error: fmt.Sprintf(msg, args...)}
}

var gStore Store
var gLex   *lexer

var gMem   *Mem
var gLoad  string
var gComp  Comp
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
	{ gComp = nil }
    | '[' expression_list '|' generator ']'
	{
		gLoad = $4
		gComp = Return(Reflect, $2)
	}
    | '[' expression_list '|' generator ',' expression ']'
	{
		gLoad = $4
		gComp = Return(Select(Reflect, $6), $2)
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
		val := Value($1)
		$$ = func(m *Mem, t Tuple) Value {
			return val
		}
	}
    | NUMBER
	{
		val := Value($1)
		$$ = func(m *Mem, t Tuple) Value {
			return val
		}
	}
    | '(' expression ')'
	{ $$ = $2 }
    ;

identifier:
      IDENT			{ $$ = $1 }
    | identifier '.' IDENT	{ $$ = fmt.Sprintf("%v.%v", $1, $3) }
    ;

postfix_expression:
      primary_expression
	{ $$ = $1 }
    | identifier
	{
		pos := gMem.AttrPos($1)
		$$ = func(m *Mem, t Tuple) Value {
			return t[m.Attrs[pos]]
		}
	}
    | identifier '(' ')'
	{
		switch $1 {
		default:
			parseError("x unknown function %v", $1)
		}
	}
    | identifier '(' expression_list ')'
	{
		switch $1 {
		case "trunc":
			if len($3) != 1 {
				parseError("trunc takes only 1 argument")
			}

			e := $3[0]
			$$ = func(m *Mem, t Tuple) Value {
				return math.Trunc(Num(e(m, t)))
			}
		case "dist":
			if len($3) != 4 {
				parseError("dist takes only 4 arguments")
			}

			lat1expr := $3[0]
			lon1expr := $3[1]
			lat2expr := $3[2]
			lon2expr := $3[3]
			$$ = func(m *Mem, t Tuple) Value {
				lat1 := Num(lat1expr(m, t))
				lon1 := Num(lon1expr(m, t))
				lat2 := Num(lat2expr(m, t))
				lon2 := Num(lon2expr(m, t))

				return Dist(lat1, lon1, lat2, lon2)
			}
		case "trim":
			if len($3) != 1 {
				parseError("trim takes only 1 argument")
			}

			e := $3[0]
			$$ = func(m *Mem, t Tuple) Value {
				return strings.Trim(Str(e(m, t)), " \t\n\r")
			}
		case "lower":
			if len($3) != 1 {
				parseError("lower takes only 1 argument")
			}

			e := $3[0]
			$$ = func(m *Mem, t Tuple) Value {
				return strings.ToLower(Str(e(m, t)))
			}
		case "upper":
			if len($3) != 1 {
				parseError("upper takes only 1 argument")
			}

			e := $3[0]
			$$ = func(m *Mem, t Tuple) Value {
				return strings.ToUpper(Str(e(m, t)))
			}
		case "fuzzy":
			if len($3) != 2 {
				parseError("fuzzy takes only 2 arguments")
			}

			se := $3[0]
			te := $3[1]
			$$ = func(m *Mem, t Tuple) Value {
				return Fuzzy(Str(se(m, t)), Str(te(m, t)))
			}
		default:
			parseError("unknown function %v", $1)
		}
	}
    ;

expression_list:
      expression			{ $$ = []Expr{$1} }
    | expression_list ',' expression	{ $$ = append($1, $3) }
    ;

unary_expression:
      postfix_expression
	{ $$ = $1 }
    | '!' postfix_expression
	{
		e := $2
		$$ = func(m *Mem, t Tuple) Value {
			return !Bool(e(m, t))
		}
	}
    | '-' postfix_expression
	{
		e := $2
		$$ = func(m *Mem, t Tuple) Value {
			return -Num(e(m, t))
		}
	}
    | '+' postfix_expression
	{
		e := $2
		$$ = func(m *Mem, t Tuple) Value {
			return +Num(e(m, t))
		}
	}
    ;

multiplicative_expression:
      unary_expression
	{ $$ = $1 }
    | multiplicative_expression '*' unary_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Num(l(m, t)) * Num(r(m, t))
		}
	}
    | multiplicative_expression '/' unary_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Num(l(m, t)) / Num(r(m, t))
		}
	}
    ;

additive_expression:
      multiplicative_expression
	{ $$ = $1 }
    | additive_expression '+' multiplicative_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Num(l(m, t)) + Num(r(m, t))
		}
	}
    | additive_expression '-' multiplicative_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Num(l(m, t)) - Num(r(m, t))
		}
	}
    | additive_expression TK_CAT multiplicative_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Str(l(m, t)) + Str(r(m, t))
		}
	}
    ;

relational_expression:
      additive_expression
	{ $$ = $1 }
    | relational_expression '<' additive_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Num(l(m, t)) < Num(r(m, t))
		}
	}
    | relational_expression '>' additive_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Num(l(m, t)) > Num(r(m, t))
		}
	}
    | relational_expression TK_LTEQ additive_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Num(l(m, t)) <= Num(r(m, t))
		}
	}
    | relational_expression TK_GTEQ additive_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Num(l(m, t)) >= Num(r(m, t))
		}
	}
    ;

equality_expression:
      relational_expression
	{ $$ = $1 }
    | equality_expression TK_EQ relational_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Eq(l(m, t), r(m, t))
		}
	}
    | equality_expression TK_NEQ relational_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return !Eq(l(m, t), r(m, t))
		}
	}
    | equality_expression TK_MATCH STRING
	{
		re, err := gMem.RegExp($3)
		if err != nil {
			parseError("%v", err)
		}

		e := $1
		$$ = func(m *Mem, t Tuple) Value {
			return m.MatchString(re, Str(e(m, t)))
		}
	}
    ;

expression:
      equality_expression
	{ $$ = $1 }
    | expression TK_AND equality_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Bool(l(m, t)) && Bool(r(m, t))
		}
	}
    | expression TK_OR equality_expression
	{
		l := $1
		r := $3
		$$ = func(m *Mem, t Tuple) Value {
			return Bool(l(m, t)) || Bool(r(m, t))
		}
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

func Parse(query string, store Store) (*Mem, string, Comp, *ParseError) {
	gStore = store
	gLex = &lexer{}

	gMem = NewMem()
	gLoad = ""
	gComp = nil
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

	return gMem, gLoad, gComp, gError
}
