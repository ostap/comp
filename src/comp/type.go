package main

// Generic type
type Type interface{}

// Scalar types represent strings and numbers.
type ScalarType int

// List type specifies the type of its elements.
type ListType struct {
	Elem Type
}

// Function type specifies the type of its result value.
type FuncType struct {
	Result Type
}

// Object type specifies the types of all its fields.
type ObjectType []struct {
	Name string
	Type Type
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

// TypeOfField{eid, "a"} references the type of a field.
type TypeOfField struct {
	eid  int64
	name string
}

// TypeOfFunc("trunc") references the result type of a function.
type TypeOfFunc string

// TypeOfElem(eid) references the element type of a list (expression).
type TypeOfElem int64

// TypeOfIdent("id") references the type of an identifier.
type TypeOfIdent string
