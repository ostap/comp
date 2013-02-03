package main

import (
	"fmt"
	"math"
	"strconv"
)

const (
	Bool   = iota
	Number = iota
	String = iota
)

// TODO: replace Value with {}interface
type Value struct {
	strval  string
	numval  float64
	boolval bool
	kind    int // Bool, Number, String
}

func BoolVal(b bool) *Value {
	return &Value{strval: "", numval: math.NaN(), boolval: b, kind: Bool}
}

func StrVal(s string) *Value {
	return &Value{strval: s, numval: math.NaN(), boolval: false, kind: String}
}

func NumVal(n float64) *Value {
	return &Value{strval: "", numval: n, boolval: false, kind: Number}
}

func (v *Value) Bool() bool {
	switch v.kind {
	case Bool:
		return v.boolval
	case Number:
		if math.IsNaN(v.numval) {
			return false
		}

		return v.numval != 0
	case String:
		return v.strval != ""
	}

	return false
}

func (v *Value) Num() float64 {
	switch v.kind {
	case Number:
		return v.numval
	case String:
		res, _ := strconv.ParseFloat(v.strval, 64)
		return res
	case Bool:
		res := 0.0
		if v.boolval {
			res = 1.0
		}

		return res
	}

	return math.NaN()
}

func (v *Value) Str() string {
	switch v.kind {
	case String:
		return v.strval
	case Number:
		return fmt.Sprintf("%v", v.numval)
	case Bool:
		return fmt.Sprintf("%v", v.boolval)
	}

	return ""
}

// TODO: check reflexivity, symmetry, transitivity
func (v *Value) Eq(arg *Value) bool {
	if v.kind == Number || arg.kind == Number {
		return v.Num() == arg.Num()
	} else if v.kind == String || arg.kind == String {
		return v.Str() == arg.Str()
	} else if v.kind == Bool || arg.kind == Bool {
		return v.Bool() == arg.Bool()
	}

	return false
}
