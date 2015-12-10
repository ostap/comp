// Copyright (c) 2013 Ostap Cherkashin. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"math"
	"strings"
)

type State interface{}

type Func struct {
	Name string
	Type FuncType
	Eval func(state State, s *Stack)

	InitState func() State
	State     State
}

func noop() State {
	return nil
}

func FuncTrunc() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0)}}
	return &Func{"trunc", t, func(state State, s *Stack) {
		val := s.PopNum()
		val = math.Trunc(val)
		s.PushNum(val)
	}, noop, nil}
}

func FuncDist() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0), ScalarType(0), ScalarType(0), ScalarType(0)}}
	return &Func{"dist", t, func(state State, s *Stack) {
		lat1 := s.PopNum()
		lon1 := s.PopNum()
		lat2 := s.PopNum()
		lon2 := s.PopNum()

		val := Dist(lat1, lon1, lat2, lon2)

		s.PushNum(val)
	}, noop, nil}
}

func FuncTrim() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0)}}
	return &Func{"trim", t, func(state State, s *Stack) {
		str := s.PopStr()
		str = strings.Trim(str, " \t\r\n")
		s.PushStr(str)
	}, noop, nil}
}

func FuncLower() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0)}}
	return &Func{"lower", t, func(state State, s *Stack) {
		str := s.PopStr()
		str = strings.ToLower(str)
		s.PushStr(str)
	}, noop, nil}
}

func FuncUpper() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0)}}
	return &Func{"upper", t, func(state State, s *Stack) {
		str := s.PopStr()
		str = strings.ToUpper(str)
		s.PushStr(str)
	}, noop, nil}
}

func FuncFuzzy() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0), ScalarType(0)}}
	return &Func{"fuzzy", t, func(state State, s *Stack) {
		f := state.(*Fuzzy)

		se := s.PopStr()
		te := s.PopStr()
		val := f.Compare(se, te)
		s.PushNum(val)
	}, func() State {
		return new(Fuzzy)
	}, new(Fuzzy)}
}

func FuncReplace() *Func {
	t := FuncType{ScalarType(0), []Type{ScalarType(0), ScalarType(0), ScalarType(0)}}
	return &Func{"replace", t, func(state State, s *Stack) {
		str := s.PopStr()
		from := s.PopStr()
		to := s.PopStr()
		str = strings.Replace(str, from, to, -1)
		s.PushStr(str)
	}, noop, nil}
}
