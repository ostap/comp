
//line src/comp/grammar.y:2
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

//line src/comp/grammar.y:106
type comp_SymType struct {
	yys int
	str  string
	num  float64
	expr Expr
	args []Expr
}

const TK_EQ = 57346
const TK_NEQ = 57347
const TK_MATCH = 57348
const TK_PROD = 57349
const TK_LTEQ = 57350
const TK_GTEQ = 57351
const TK_CAT = 57352
const TK_AND = 57353
const TK_OR = 57354
const NUMBER = 57355
const IDENT = 57356
const STRING = 57357

var comp_Toknames = []string{
	"TK_EQ",
	"TK_NEQ",
	"TK_MATCH",
	"TK_PROD",
	"TK_LTEQ",
	"TK_GTEQ",
	"TK_CAT",
	"TK_AND",
	"TK_OR",
	"NUMBER",
	"IDENT",
	"STRING",
}
var comp_Statenames = []string{}

const comp_EofCode = 1
const comp_ErrCode = 2
const comp_MaxDepth = 200

//line src/comp/grammar.y:394


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

//line yacctab:1
var comp_Exca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const comp_Nprod = 36
const comp_Private = 57344

var comp_TokenNames []string
var comp_States []string

const comp_Last = 81

var comp_Act = []int{

	4, 3, 9, 7, 28, 29, 17, 15, 16, 38,
	8, 33, 34, 18, 56, 11, 12, 13, 20, 39,
	60, 41, 2, 26, 27, 32, 5, 46, 21, 22,
	47, 48, 49, 50, 6, 65, 54, 55, 31, 30,
	57, 51, 52, 53, 17, 15, 16, 59, 42, 43,
	19, 18, 20, 11, 12, 13, 21, 22, 44, 45,
	17, 15, 16, 10, 64, 61, 58, 18, 62, 63,
	40, 21, 22, 1, 14, 35, 36, 37, 23, 24,
	25,
}
var comp_Pact = []int{

	6, -1000, 31, 33, 60, 74, -4, 15, -14, -1000,
	-1000, 47, 47, 47, -1000, -11, -1000, -1000, 31, 56,
	31, 31, 31, 31, 31, 12, 31, 31, 31, 31,
	31, 31, 31, 31, 31, -1000, -1000, -1000, -7, 45,
	40, 60, 74, 74, -4, -4, -1000, 15, 15, 15,
	15, -14, -14, -14, -1000, -1000, -1000, -1, -1000, 51,
	-1000, 50, -1000, 31, 17, -1000,
}
var comp_Pgo = []int{

	0, 74, 63, 2, 10, 3, 34, 26, 0, 1,
	73,
}
var comp_R1 = []int{

	0, 10, 10, 10, 1, 1, 1, 2, 2, 2,
	2, 9, 9, 3, 3, 3, 3, 4, 4, 4,
	5, 5, 5, 5, 6, 6, 6, 6, 6, 7,
	7, 7, 7, 8, 8, 8,
}
var comp_R2 = []int{

	0, 0, 7, 9, 1, 1, 3, 1, 1, 3,
	4, 1, 3, 1, 2, 2, 2, 1, 3, 3,
	1, 3, 3, 3, 1, 3, 3, 3, 3, 1,
	3, 3, 3, 1, 3, 3,
}
var comp_Chk = []int{

	-1000, -10, 16, -9, -8, -7, -6, -5, -4, -3,
	-2, 22, 23, 24, -1, 14, 15, 13, 20, 17,
	19, 11, 12, 4, 5, 6, 27, 28, 8, 9,
	24, 23, 10, 25, 26, -2, -2, -2, 20, -8,
	14, -8, -7, -7, -6, -6, 15, -5, -5, -5,
	-5, -4, -4, -4, -3, -3, 21, -9, 21, 7,
	21, 14, 18, 19, -8, 18,
}
var comp_Def = []int{

	1, -2, 0, 0, 11, 33, 29, 24, 20, 17,
	13, 0, 0, 0, 7, 8, 4, 5, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 14, 15, 16, 0, 0,
	0, 12, 34, 35, 30, 31, 32, 25, 26, 27,
	28, 21, 22, 23, 18, 19, 9, 0, 6, 0,
	10, 0, 2, 0, 0, 3,
}
var comp_Tok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 22, 3, 3, 3, 3, 3, 3,
	20, 21, 25, 24, 19, 23, 3, 26, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	27, 3, 28, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 16, 3, 18, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 17,
}
var comp_Tok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15,
}
var comp_Tok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var comp_Debug = 0

