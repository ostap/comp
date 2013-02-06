package main

import (
	"fmt"
	"math"
	"strconv"
)

type Value interface{}

func Bool(v Value) bool {
	switch value := v.(type) {
	case bool:
		return value
	case float64:
		if math.IsNaN(value) {
			return false
		}

		return value != 0
	case string:
		return value != ""
	}

	return false
}

func Num(v Value) float64 {
	switch value := v.(type) {
	case bool:
		res := 0.0
		if value {
			res = 1.0
		}

		return res
	case float64:
		return value
	case string:
		res, _ := strconv.ParseFloat(value, 64)
		return res
	}

	return math.NaN()
}

func Str(v Value) string {
	switch value := v.(type) {
	case bool, float64:
		return fmt.Sprintf("%v", value)
	case string:
		return value
	}

	return ""
}

// TODO: check reflexivity, symmetry, transitivity
func Eq(l, r Value) bool {
	switch lv := l.(type) {
	case bool:
		switch rv := r.(type) {
		case bool:
			return lv == rv
		case float64:
			return Num(lv) == rv
		case string:
			return Str(lv) == rv
		}
	case float64:
		switch rv := r.(type) {
		case float64:
			return lv == rv
		case bool, string:
			return lv == Num(rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv == rv
		case bool:
			return Bool(lv) == rv
		case float64:
			return Num(lv) == rv
		}
	}

	return false
}
