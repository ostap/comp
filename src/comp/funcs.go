// Copyright (c) 2013 Ostap Cherkashin. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"math"
	"strings"
)

type Func struct {
	Name string
	Type FuncType
	Eval func(s *Stack)
}

func FuncTrunc() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0)}}
	return &Func{"trunc", t, func(s *Stack) {
		val := s.PopNum()
		val = math.Trunc(val)
		s.PushNum(val)
	}}
}

func FuncDist() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0), ScalarType(0), ScalarType(0), ScalarType(0)}}
	return &Func{"dist", t, func(s *Stack) {
		lat1 := s.PopNum()
		lon1 := s.PopNum()
		lat2 := s.PopNum()
		lon2 := s.PopNum()

		val := Dist(lat1, lon1, lat2, lon2)

		s.PushNum(val)
	}}
}

func FuncTrim() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0)}}
	return &Func{"trim", t, func(s *Stack) {
		str := s.PopStr()
		str = strings.Trim(str, " \t\r\n")
		s.PushStr(str)
	}}
}

func FuncLower() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0)}}
	return &Func{"lower", t, func(s *Stack) {
		str := s.PopStr()
		str = strings.ToLower(str)
		s.PushStr(str)
	}}
}

func FuncUpper() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0)}}
	return &Func{"upper", t, func(s *Stack) {
		str := s.PopStr()
		str = strings.ToUpper(str)
		s.PushStr(str)
	}}
}

func FuncFuzzy() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0), ScalarType(0)}}
	return &Func{"fuzzy", t, func(s *Stack) {
		se := s.PopStr()
		te := s.PopStr()
		val := Fuzzy(se, te)
		s.PushNum(val)
	}}
}

func FuncReplace() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0), ScalarType(0), ScalarType(0)}}
	return &Func{"replace", t, func(s *Stack) {
		str := s.PopStr()
		from := s.PopStr()
		to := s.PopStr()
		str = strings.Replace(str, from, to, -1)
		s.PushStr(str)
	}}
}
