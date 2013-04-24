package main

import (
	"fmt"
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

// TODO: check redeclarations
func (d *Decls) Declare(name string, t Type) int {
	d.names[name] = t
	return d.find(name)
}

func (d *Decls) UseIdent(name string) int {
	if d.strict && d.names[name] == nil {
		d.err("unknown identifier '%v'", name)
	}

	return d.find(name)
}

func (d *Decls) UseScalar() int {
	name := fmt.Sprintf("+%d", len(d.idents))
	return d.find(name)
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

func (d *Decls) Verify() []string {
	// check identifiers
	for _, n := range d.idents {
		if n[0] != '+' && d.names[n] == nil {
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
		_, ok := d.resolve(t)
		if !ok {
			d.err("cannot resolve type of '%v'", d.code[eid])
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

	return d.errors
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

	if addr < 0 {
		addr = len(d.idents)
		d.idents = append(d.idents, name)
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
		ft, ok := d.resolve(d.names[string(st)])
		if ok {
			ft, isFunc := ft.(FuncType)
			if isFunc {
				return ft, true
			}
		}

		return nil, false
	case ScalarType:
		return ScalarType(st), true
	case ListType:
		t, ok := d.resolve(st.Elem)
		return ListType{t}, ok
	case FuncType:
		t, ok := d.resolve(st.Result)
		return FuncType{t}, ok
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
