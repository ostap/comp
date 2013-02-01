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

const (
	Bool   = iota
	Number = iota
	String = iota
)

type Value struct {
	strval  string
	numval  float64
	boolval bool
	kind    int  // Bool, Number, String
}

func BoolVal(b bool) *Value {
	return &Value{strval: "", numval: math.NaN(), boolval: b, kind: Bool}
}

func StrVal(s string) *Value {
	return &Value{strval: s, numval: math.NaN(), boolval: false, kind: String}
}

func NumVal(n float64) *Value {
	return &Value{strval: "", numval: n, boolval: false, kind: Number}
}

func (v *Value) Bool() bool {
	switch v.kind {
	case Bool:
		return v.boolval
	case Number:
		if math.IsNaN(v.numval) {
			return false
		}

		return v.numval != 0
	case String:
		return v.strval != ""
	}

	return false
}

func (v *Value) Num() float64 {
	switch v.kind {
	case Number:
		return v.numval
	case String:
		res, _ := strconv.ParseFloat(v.strval, 64)
		return res
	case Bool:
		res := 0.0
		if v.boolval {
			res = 1.0
		}

		return res
	}

	return math.NaN()
}

func (v *Value) Str() string {
	switch v.kind {
	case String:
		return v.strval
	case Number:
		return fmt.Sprintf("%v", v.numval)
	case Bool:
		return fmt.Sprintf("%v", v.boolval)
	}

	return ""
}

// TODO: check reflexivity, symmetry, transitivity
func (v *Value) Eq(arg *Value) bool {
	if v.kind == Number || arg.kind == Number {
		return v.Num() == arg.Num()
	} else if v.kind == String || arg.kind == String {
		return v.Str() == arg.Str()
	} else if v.kind == Bool || arg.kind == Bool {
		return v.Bool() == arg.Bool()
	}

	return false
}

type Expr func(h Head, t Tuple) *Value

var gComp Body
var gError bool
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
    | '[' expression_list '|' IDENT TK_PROD IDENT ']'
	{ gComp = Load($6).Return($2) }
    | '[' expression_list '|' IDENT TK_PROD IDENT ',' expression ']'
	{ gComp = Load($6).Select($8).Return($2) }
    ;

primary_expression:
      STRING
	{
		val := StrVal($1)
		$$ = func(h Head, t Tuple) *Value {
			return val
		}
	}
    | NUMBER
	{
		val := NumVal($1)
		$$ = func(h Head, t Tuple) *Value {
			return val
		}
	}
    | '(' expression ')'
	{ $$ = $2 }
    ;

postfix_expression:
      primary_expression
	{ $$ = $1 }
    | IDENT
	{
		attr := $1
		$$ = func(h Head, t Tuple) *Value {
			idx, ok := h[attr]
			if ok {
				return StrVal(t[idx])
			}

			return BoolVal(false)
		}
	}
    | IDENT '(' ')'
	{
		switch $1 {
		default:
			parseError("x unknown function %v", $1)
		}
	}
    | IDENT '(' expression_list ')'
	{
		switch $1 {
		case "trunc":
			if len($3) != 1 {
				parseError("trunc takes only 1 argument")
			}

			expr := $3[0]
			$$ = func(h Head, t Tuple) *Value {
				return NumVal(math.Trunc(expr(h, t).Num()))
			}
		case "dist":
			if len($3) != 4 {
				parseError("dist takes only 4 arguments")
			}

			lat1expr := $3[0]
			lon1expr := $3[1]
			lat2expr := $3[2]
			lon2expr := $3[3]
			$$ = func(h Head, t Tuple) *Value {
				lat1 := lat1expr(h, t).Num()
				lon1 := lon1expr(h, t).Num()
				lat2 := lat2expr(h, t).Num()
				lon2 := lon2expr(h, t).Num()

				return NumVal(Dist(lat1, lon1, lat2, lon2))
			}
		case "trim":
			if len($3) != 1 {
				parseError("trim takes only 1 argument")
			}

			expr := $3[0]
			$$ = func(h Head, t Tuple) *Value {
				return StrVal(strings.Trim(expr(h, t).Str(), " \t\n\r"))
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
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(!expr(h, t).Bool())
		}
	}
    | '-' postfix_expression
	{
		expr := $2
		$$ = func(h Head, t Tuple) *Value {
			return NumVal(-expr(h, t).Num())
		}
	}
    | '+' postfix_expression
	{
		expr := $2
		$$ = func(h Head, t Tuple) *Value {
			return NumVal(+expr(h, t).Num())
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
		$$ = func(h Head, t Tuple) *Value {
			return NumVal(l(h, t).Num() * r(h, t).Num())
		}
	}
    | multiplicative_expression '/' unary_expression
	{
		l := $1
		r := $3
		$$ = func(h Head, t Tuple) *Value {
			return NumVal(l(h, t).Num() / r(h, t).Num())
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
		$$ = func(h Head, t Tuple) *Value {
			return NumVal(l(h, t).Num() + r(h, t).Num())
		}
	}
    | additive_expression '-' multiplicative_expression
	{
		l := $1
		r := $3
		$$ = func(h Head, t Tuple) *Value {
			return NumVal(l(h, t).Num() - r(h, t).Num())
		}
	}
    | additive_expression TK_CAT multiplicative_expression
	{
		l := $1
		r := $3
		$$ = func(h Head, t Tuple) *Value {
			return StrVal(l(h, t).Str() + r(h, t).Str())
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
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(l(h, t).Num() < r(h, t).Num())
		}
	}
    | relational_expression '>' additive_expression
	{
		l := $1
		r := $3
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(l(h, t).Num() > r(h, t).Num())
		}
	}
    | relational_expression TK_LTEQ additive_expression
	{
		l := $1
		r := $3
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(l(h, t).Num() <= r(h, t).Num())
		}
	}
    | relational_expression TK_GTEQ additive_expression
	{
		l := $1
		r := $3
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(l(h, t).Num() >= r(h, t).Num())
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
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(l(h, t).Eq(r(h, t)))
		}
	}
    | equality_expression TK_NEQ relational_expression
	{
		l := $1
		r := $3
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(!l(h, t).Eq(r(h, t)))
		}
	}
    | equality_expression TK_MATCH STRING
	{
		re, err := regexp.Compile($3)
		if err != nil {
			parseError("%v", err)
		}

		expr := $1
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(re.MatchString(expr(h, t).Str()))
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
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(l(h, t).Bool() && r(h, t).Bool())
		}
	}
    | expression TK_OR equality_expression
	{
		l := $1
		r := $3
		$$ = func(h Head, t Tuple) *Value {
			return BoolVal(l(h, t).Bool() || r(h, t).Bool())
		}
	}
    ;

%%

func parseError(s string, v ...interface{}) {
	fmt.Printf(s, v...)
	fmt.Printf("\n")
	gError = true
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
	fmt.Printf("%+v - %v\n", l.scan.Pos(), s)
}

func Parse(query string) Body {
	gError = false

	lex := &lexer{}
	reader := strings.NewReader(query)
	lex.scan.Init(reader)
	comp_Parse(lex)

	if gError {
		return nil
	}

	return gComp
}
