package main

import (
	"bytes"
	"fmt"
	"strconv"
	"sync/atomic"
)

type Expr struct {
	Id   int64
	Name string
	Code func() []Op
}

var BadExpr = Expr{0, "", nil}
var exprSeqNum int64 = 1

func nextEID() int64 {
	return atomic.AddInt64(&exprSeqNum, 1)
}

func ExprLoad(name string, addr int) Expr {
	return Expr{nextEID(), name, func() []Op {
		return []Op{OpLoad(addr)}
	}}
}

func ExprObject(fields []Expr) Expr {
	name := new(bytes.Buffer)
	fmt.Fprintf(name, "{")
	for i, f := range fields {
		if i != 0 {
			fmt.Fprintf(name, ", ")
		}
		fmt.Fprintf(name, f.Name)
	}
	fmt.Fprintf(name, "}")

	return Expr{nextEID(), name.String(), func() []Op {
		code := []Op{OpObject(len(fields))}
		for i, f := range fields {
			for _, c := range f.Code() {
				code = append(code, c)
			}
			code = append(code, OpSet(i))
		}

		return code
	}}
}

func ExprList(elems []Expr) Expr {
	// TODO: compose a name
	return Expr{nextEID(), "", func() []Op {
		code := []Op{OpList}
		for _, e := range elems {
			for _, c := range e.Code() {
				code = append(code, c)
			}
			code = append(code, OpAppend)
		}

		return code
	}}
}

func ExprCompSelect(listAddr int, list Expr, elemAddr int, sel, ret Expr) Expr {
	// TODO: compose a name
	return Expr{nextEID(), "", func() []Op {
		code := list.Code()
		selCode := sel.Code()
		retCode := ret.Code()

		loopJump := 1 + len(selCode) + 2 + len(retCode) + 4
		nextJump := -2 - len(retCode) - 2 - len(selCode) - 1
		testJump := 1 + len(retCode) + 3

		code = append(code, OpLoop(loopJump))
		code = append(code, OpStore(elemAddr))
		for _, c := range selCode {
			code = append(code, c)
		}
		code = append(code, OpTest(testJump))
		code = append(code, OpLoad(listAddr))
		for _, c := range retCode {
			code = append(code, c)
		}
		code = append(code, OpAppend)
		code = append(code, OpStore(listAddr))
		code = append(code, OpNext(nextJump))
		return append(code, OpLoad(listAddr))
	}}
}

func ExprComp(listAddr int, list Expr, elemAddr int, ret Expr) Expr {
	// TODO: compose a name
	return Expr{nextEID(), "", func() []Op {
		code := list.Code()
		retCode := ret.Code()

		loopJump := 2 + len(retCode) + 4
		nextJump := -2 - len(retCode) - 2

		code = append(code, OpLoop(loopJump))
		code = append(code, OpStore(elemAddr))
		code = append(code, OpLoad(listAddr))
		for _, c := range retCode {
			code = append(code, c)
		}
		code = append(code, OpAppend)
		code = append(code, OpStore(listAddr))
		code = append(code, OpNext(nextJump))
		return append(code, OpLoad(listAddr))
	}}
}

func (e Expr) Field(name string, pos *int) Expr {
	return Expr{nextEID(), fmt.Sprintf("%v.%v", e.Name, name), func() []Op {
		return append(e.Code(), OpGet(*pos))
	}}
}

func (l Expr) Binary(r Expr, op Op, name string) Expr {
	return Expr{nextEID(), fmt.Sprintf("%v %v %v", l.Name, name, r.Name), func() []Op {
		lc := l.Code()
		rc := r.Code()
		code := make([]Op, len(rc)+len(lc)+1)
		copy(code, rc)
		copy(code[len(rc):], lc)
		code[len(code)-1] = op

		return code
	}}
}

func (e Expr) Unary(op Op, name string) Expr {
	return Expr{nextEID(), fmt.Sprintf("%v%v", name, e.Name), func() []Op {
		return append(e.Code(), op)
	}}
}

func (e Expr) Match(pattern string, re int) Expr {
	name := fmt.Sprintf("%v =~ %v", e.Name, strconv.Quote(pattern))
	return Expr{nextEID(), name, func() []Op {
		return append(e.Code(), OpMatch(re))
	}}
}

func (e Expr) Call(args []Expr) (expr Expr, err error) {
	// TODO: compose a name
	expr = BadExpr
	err = nil

	/* TODO: enable functions
	switch e.Name {
	case "trunc":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{nextEID(), "", func() []Op {
				return Number(math.Trunc(ToNum(e, m)))
			}}
		} else {
			err = fmt.Errorf("trunc takes only 1 argument")
		}
	case "dist":
		if len(args) == 4 {
			lat1expr := args[0]
			lon1expr := args[1]
			lat2expr := args[2]
			lon2expr := args[3]

			expr = Expr{nextEID(), "", func() []Op {
				lat1 := ToNum(lat1expr, m)
				lon1 := ToNum(lon1expr, m)
				lat2 := ToNum(lat2expr, m)
				lon2 := ToNum(lon2expr, m)

				return Number(Dist(lat1, lon1, lat2, lon2))
			}}
		} else {
			err = fmt.Errorf("dist takes only 4 arguments")
		}
	case "trim":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{nextEID(), "", func() []Op {
				return String(strings.Trim(ToStr(e, m), " \t\n\r"))
			}}
		} else {
			err = fmt.Errorf("trim takes only 1 argument")
		}
	case "lower":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{nextEID(), "", func() []Op {
				return String(strings.ToLower(ToStr(e, m)))
			}}
		} else {
			err = fmt.Errorf("lower takes only 1 argument")
		}
	case "upper":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{nextEID(), "", func() []Op {
				return String(strings.ToUpper(ToStr(e, m)))
			}}
		} else {
			err = fmt.Errorf("upper takes only 1 argument")
		}
	case "fuzzy":
		if len(args) == 2 {
			se := args[0]
			te := args[1]
			expr = Expr{nextEID(), "", func() []Op {
				return Number(Fuzzy(ToStr(se, m), ToStr(te, m)))
			}}
		} else {
			err = fmt.Errorf("fuzzy takes only 2 arguments")
		}
	default:
		err = fmt.Errorf("unknown function %v(%d)", e.Name, len(args))
	}
	*/

	return
}
