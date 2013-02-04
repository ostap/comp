%{
package main

import (
	"regexp"
	"text/scanner"
	"strings"
	"strconv"
	"fmt"
	"math"
)

type Expr func(t Tuple) *Value

var gMem   *Mem
var gViews Views
var gComp  Body
var gError error
var gLex   *lexer
%}

%union {
	str  string
	num  float64
	body Body
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
%type <body> generator
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
	{ gComp = $4.Return($2) }
    | '[' expression_list '|' generator ',' expression ']'
	{ gComp = $4.Select($6).Return($2) }
    ;

generator:
      IDENT TK_PROD IDENT
	{
		if !gViews.Has($3) {
			parseError("unknown dataset %v", $3)
		} else {
			head, body := gViews.Load($3)
			gMem.Decl($1, head)
			$$ = body
		}
	}
    ;

primary_expression:
      STRING
	{
		val := StrVal($1)
		$$ = func(t Tuple) *Value {
			return val
		}
	}
    | NUMBER
	{
		val := NumVal($1)
		$$ = func(t Tuple) *Value {
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
		pos := gMem.PosPtr($1)
		$$ = func(t Tuple) *Value {
			return StrVal(t[*pos])
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

			expr := $3[0]
			$$ = func(t Tuple) *Value {
				return NumVal(math.Trunc(expr(t).Num()))
			}
		case "dist":
			if len($3) != 4 {
				parseError("dist takes only 4 arguments")
			}

			lat1expr := $3[0]
			lon1expr := $3[1]
			lat2expr := $3[2]
			lon2expr := $3[3]
			$$ = func(t Tuple) *Value {
				lat1 := lat1expr(t).Num()
				lon1 := lon1expr(t).Num()
				lat2 := lat2expr(t).Num()
				lon2 := lon2expr(t).Num()

				return NumVal(Dist(lat1, lon1, lat2, lon2))
			}
		case "trim":
			if len($3) != 1 {
				parseError("trim takes only 1 argument")
			}

			expr := $3[0]
			$$ = func(t Tuple) *Value {
				return StrVal(strings.Trim(expr(t).Str(), " \t\n\r"))
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
		expr := $2
		$$ = func(t Tuple) *Value {
			return BoolVal(!expr(t).Bool())
		}
	}
    | '-' postfix_expression
	{
		expr := $2
		$$ = func(t Tuple) *Value {
			return NumVal(-expr(t).Num())
		}
	}
    | '+' postfix_expression
	{
		expr := $2
		$$ = func(t Tuple) *Value {
			return NumVal(+expr(t).Num())
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
		$$ = func(t Tuple) *Value {
			return NumVal(l(t).Num() * r(t).Num())
		}
	}
    | multiplicative_expression '/' unary_expression
	{
		l := $1
		r := $3
		$$ = func(t Tuple) *Value {
			return NumVal(l(t).Num() / r(t).Num())
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
		$$ = func(t Tuple) *Value {
			return NumVal(l(t).Num() + r(t).Num())
		}
	}
    | additive_expression '-' multiplicative_expression
	{
		l := $1
		r := $3
		$$ = func(t Tuple) *Value {
			return NumVal(l(t).Num() - r(t).Num())
		}
	}
    | additive_expression TK_CAT multiplicative_expression
	{
		l := $1
		r := $3
		$$ = func(t Tuple) *Value {
			return StrVal(l(t).Str() + r(t).Str())
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
		$$ = func(t Tuple) *Value {
			return BoolVal(l(t).Num() < r(t).Num())
		}
	}
    | relational_expression '>' additive_expression
	{
		l := $1
		r := $3
		$$ = func(t Tuple) *Value {
			return BoolVal(l(t).Num() > r(t).Num())
		}
	}
    | relational_expression TK_LTEQ additive_expression
	{
		l := $1
		r := $3
		$$ = func(t Tuple) *Value {
			return BoolVal(l(t).Num() <= r(t).Num())
		}
	}
    | relational_expression TK_GTEQ additive_expression
	{
		l := $1
		r := $3
		$$ = func(t Tuple) *Value {
			return BoolVal(l(t).Num() >= r(t).Num())
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
		$$ = func(t Tuple) *Value {
			return BoolVal(l(t).Eq(r(t)))
		}
	}
    | equality_expression TK_NEQ relational_expression
	{
		l := $1
		r := $3
		$$ = func(t Tuple) *Value {
			return BoolVal(!l(t).Eq(r(t)))
		}
	}
    | equality_expression TK_MATCH STRING
	{
		re, err := regexp.Compile($3)
		if err != nil {
			parseError("%v", err)
		}

		expr := $1
		$$ = func(t Tuple) *Value {
			return BoolVal(re.MatchString(expr(t).Str()))
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
		$$ = func(t Tuple) *Value {
			return BoolVal(l(t).Bool() && r(t).Bool())
		}
	}
    | expression TK_OR equality_expression
	{
		l := $1
		r := $3
		$$ = func(t Tuple) *Value {
			return BoolVal(l(t).Bool() || r(t).Bool())
		}
	}
    ;

%%

func parseError(s string, v ...interface{}) {
	gError = fmt.Errorf("%+v - %v", gLex.scan.Pos(), fmt.Sprintf(s, v...))
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

func Parse(query string, views Views) (Body, error) {
	gError = nil
	gViews = views
	gMem = NewMem()
	gLex = &lexer{}

	reader := strings.NewReader(query)
	gLex.scan.Init(reader)
	comp_Parse(gLex)

	if gError == nil {
		bad := gMem.BadAttrs()
		if len(bad) > 0 {
			gError = fmt.Errorf("unknown identifier(s): %v", bad)
		}
	}

	return gComp, gError
}
