package main

import (
	"fmt"
	"math"
	"strings"
	"sync/atomic"
)

type Expr struct {
	Id   int64
	Name string
	Eval func(mem *Mem) Value
}

var BadExpr = Expr{0, "", nil}
var exprSeqNum int64 = 1

func nextEID() int64 {
	return atomic.AddInt64(&exprSeqNum, 1)
}

func ToBool(e Expr, m *Mem) bool {
	return bool(e.Eval(m).Bool())
}

func ToNum(e Expr, m *Mem) float64 {
	return float64(e.Eval(m).Number())
}

func ToStr(e Expr, m *Mem) string {
	return string(e.Eval(m).String())
}

func ToList(e Expr, m *Mem) List {
	return e.Eval(m).List()
}

func ToObject(e Expr, m *Mem) Object {
	return e.Eval(m).Object()
}

func ExprConst(c Value) Expr {
	return Expr{nextEID(), "", func(mem *Mem) Value {
		return c
	}}
}

func ExprLoad(name string) Expr {
	id := nextEID()
	return Expr{id, name, func(mem *Mem) Value {
		return mem.Load(id, name)
	}}
}

func ExprList(elems []Expr) Expr {
	return Expr{nextEID(), "", func(mem *Mem) Value {
		res := make(List, len(elems))
		for i, e := range elems {
			res[i] = e.Eval(mem)
		}

		return res
	}}
}

func ExprComp(loop *Loop) Expr {
	return Expr{nextEID(), "", func(mem *Mem) Value {
		return loop.Eval(mem)
	}}
}

func ExprObject(fields []Expr) Expr {
	id := nextEID()
	head := make(Head)
	for i, f := range fields {
		head[f.Name] = i
	}

	return Expr{id, "", func(mem *Mem) Value {
		obj := make(Object, len(fields))
		mem.SetHead(id, head)

		for i, f := range fields {
			obj[i] = f.Eval(mem)
		}

		return obj
	}}
}

func (e Expr) Field(name string) Expr {
	return Expr{nextEID(), name, func(mem *Mem) Value {
		val := ToObject(e, mem)
		pos := mem.Field(e.Id, name)
		if pos > -1 {
			return val[pos]
		}

		return Bool(false)
	}}
}

func (e Expr) Call(args []Expr) (expr Expr, err error) {
	expr = BadExpr
	err = nil

	switch e.Name {
	case "trunc":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{nextEID(), "", func(m *Mem) Value {
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

			expr = Expr{nextEID(), "", func(m *Mem) Value {
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
			expr = Expr{nextEID(), "", func(m *Mem) Value {
				return String(strings.Trim(ToStr(e, m), " \t\n\r"))
			}}
		} else {
			err = fmt.Errorf("trim takes only 1 argument")
		}
	case "lower":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{nextEID(), "", func(m *Mem) Value {
				return String(strings.ToLower(ToStr(e, m)))
			}}
		} else {
			err = fmt.Errorf("lower takes only 1 argument")
		}
	case "upper":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{nextEID(), "", func(m *Mem) Value {
				return String(strings.ToUpper(ToStr(e, m)))
			}}
		} else {
			err = fmt.Errorf("upper takes only 1 argument")
		}
	case "fuzzy":
		if len(args) == 2 {
			se := args[0]
			te := args[1]
			expr = Expr{nextEID(), "", func(m *Mem) Value {
				return Number(Fuzzy(ToStr(se, m), ToStr(te, m)))
			}}
		} else {
			err = fmt.Errorf("fuzzy takes only 2 arguments")
		}
	default:
		err = fmt.Errorf("unknown function %v(%d)", e.Name, len(args))
	}

	return
}

func (e Expr) Not() Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(!ToBool(e, m))
	}}
}

func (e Expr) Neg() Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Number(-ToNum(e, m))
	}}
}

func (e Expr) Pos() Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Number(+ToNum(e, m))
	}}
}

func (l Expr) Mul(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Number(ToNum(l, m) * ToNum(r, m))
	}}
}

func (l Expr) Div(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Number(ToNum(l, m) / ToNum(r, m))
	}}
}

func (l Expr) Add(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Number(ToNum(l, m) + ToNum(r, m))
	}}
}

func (l Expr) Sub(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Number(ToNum(l, m) - ToNum(r, m))
	}}
}

func (l Expr) Cat(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return String(ToStr(l, m) + ToStr(r, m))
	}}
}

func (l Expr) LT(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(ToNum(l, m) < ToNum(r, m))
	}}
}

func (l Expr) GT(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(ToNum(l, m) > ToNum(r, m))
	}}
}

func (l Expr) LTE(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(ToNum(l, m) <= ToNum(r, m))
	}}
}

func (l Expr) GTE(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(ToNum(l, m) >= ToNum(r, m))
	}}
}

func (l Expr) Eq(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return l.Eval(m).Equals(r.Eval(m))
	}}
}

func (l Expr) NotEq(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(!l.Eval(m).Equals(r.Eval(m)))
	}}
}

func (e Expr) Match(re int) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(m.MatchString(re, ToStr(e, m)))
	}}
}

func (l Expr) And(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(ToBool(l, m) && ToBool(r, m))
	}}
}

func (l Expr) Or(r Expr) Expr {
	return Expr{nextEID(), "", func(m *Mem) Value {
		return Bool(ToBool(l, m) || ToBool(r, m))
	}}
}
