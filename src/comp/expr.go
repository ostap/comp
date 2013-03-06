package main

import (
	"fmt"
	"math"
	"strings"
)

type Expr struct {
	Name string
	Eval func(m *Mem, t Tuple) Value
}

var BadExpr = Expr{"", nil}

func ExprValue(value Value) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return value
	}}
}

func ExprAttr(name string, pos int) Expr {
	return Expr{name, func(m *Mem, t Tuple) Value {
		if idx := m.Attrs[pos]; idx > -1 {
			return t[m.Attrs[pos]]
		}

		return t
	}}
}

func ExprFunc(name string, args []Expr) (expr Expr, err error) {
	expr = BadExpr
	err = nil

	switch name {
	case "trunc":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{"", func(m *Mem, t Tuple) Value {
				return math.Trunc(NumEval(e, m, t))
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

			expr = Expr{"", func(m *Mem, t Tuple) Value {
				lat1 := NumEval(lat1expr, m, t)
				lon1 := NumEval(lon1expr, m, t)
				lat2 := NumEval(lat2expr, m, t)
				lon2 := NumEval(lon2expr, m, t)

				return Dist(lat1, lon1, lat2, lon2)
			}}
		} else {
			err = fmt.Errorf("dist takes only 4 arguments")
		}
	case "trim":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{"", func(m *Mem, t Tuple) Value {
				return strings.Trim(StrEval(e, m, t), " \t\n\r")
			}}
		} else {
			err = fmt.Errorf("trim takes only 1 argument")
		}
	case "lower":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{"", func(m *Mem, t Tuple) Value {
				return strings.ToLower(StrEval(e, m, t))
			}}
		} else {
			err = fmt.Errorf("lower takes only 1 argument")
		}
	case "upper":
		if len(args) == 1 {
			e := args[0]
			expr = Expr{"", func(m *Mem, t Tuple) Value {
				return strings.ToUpper(StrEval(e, m, t))
			}}
		} else {
			err = fmt.Errorf("upper takes only 1 argument")
		}
	case "fuzzy":
		if len(args) == 2 {
			se := args[0]
			te := args[1]
			expr = Expr{"", func(m *Mem, t Tuple) Value {
				return Fuzzy(StrEval(se, m, t), StrEval(te, m, t))
			}}
		} else {
			err = fmt.Errorf("fuzzy takes only 2 arguments")
		}
	default:
		err = fmt.Errorf("unknown function %v(%d)", name, len(args))
	}

	return
}

func ExprHead(m *Mem, exprs []Expr) []string {
	var head []string
	for _, e := range exprs {
		if e.Name == "" {
			head = append(head, e.Name)
		} else if idx := strings.Index(e.Name, "."); idx > 0 && (idx+1) < len(e.Name) {
			head = append(head, e.Name[idx+1:])
		} else {
			for _, a := range m.Head(e.Name) {
				head = append(head, a)
			}
		}
	}

	return head
}

func BoolEval(e Expr, m *Mem, t Tuple) bool {
	return Bool(e.Eval(m, t))
}

func NumEval(e Expr, m *Mem, t Tuple) float64 {
	return Num(e.Eval(m, t))
}

func StrEval(e Expr, m *Mem, t Tuple) string {
	return Str(e.Eval(m, t))
}

func (e Expr) Not() Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return !BoolEval(e, m, t)
	}}
}

func (e Expr) Neg() Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return -NumEval(e, m, t)
	}}
}

func (e Expr) Pos() Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return +NumEval(e, m, t)
	}}
}

func (l Expr) Mul(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return NumEval(l, m, t) * NumEval(r, m, t)
	}}
}

func (l Expr) Div(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return NumEval(l, m, t) / NumEval(r, m, t)
	}}
}

func (l Expr) Add(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return NumEval(l, m, t) + NumEval(r, m, t)
	}}
}

func (l Expr) Sub(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return NumEval(l, m, t) - NumEval(r, m, t)
	}}
}

func (l Expr) Cat(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return StrEval(l, m, t) + StrEval(r, m, t)
	}}
}

func (l Expr) LT(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return NumEval(l, m, t) < NumEval(r, m, t)
	}}
}

func (l Expr) GT(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return NumEval(l, m, t) > NumEval(r, m, t)
	}}
}

func (l Expr) LTE(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return NumEval(l, m, t) <= NumEval(r, m, t)
	}}
}

func (l Expr) GTE(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return NumEval(l, m, t) >= NumEval(r, m, t)
	}}
}

func (l Expr) Eq(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return Eq(l.Eval(m, t), r.Eval(m, t))
	}}
}

func (l Expr) NotEq(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return !Eq(l.Eval(m, t), r.Eval(m, t))
	}}
}

func (e Expr) Match(re int) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return m.MatchString(re, StrEval(e, m, t))
	}}
}

func (l Expr) And(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return BoolEval(l, m, t) && BoolEval(r, m, t)
	}}
}

func (l Expr) Or(r Expr) Expr {
	return Expr{"", func(m *Mem, t Tuple) Value {
		return BoolEval(l, m, t) || BoolEval(r, m, t)
	}}
}
