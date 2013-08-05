// Copyright (c) 2013 Ostap Cherkashin. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

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
		code := []Op{OpList()}
		for _, e := range elems {
			for _, c := range e.Code() {
				code = append(code, c)
			}
			code = append(code, OpAppend())
		}

		return code
	}}
}

func ExprComp(loop *Loop, resAddr int) Expr {
	// TODO: compose a name
	return Expr{nextEID(), "", func() []Op {
		return append(loop.Code(), OpLoad(resAddr))
	}}
}

func (e Expr) Field(name string, pos *int) Expr {
	return Expr{nextEID(), fmt.Sprintf("%v.%v", e.Name, name), func() []Op {
		return append(e.Code(), OpGet(*pos))
	}}
}

func (e Expr) Index(name string, pos *int) Expr {
	return Expr{nextEID(), fmt.Sprintf("%v[%v]", e.Name, name), func() []Op {
		return append(e.Code(), OpIndex(*pos))
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

func ExprCall(fn int, args []Expr) Expr {
	// TODO: compose a name
	return Expr{nextEID(), "", func() []Op {
		code := make([]Op, 0)
		for i := len(args) - 1; i > -1; i-- {
			for _, c := range args[i].Code() {
				code = append(code, c)
			}
		}

		return append(code, OpCall(fn))
	}}
}
