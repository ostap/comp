// Copyright (c) 2013 Ostap Cherkashin. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

// Generic type
type Type interface {
	Name() string
}

// Scalar types represent strings and numbers.
type ScalarType int

func (s ScalarType) Name() string {
	return "scalar"
}

// List type specifies the type of its elements.
type ListType struct {
	Elem Type
}

func (lt ListType) Name() string {
	return "list"
}

// Function type specifies the type of its arguments and return value.
type FuncType struct {
	Return Type
	Args   []Type
}

func (ft FuncType) Name() string {
	return "function"
}

// Object type specifies the types of all its fields.
type ObjectType []struct {
	Name string
	Type Type
}

func (o ObjectType) Name() string {
	return "object"
}

func (o ObjectType) Has(field string) bool {
	return o.Pos(field) > -1
}

func (o ObjectType) Pos(field string) int {
	pos := -1
	for i, e := range o {
		if e.Name == field {
			pos = i
			break
		}
	}

	return pos
}

func (o ObjectType) Type(field string) Type {
	pos := o.Pos(field)
	if pos > -1 {
		return o[pos].Type
	}

	return nil
}

// TypeOfExpr(eid) references the type of an expression.
type TypeOfExpr int64

func (toe TypeOfExpr) Name() string {
	return "typeOfExpr"
}

// TypeOfField{eid, "a"} references the type of a field.
type TypeOfField struct {
	eid  int64
	name string
}

func (tof TypeOfField) Name() string {
	return "typeOfField"
}

// TypeOfFunc("trunc") references the result type of a function.
type TypeOfFunc string

func (tof TypeOfFunc) Name() string {
	return "typeOfFunc"
}

// TypeOfElem(eid) references the element type of a list (expression).
type TypeOfElem int64

func (toe TypeOfElem) Name() string {
	return "typeOfElem"
}

// TypeOfIdent("id") references the type of an identifier.
type TypeOfIdent string

func (toi TypeOfIdent) Name() string {
	return "typeOfIdent"
}
