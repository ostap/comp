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
		name = fmt.Sprintf("__tmp_var_%d", len(d.idents))
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
	for _, t := range d.names {
		_, err := d.resolve(t)
		if err != nil {
			d.err("%v", err)
		}
	}

	// resolve expression types
	for eid, t := range d.exprs {
		rt, err := d.resolve(t)
		if err != nil {
			d.err("%v", err)
		} else if eid == resEID {
			resType = rt
		}
	}

	// calculate field positions
	for _, f := range d.fields {
		if *f.pos < 0 {
			t, err := d.resolve(d.exprs[f.eid])
			if err != nil {
				d.err("%v", err)
			} else {
				ot, _ := t.(ObjectType)
				*f.pos = ot.Pos(f.name)
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

func (d *Decls) resolve(t Type) (Type, error) {
	switch st := t.(type) {
	case TypeOfExpr:
		return d.resolve(d.exprs[int64(st)])
	case TypeOfField:
		ot, err := d.resolve(d.exprs[st.eid])
		if err != nil {
			return nil, err
		}

		o, isObject := ot.(ObjectType)
		if !isObject {
			return nil, fmt.Errorf("'%v' is not an object", d.code[st.eid])
		}

		if !o.Has(st.name) {
			return nil, fmt.Errorf("object '%v' does not have field '%v'", d.code[st.eid], st.name)
		}

		return o.Type(st.name), nil
	case TypeOfElem:
		eid := int64(st)
		lt, err := d.resolve(d.exprs[eid])
		if err != nil {
			return nil, err
		}

		l, isList := lt.(ListType)
		if !isList {
			return nil, fmt.Errorf("'%v' is not a list", d.code[eid])
		}

		return l.Elem, nil
	case TypeOfIdent:
		return d.resolve(d.names[string(st)])
	case TypeOfFunc:
		return d.resolve(d.names[string(st)])
	case ScalarType:
		return ScalarType(st), nil
	case ListType:
		t, err := d.resolve(st.Elem)
		if err != nil {
			return nil, err
		}

		return ListType{t}, nil
	case FuncType:
		ret, err := d.resolve(st.Return)
		if err != nil {
			return nil, err
		}

		args := make([]Type, len(st.Args))
		for i, t := range st.Args {
			rt, err := d.resolve(t)
			if err != nil {
				return nil, err
			}

			args[i] = rt
		}

		return FuncType{ret, args}, nil
	case ObjectType:
		attrs := make(map[string]bool)
		ot := make(ObjectType, len(st))
		for i, f := range st {
			if attrs[f.Name] {
				return nil, fmt.Errorf("duplicate attribute '%v' in object literal", f.Name)
			}

			t, err := d.resolve(f.Type)
			if err != nil {
				return nil, err
			}

			ot[i].Name = f.Name
			ot[i].Type = t

			attrs[f.Name] = true
		}

		return ot, nil
	}

	return nil, fmt.Errorf("unknown type %v", t)
}