type comp_Lexer interface {
	Lex(lval *comp_SymType) int
	Error(s string)
}

const comp_Flag = -1000

func comp_Tokname(c int) string {
	if c > 0 && c <= len(comp_Toknames) {
		if comp_Toknames[c-1] != "" {
			return comp_Toknames[c-1]
		}
	}
	return fmt.Sprintf("tok-%v", c)
}

func comp_Statname(s int) string {
	if s >= 0 && s < len(comp_Statenames) {
		if comp_Statenames[s] != "" {
			return comp_Statenames[s]
		}
	}
	return fmt.Sprintf("state-%v", s)
}

func comp_lex1(lex comp_Lexer, lval *comp_SymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = comp_Tok1[0]
		goto out
	}
	if char < len(comp_Tok1) {
		c = comp_Tok1[char]
		goto out
	}
	if char >= comp_Private {
		if char < comp_Private+len(comp_Tok2) {
			c = comp_Tok2[char-comp_Private]
			goto out
		}
	}
	for i := 0; i < len(comp_Tok3); i += 2 {
		c = comp_Tok3[i+0]
		if c == char {
			c = comp_Tok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = comp_Tok2[1] /* unknown char */
	}
	if comp_Debug >= 3 {
		fmt.Printf("lex %U %s\n", uint(char), comp_Tokname(c))
	}
	return c
}

