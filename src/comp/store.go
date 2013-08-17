// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this
// source code under the terms of the MIT License found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Body chan Value

type Store struct {
	types  map[string]Type
	values map[string]Value
}

type Stats struct {
	Total int
	Found int
}

type line struct {
	lineNo  int
	lineStr string
}

var StatsFailed = Stats{-1, -1}

func NewStore() Store {
	return Store{make(map[string]Type), make(map[string]Value)}
}

func (s Store) IsDef(name string) bool {
	return s.types[name] != nil
}

func (s Store) Add(fileName string) error {
	name := path.Base(fileName)
	if dot := strings.Index(name, "."); dot > 0 {
		name = name[:dot]
	}

	if !IsIdent(name) {
		return fmt.Errorf("invalid file name: '%v' cannot be used as an identifier (ignoring)", name)
	}

	var t Type
	var v Value
	var err error

	if path.Ext(fileName) == ".json" {
		t, v, err = readJSON(fileName)
	} else {
		t, err = readHead(fileName)
		if err == nil {
			v, err = readBody(t.(ListType), fileName)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to load %v: %v", fileName, err)
	}

	s.types[name] = t
	s.values[name] = v

	switch v.(type) {
	case List:
		log.Printf("stored %v (recs %v)", name, len(v.(List)))
	default:
		log.Printf("stored %v (single object)", name)
	}
	return nil
}

func (s Store) Decls() *Decls {
	decls := NewDecls()
	for k, v := range s.values {
		decls.Declare(k, v, s.types[k])
	}

	decls.AddFunc(FuncTrunc())
	decls.AddFunc(FuncDist())
	decls.AddFunc(FuncTrim())
	decls.AddFunc(FuncLower())
	decls.AddFunc(FuncUpper())
	decls.AddFunc(FuncFuzzy())
	decls.AddFunc(FuncReplace())

	return decls
}

func IsIdent(s string) bool {
	ident, _ := regexp.MatchString("^\\w+$", s)
	return ident
}

func traverse(h Type, v interface{}) (Type, Value, error) {
	switch v.(type) {
	case map[string]interface{}:
		elems := v.(map[string]interface{})
		val := make(Object, len(elems))

		var head ObjectType
		switch h.(type) {
		case ObjectType:
			head = h.(ObjectType)
			if len(head) != len(elems) { /* very strict */
				return nil, nil, errors.New("invalid type")
			}
		case nil:
			head = make(ObjectType, len(elems))
		default:
			return nil, nil, errors.New("invalid type")
		}

		idx := 0
		for name, value := range elems {
			t, v, e := traverse(head.Type(name), value)
			if e != nil {
				return nil, nil, e
			}

			if h == nil {
				head[idx].Name = name
				head[idx].Type = t
			}
			i := head.Pos(name)
			if i < 0 {
				return nil, nil, errors.New("invalid type")
			}
			val[i] = v

			idx++
		}

		return head, val, nil
	case []interface{}:
		elems := v.([]interface{})
		val := make(List, len(elems))

		var head ListType
		switch h.(type) {
		case ListType:
			head = h.(ListType)
		case nil:
			head = ListType{}
		default:
			return nil, nil, errors.New("invalid type")
		}

		for idx, value := range elems {
			t, v, e := traverse(head.Elem, value)
			if e != nil {
				return nil, nil, e
			}

			head.Elem = t
			val[idx] = v
		}

		return head, val, nil
	case bool:
		switch h.(type) {
		case nil, ScalarType:
			return ScalarType(0), Bool(v.(bool)), nil
		default:
			return nil, nil, errors.New("invalid type")
		}
	case float64:
		switch h.(type) {
		case nil, ScalarType:
			return ScalarType(0), Number(v.(float64)), nil
		default:
			return nil, nil, errors.New("invalid type")
		}
	default:
		switch h.(type) {
		case nil, ScalarType:
			return ScalarType(0), String(v.(string)), nil
		default:
			return nil, nil, errors.New("invalid type")
		}
	}
}

func readJSON(fileName string) (Type, Value, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	dec := json.NewDecoder(buf)

	var data interface{}
	err = dec.Decode(&data) /* reading a single valid JSON value */
	if err != nil {
		return nil, nil, err
	}

	return traverse(nil, data)
}

func readHead(fileName string) (ListType, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return ListType{}, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	str, err := buf.ReadString('\n')
	if err != nil {
		return ListType{}, err
	}

	fields := strings.Split(str, "\t")
	res := make(ObjectType, len(fields))
	for i, f := range fields {
		res[i].Name = strings.Trim(f, " \r\n")
		res[i].Type = ScalarType(0)
	}

	return ListType{Elem: res}, nil
}

func readBody(t ListType, fileName string) (List, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make(chan line, 1024)
	go func() {
		buf := bufio.NewReader(file)

		for lineNo := 0; ; lineNo++ {
			lineStr, _ := buf.ReadString('\n')
			if len(lineStr) == 0 {
				break
			}
			if lineNo == 0 {
				continue
			}

			lines <- line{lineNo, lineStr}
		}
		close(lines)
	}()

	tuples := make(Body, 1024)
	ctl := make(chan int)

	ot := t.Elem.(ObjectType)
	for i := 0; i < runtime.NumCPU(); i++ {
		go tabDelimParser(i, ot, lines, tuples, ctl)
	}
	go func() {
		for i := 0; i < runtime.NumCPU(); i++ {
			<-ctl
		}
		close(tuples)
	}()

	ticker := time.NewTicker(1 * time.Second)
	list := make(List, 0)

	count := 0
	stop := false
	for !stop {
		select {
		case <-ticker.C:
			log.Printf("loading %v (%d tuples)", fileName, count)
		case t, ok := <-tuples:
			if !ok {
				stop = true
				break
			}

			list = append(list, t)
			count++
		}
	}
	ticker.Stop()

	return list, nil
}

func tabDelimParser(id int, ot ObjectType, in chan line, out Body, ctl chan int) {
	count := 0
	for l := range in {
		fields := strings.Split(l.lineStr[:len(l.lineStr)-1], "\t")
		if len(fields) > len(ot) {
			log.Printf("line %d: truncating object (-%d fields)", l.lineNo, len(fields)-len(ot))
			fields = fields[:len(ot)]
		} else if len(fields) < len(ot) {
			log.Printf("line %d: missing fields, appending blank strings", l.lineNo)
			for len(fields) < len(ot) {
				fields = append(fields, "")
			}
		}

		obj := make(Object, len(ot))
		for i, s := range fields {
			num, err := strconv.ParseFloat(s, 64)
			if err != nil || math.IsNaN(num) || math.IsInf(num, 0) {
				obj[i] = String(s)
			} else {
				obj[i] = Number(num)
				count++
			}
		}

		out <- obj
	}

	log.Printf("parser %d found %d numbers\n", id, count)
	ctl <- 1
}
