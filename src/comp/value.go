// Copyright (c) 2013 Ostap Cherkashin. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"math"
	"strconv"
)

type Bool bool
type Number float64
type String string
type List []Value
type Object []Value

var True Value = Bool(true)
var False Value = Bool(false)

type Value interface {
	Bool() Bool
	String() String
	Number() Number
	List() List
	Object() Object

	Quote(w io.Writer, t Type, limit int) error
	// TODO: check reflexivity, symmetry, transitivity
	Equals(v Value) Bool
}

func (b Bool) Bool() Bool {
	return b
}

func (b Bool) String() String {
	if b {
		return String("true")
	}

	return String("")
}

func (b Bool) Number() Number {
	if b {
		return 1.0
	}

	return 0.0
}

func (b Bool) List() List {
	return List{b}
}

func (b Bool) Object() Object {
	return Object{b}
}

func (b Bool) Quote(w io.Writer, t Type, limit int) error {
	var err error
	if b {
		_, err = io.WriteString(w, "true")
	} else {
		_, err = io.WriteString(w, "false")
	}

	return err
}

func (b Bool) Equals(v Value) Bool {
	return b == v.Bool()
}

func (n Number) Bool() Bool {
	if math.IsNaN(float64(n)) {
		return Bool(false)
	}

	return Bool(n != 0)
}

func (n Number) Number() Number {
	return n
}

func (n Number) String() String {
	return String(fmt.Sprintf("%v", n))
}

func (n Number) List() List {
	return List{n}
}

func (n Number) Object() Object {
	return Object{n}
}

func (n Number) Quote(w io.Writer, t Type, limit int) error {
	var err error
	if math.IsInf(float64(n), 0) || math.IsNaN(float64(n)) {
		_, err = fmt.Fprintf(w, `"%v"`, n)
	} else {
		_, err = fmt.Fprintf(w, "%v", n)
	}

	return err
}

func (n Number) Equals(v Value) Bool {
	return n == v.Number()
}

func (s String) Bool() Bool {
	return s != ""
}

func (s String) Number() Number {
	res, _ := strconv.ParseFloat(string(s), 64)
	return Number(res)
}

func (s String) String() String {
	return s
}

func (s String) List() List {
	return List{s}
}

func (s String) Object() Object {
	return Object{s}
}

func (s String) Quote(w io.Writer, t Type, limit int) error {
	_, err := io.WriteString(w, strconv.Quote(string(s)))
	return err
}

func (s String) Equals(v Value) Bool {
	return s == v.String()
}

func (l List) Bool() Bool {
	return len(l) > 0
}

func (l List) Number() Number {
	return Number(math.NaN())
}

func (l List) String() String {
	return ""
}

func (l List) List() List {
	return l
}

func (l List) Object() Object {
	return nil
}

func (l List) Quote(w io.Writer, t Type, limit int) error {
	_, err := io.WriteString(w, "[ ")
	if err != nil {
		return err
	}

	lt, isList := t.(ListType)
	if !isList {
		return fmt.Errorf("internal error: %v is not a list", t.Name())
	}

	cnt := 0
	for i, v := range l {
		if limit >= 0 && limit <= cnt {
			break
		}

		if i != 0 {
			_, err = io.WriteString(w, ", ")
			if err != nil {
				return err
			}
		}

		if err := v.Quote(w, lt.Elem, -1); err != nil {
			return err
		}
		cnt++
	}

	_, err = io.WriteString(w, " ]")
	return err
}

func (l List) Equals(v Value) Bool {
	r := v.List()
	if len(l) != len(r) {
		return false
	}

	for i := 0; i < len(l); i++ {
		if !l[i].Equals(r[i]) {
			return false
		}
	}

	return true
}

func (o Object) Bool() Bool {
	return len(o) > 0
}

func (o Object) Number() Number {
	return Number(math.NaN())
}

func (o Object) String() String {
	return ""
}

func (o Object) List() List {
	return nil
}

func (o Object) Object() Object {
	return o
}

func (o Object) Quote(w io.Writer, t Type, limit int) error {
	_, err := io.WriteString(w, "{ ")
	if err != nil {
		return err
	}

	ot, isObject := t.(ObjectType)
	if !isObject {
		return fmt.Errorf("internal error: %v is not an object", t.Name())
	}

	for i, v := range o {
		if i != 0 {
			_, err = io.WriteString(w, ", ")
			if err != nil {
				return err
			}
		}

		_, err = fmt.Fprintf(w, `"%v": `, ot[i].Name)
		if err != nil {
			return err
		}

		if err := v.Quote(w, ot[i].Type, -1); err != nil {
			return err
		}
	}

	_, err = io.WriteString(w, " }")
	return err
}

func (o Object) Equals(v Value) Bool {
	// FIXME: algorithm assumes the same field ordering for both objects
	r := v.Object()
	if len(o) != len(r) {
		return false
	}

	for i := 0; i < len(o); i++ {
		if !bool(o[i].Equals(r[i])) {
			return false
		}
	}

	return true
}