func comp_Parse(comp_lex comp_Lexer) int {
	var comp_n int
	var comp_lval comp_SymType
	var comp_VAL comp_SymType
	comp_S := make([]comp_SymType, comp_MaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	comp_state := 0
	comp_char := -1
	comp_p := -1
	goto comp_stack

ret0:
	return 0

ret1:
	return 1

comp_stack:
	/* put a state and value onto the stack */
	if comp_Debug >= 4 {
		fmt.Printf("char %v in %v\n", comp_Tokname(comp_char), comp_Statname(comp_state))
	}

	comp_p++
	if comp_p >= len(comp_S) {
		nyys := make([]comp_SymType, len(comp_S)*2)
		copy(nyys, comp_S)
		comp_S = nyys
	}
	comp_S[comp_p] = comp_VAL
	comp_S[comp_p].yys = comp_state

comp_newstate:
	comp_n = comp_Pact[comp_state]
	if comp_n <= comp_Flag {
		goto comp_default /* simple state */
	}
	if comp_char < 0 {
		comp_char = comp_lex1(comp_lex, &comp_lval)
	}
	comp_n += comp_char
	if comp_n < 0 || comp_n >= comp_Last {
		goto comp_default
	}
	comp_n = comp_Act[comp_n]
	if comp_Chk[comp_n] == comp_char { /* valid shift */
		comp_char = -1
		comp_VAL = comp_lval
		comp_state = comp_n
		if Errflag > 0 {
			Errflag--
		}
		goto comp_stack
	}

comp_default:
	/* default state action */
	comp_n = comp_Def[comp_state]
	if comp_n == -2 {
		if comp_char < 0 {
			comp_char = comp_lex1(comp_lex, &comp_lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if comp_Exca[xi+0] == -1 && comp_Exca[xi+1] == comp_state {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			comp_n = comp_Exca[xi+0]
			if comp_n < 0 || comp_n == comp_char {
				break
			}
		}
		comp_n = comp_Exca[xi+1]
		if comp_n < 0 {
			goto ret0
		}
	}
	if comp_n == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			comp_lex.Error("syntax error")
			Nerrs++
			if comp_Debug >= 1 {
				fmt.Printf("%s", comp_Statname(comp_state))
				fmt.Printf("saw %s\n", comp_Tokname(comp_char))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for comp_p >= 0 {
				comp_n = comp_Pact[comp_S[comp_p].yys] + comp_ErrCode
				if comp_n >= 0 && comp_n < comp_Last {
					comp_state = comp_Act[comp_n] /* simulate a shift of "error" */
					if comp_Chk[comp_state] == comp_ErrCode {
						goto comp_stack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if comp_Debug >= 2 {
					fmt.Printf("error recovery pops state %d\n", comp_S[comp_p].yys)
				}
				comp_p--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if comp_Debug >= 2 {
				fmt.Printf("error recovery discards %s\n", comp_Tokname(comp_char))
			}
			if comp_char == comp_EofCode {
				goto ret1
			}
			comp_char = -1
			goto comp_newstate /* try again in the same state */
		}
	}

	/* reduction by production comp_n */
	if comp_Debug >= 2 {
		fmt.Printf("reduce %v in:\n\t%v\n", comp_n, comp_Statname(comp_state))
	}

	comp_nt := comp_n
	comp_pt := comp_p
	_ = comp_pt // guard against "declared and not used"

	comp_p -= comp_R2[comp_n]
	comp_VAL = comp_S[comp_p+1]

	/* consult goto table to find next state */
	comp_n = comp_R1[comp_n]
	comp_g := comp_Pgo[comp_n]
	comp_j := comp_g + comp_S[comp_p].yys + 1

	if comp_j >= comp_Last {
		comp_state = comp_Act[comp_g]
	} else {
		comp_state = comp_Act[comp_j]
		if comp_Chk[comp_state] != -comp_n {
			comp_state = comp_Act[comp_g]
		}
	}
	// dummy call; replaced with literal code
	switch comp_nt {

	case 1:
		//line src/comp/grammar.y:139
		{ gComp = nil }
	case 2:
		//line src/comp/grammar.y:141
		{ gComp = Load(comp_S[comp_pt-1].str).Return(comp_S[comp_pt-5].args) }
	case 3:
		//line src/comp/grammar.y:143
		{ gComp = Load(comp_S[comp_pt-3].str).Select(comp_S[comp_pt-1].expr).Return(comp_S[comp_pt-7].args) }
	case 4:
		//line src/comp/grammar.y:148
		{
			val := StrVal(comp_S[comp_pt-0].str)
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return val
			}
		}
	case 5:
		//line src/comp/grammar.y:155
		{
			val := NumVal(comp_S[comp_pt-0].num)
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return val
			}
		}
	case 6:
		//line src/comp/grammar.y:162
		{ comp_VAL.expr = comp_S[comp_pt-1].expr }
	case 7:
		//line src/comp/grammar.y:167
		{ comp_VAL.expr = comp_S[comp_pt-0].expr }
	case 8:
		//line src/comp/grammar.y:169
		{
			attr := comp_S[comp_pt-0].str
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				idx, ok := h[attr]
				if ok {
					return StrVal(t[idx])
				}
	
				return BoolVal(false)
			}
		}
	case 9:
		//line src/comp/grammar.y:181
		{
			switch comp_S[comp_pt-2].str {
			default:
				parseError("x unknown function %v", comp_S[comp_pt-2].str)
			}
		}
	case 10:
		//line src/comp/grammar.y:188
		{
			switch comp_S[comp_pt-3].str {
			case "trunc":
				if len(comp_S[comp_pt-1].args) != 1 {
					parseError("trunc takes only 1 argument")
				}
	
				expr := comp_S[comp_pt-1].args[0]
				comp_VAL.expr = func(h Head, t Tuple) *Value {
					return NumVal(math.Trunc(expr(h, t).Num()))
				}
			case "dist":
				if len(comp_S[comp_pt-1].args) != 4 {
					parseError("dist takes only 4 arguments")
				}
	
				lat1expr := comp_S[comp_pt-1].args[0]
				lon1expr := comp_S[comp_pt-1].args[1]
				lat2expr := comp_S[comp_pt-1].args[2]
				lon2expr := comp_S[comp_pt-1].args[3]
				comp_VAL.expr = func(h Head, t Tuple) *Value {
					lat1 := lat1expr(h, t).Num()
					lon1 := lon1expr(h, t).Num()
					lat2 := lat2expr(h, t).Num()
					lon2 := lon2expr(h, t).Num()
	
					return NumVal(Dist(lat1, lon1, lat2, lon2))
				}
			default:
				parseError("unknown function %v", comp_S[comp_pt-3].str)
			}
		}
	case 11:
		//line src/comp/grammar.y:223
		{ comp_VAL.args = []Expr{comp_S[comp_pt-0].expr} }
	case 12:
		//line src/comp/grammar.y:224
		{ comp_VAL.args = append(comp_S[comp_pt-2].args, comp_S[comp_pt-0].expr) }
	case 13:
		//line src/comp/grammar.y:229
		{ comp_VAL.expr = comp_S[comp_pt-0].expr }
	case 14:
		//line src/comp/grammar.y:231
		{
			expr := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(!expr(h, t).Bool())
			}
		}
	case 15:
		//line src/comp/grammar.y:238
		{
			expr := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return NumVal(-expr(h, t).Num())
			}
		}
	case 16:
		//line src/comp/grammar.y:245
		{
			expr := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return NumVal(+expr(h, t).Num())
			}
		}
	case 17:
		//line src/comp/grammar.y:255
		{ comp_VAL.expr = comp_S[comp_pt-0].expr }
	case 18:
		//line src/comp/grammar.y:257
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return NumVal(l(h, t).Num() * r(h, t).Num())
			}
		}
	case 19:
		//line src/comp/grammar.y:265
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return NumVal(l(h, t).Num() / r(h, t).Num())
			}
		}
	case 20:
		//line src/comp/grammar.y:276
		{ comp_VAL.expr = comp_S[comp_pt-0].expr }
	case 21:
		//line src/comp/grammar.y:278
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return NumVal(l(h, t).Num() + r(h, t).Num())
			}
		}
	case 22:
		//line src/comp/grammar.y:286
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return NumVal(l(h, t).Num() - r(h, t).Num())
			}
		}
	case 23:
		//line src/comp/grammar.y:294
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return StrVal(l(h, t).Str() + r(h, t).Str())
			}
		}
	case 24:
		//line src/comp/grammar.y:305
		{ comp_VAL.expr = comp_S[comp_pt-0].expr }
	case 25:
		//line src/comp/grammar.y:307
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(l(h, t).Num() < r(h, t).Num())
			}
		}
	case 26:
		//line src/comp/grammar.y:315
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(l(h, t).Num() > r(h, t).Num())
			}
		}
	case 27:
		//line src/comp/grammar.y:323
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(l(h, t).Num() <= r(h, t).Num())
			}
		}
	case 28:
		//line src/comp/grammar.y:331
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(l(h, t).Num() >= r(h, t).Num())
			}
		}
	case 29:
		//line src/comp/grammar.y:342
		{ comp_VAL.expr = comp_S[comp_pt-0].expr }
	case 30:
		//line src/comp/grammar.y:344
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(l(h, t).Eq(r(h, t)))
			}
		}
	case 31:
		//line src/comp/grammar.y:352
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(!l(h, t).Eq(r(h, t)))
			}
		}
	case 32:
		//line src/comp/grammar.y:360
		{
			re, err := regexp.Compile(comp_S[comp_pt-0].str)
			if err != nil {
				parseError("%v", err)
			}
	
			expr := comp_S[comp_pt-2].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(re.MatchString(expr(h, t).Str()))
			}
		}
	case 33:
		//line src/comp/grammar.y:375
		{ comp_VAL.expr = comp_S[comp_pt-0].expr }
	case 34:
		//line src/comp/grammar.y:377
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(l(h, t).Bool() && r(h, t).Bool())
			}
		}
	case 35:
		//line src/comp/grammar.y:385
		{
			l := comp_S[comp_pt-2].expr
			r := comp_S[comp_pt-0].expr
			comp_VAL.expr = func(h Head, t Tuple) *Value {
				return BoolVal(l(h, t).Bool() || r(h, t).Bool())
			}
		}
	}
	goto comp_stack /* stack new state and value */
}
