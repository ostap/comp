// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this
// source code under the terms of the MIT License found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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

func (s Store) Add(fileName string, r *bufio.Reader) error {
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
		t, v, err = readJSON(r)
	} else if path.Ext(fileName) == ".xml" {
		t, v, err = readXML(r)
	} else {
		t, err = readHead(r)
		if err == nil {
			v, err = readBody(t.(ListType), fileName, r)
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

func parseFloat(s string) interface{} {
	num, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(num) || math.IsInf(num, 0) {
		return s
	} else {
		return num
	}
}

func name(prefix string, n xml.Name) string {
	if n.Space == "" {
		return prefix + n.Local
	}

	return prefix + n.Space + ":" + n.Local
}

func alloc() map[string]interface{} {
	res := make(map[string]interface{})
	res["text()"] = ""
	return res
}

func readXML(r *bufio.Reader) (Type, Value, error) {
	dec := xml.NewDecoder(r)

	value := alloc() /* root element */
	stack := append(make([]*map[string]interface{}, 0), &value)
	names := append(make([]string, 0), "")
	top := 1

	for {
		tok, err := dec.RawToken()
		if err == io.EOF && top == 1 {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			val := alloc()
			n := name("", t.Name)

			for _, v := range t.Attr {
				val[name("@", v.Name)] = parseFloat(v.Value)
			}

			parent := stack[top-1]
			if prev, ok := (*parent)[n]; ok {
				switch e := prev.(type) {
				case []interface{}:
					(*parent)[n] = append(e, val)
				default:
					(*parent)[n] = append(append(make([]interface{}, 0), e), val)
				}
			} else {
				(*parent)[n] = val
			}

			if top >= len(stack) {
				stack = append(stack, &val)
				names = append(names, n)
			} else {
				stack[top] = &val
				names[top] = n
			}
			top++
		case xml.EndElement:
			exp := names[top-1]
			got := name("", t.Name)
			if exp != got {
				return nil, nil, errors.New(fmt.Sprintf("XML syntax error: element <%v> closed by </%v>", exp, got))
			}
			top--
		case xml.CharData:
			parent := stack[top-1]
			(*parent)["text()"] = (*parent)["text()"].(string) + string(t)
		default:
			/* ignoring the following token types
			   xml.Comment
			   xml.ProcInst
			   xml.Directive
			*/
		}
	}

	return traverse(nil, value)
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

func readJSON(r *bufio.Reader) (Type, Value, error) {
	dec := json.NewDecoder(r)

	var data interface{}
	err := dec.Decode(&data) /* reading a single valid JSON value */
	if err != nil {
		return nil, nil, err
	}

	return traverse(nil, data)
}

func readHead(r *bufio.Reader) (ListType, error) {
	str, err := r.ReadString('\n')
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

func readBody(t ListType, fileName string, r *bufio.Reader) (List, error) {
	lines := make(chan line, 1024)
	go func() {
		for lineNo := 0; ; lineNo++ {
			lineStr, _ := r.ReadString('\n')
			if len(lineStr) == 0 {
				break
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
