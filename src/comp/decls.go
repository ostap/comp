// Copyright (c) 2013 Ostap Cherkashin. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"fmt"
	"regexp"
)

type Decls struct {
	names  map[string]Type
	exprs  map[int64]Type
	code   map[int64]string
	strict bool
	idents []string
	errors []string
	fields []struct {
		eid  int64
		name string
		pos  *int
	}
	sameTypes [][]int64
	values    []Value
	regexps   []*regexp.Regexp
	funcs     []*Func
}

func NewDecls() *Decls {
	res := new(Decls)
	res.names = make(map[string]Type)
	res.exprs = make(map[int64]Type)
	res.code = make(map[int64]string)
	return res
}

func (d *Decls) Strict(on bool) {
	d.strict = on
}

func (d *Decls) Declare(name string, v Value, t Type) (int, error) {
	if name == "" {
		name = fmt.Sprintf("+%d", len(d.idents))
	}

	if d.find(name) > -1 && d.names[name] != nil {
		return -1, fmt.Errorf("'%v' is already declared", name)
	}

	addr := d.insert(name)

	d.names[name] = t
	d.values[addr] = v

	return addr, nil
}

func (d *Decls) AddFunc(fn *Func) {
	d.funcs = append(d.funcs, fn)
	d.names[fn.Name] = fn.Type
}

func (d *Decls) RegExp(pattern string) (int, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return -1, err
	}

	pos := len(d.regexps)
	d.regexps = append(d.regexps, re)

	return pos, nil
}

func (d *Decls) UseIdent(name string) int {
	if d.strict && d.names[name] == nil {
		d.err("unknown identifier '%v'", name)
	}

	return d.insert(name)
}

func (d *Decls) UseFunc(name string, eids []int64) int {
	fn := -1
	for i, _ := range d.funcs {
		if d.funcs[i].Name == name {
			fn = i
			break
		}
	}

	if fn < 0 {
		d.err("unknown function %s", name)
	} else if len(d.funcs[fn].Type.Args) != len(eids) {
		d.err("function %v takes %v arguments", name, len(d.funcs[fn].Type.Args))
	}

	// TODO: check the argument types

	return fn
}

func (d *Decls) UseField(eid int64, name string) *int {
	pos := -1
	for i, f := range d.fields {
		if f.eid == eid && f.name == name {
			pos = i
			break
		}
	}

	if pos < 0 {
		var f struct {
			eid  int64
			name string
			pos  *int
		}
		f.eid = eid
		f.name = name
		f.pos = new(int)
		*f.pos = -1

		pos = len(d.fields)
		d.fields = append(d.fields, f)
	}

	return d.fields[pos].pos
}

func (d *Decls) SameType(eids []int64) {
	d.sameTypes = append(d.sameTypes, eids)
}

func (d *Decls) SetType(e Expr, t Type) {
	d.exprs[e.Id] = t
	d.code[e.Id] = e.Name
}

func (d *Decls) Verify(resEID int64) (Type, []string) {
	var resType Type = nil

	// check identifiers
	for _, n := range d.idents {
		if d.names[n] == nil {
			d.err("unknown identifier '%v'", n)
		}
	}

	// resolve named types
	for n, t := range d.names {
		_, ok := d.resolve(t)
		if !ok {
			d.err("cannot resolve type of '%v'", n)
		}
	}

	// resolve expression types
	for eid, t := range d.exprs {
		rt, ok := d.resolve(t)
		if !ok {
			d.err("cannot resolve type of '%v'", d.code[eid])
		} else if eid == resEID {
			resType = rt
		}
	}

	// calculate field positions
	for _, f := range d.fields {
		if *f.pos < 0 {
			t, ok := d.resolve(d.exprs[f.eid])
			if ok {
				ot, isObject := t.(ObjectType)
				if !isObject {
					d.err("expression '%v' is not an object", d.code[f.eid])
				} else {
					*f.pos = ot.Pos(f.name)
					if *f.pos < 0 {
						d.err("object '%v' has no field '%v'", d.code[f.eid], f.name)
					}
				}
			}
		}
	}

	// TODO: check sameTypes + web_test

	return resType, d.errors
}

func (d *Decls) err(msg string, args ...interface{}) {
	d.errors = append(d.errors, fmt.Sprintf(msg, args...))
}

func (d *Decls) find(name string) int {
	addr := -1
	for i, n := range d.idents {
		if n == name {
			addr = i
			break
		}
	}

	return addr
}

func (d *Decls) insert(name string) int {
	addr := d.find(name)
	if addr < 0 {
		addr = len(d.idents)
		d.idents = append(d.idents, name)
		d.values = append(d.values, nil)
	}

	return addr
}

func (d *Decls) resolve(t Type) (Type, bool) {
	switch st := t.(type) {
	case TypeOfExpr:
		return d.resolve(d.exprs[int64(st)])
	case TypeOfField:
		ot, ok := d.resolve(d.exprs[st.eid])
		if ok {
			ot, isObject := ot.(ObjectType)
			if isObject && ot.Has(st.name) {
				return ot.Type(st.name), true
			}
		}

		return nil, false
	case TypeOfElem:
		lt, ok := d.resolve(d.exprs[int64(st)])
		if ok {
			lt, isList := lt.(ListType)
			if isList {
				return lt.Elem, true
			}
		}

		return nil, false
	case TypeOfIdent:
		return d.resolve(d.names[string(st)])
	case TypeOfFunc:
		return d.resolve(d.names[string(st)])
	case ScalarType:
		return ScalarType(st), true
	case ListType:
		t, ok := d.resolve(st.Elem)
		return ListType{t}, ok
	case FuncType:
		ret, ok := d.resolve(st.Return)
		if ok {
			args := make([]Type, len(st.Args))
			for i, t := range st.Args {
				rt, ok := d.resolve(t)
				if !ok {
					return nil, false
				}

				args[i] = rt
			}

			return FuncType{ret, args}, ok
		}

		return nil, false
	case ObjectType:
		ot := make(ObjectType, len(st))
		for i, f := range st {
			t, ok := d.resolve(f.Type)
			if !ok {
				return nil, false
			}

			ot[i].Name = f.Name
			ot[i].Type = t
		}

		return ot, true
	}

	return nil, false
}
