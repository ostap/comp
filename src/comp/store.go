// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this
// source code under the terms of the MIT License found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
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
	lineNo int
	rec    []string
}

type LineReader interface {
	Read() (rec []string, err error)
}

type TabLineReader struct {
	reader *bufio.Reader
}

func (r *TabLineReader) Read() ([]string, error) {
	b := r.reader
	line, err := b.ReadString('\n')
	if err != nil {
		return nil, err
	}

	return strings.Split(line[:len(line)-1], "\t"), nil
}

var StatsFailed = Stats{-1, -1}

func BuildStore(files string) (Store, error) {
	var err error
	store := Store{make(map[string]Type), make(map[string]Value)}
	if files != "" {
		for _, fileName := range strings.Split(files, ",") {
			var file *os.File
			if fileName[0] == '@' {
				fileName = fmt.Sprintf("in.%v", fileName[1:])
				file = os.Stdin
			} else {
				f, e := os.Open(fileName)
				if e != nil {
					err = e
					continue
				}
				defer f.Close()
				file = f
			}

			if e := store.Add(fileName, file); e != nil {
				err = e
				continue
			}
		}
	}

	return store, err
}

func (s Store) IsDef(name string) bool {
	return s.types[name] != nil
}

func (s Store) Add(fileName string, r io.Reader) error {
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

	switch path.Ext(fileName) {
	case ".json":
		t, v, err = readJSON(r)
	case ".xml":
		t, v, err = readXML(r)
	case ".csv":
		t, v, err = readText(csvReader(r), fileName)
	case ".txt":
		t, v, err = readText(tsvReader(r), fileName)
	default:
		err = fmt.Errorf("unknown content type %v (use one of json, xml, csv, txt)", path.Ext(fileName))
	}

	if err != nil {
		return fmt.Errorf("failed to load %v: %v", fileName, err)
	}

	s.types[name] = t
	s.values[name] = v

	return nil
}

func (s Store) PrintSymbols() {
	log.Printf("available symbols:")
	for n, v := range s.values {
		info := fmt.Sprintf("  %v (%v", n, s.types[n].Name())
		switch value := v.(type) {
		case List:
			info = fmt.Sprintf("%v, %v elements", info, len(value))
		case Object:
			info = fmt.Sprintf("%v, %v fields", info, len(value))
		}
		log.Printf("%v)", info)
	}
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

func toScalar(s string) interface{} {
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

func readXML(r io.Reader) (Type, Value, error) {
	dec := xml.NewDecoder(r)

	value := alloc() /* root element */
	stack := append(make([]map[string]interface{}, 0), value)
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
				val[name("@", v.Name)] = toScalar(v.Value)
			}

			parent := stack[top-1]
			if prev, ok := parent[n]; ok {
				switch e := prev.(type) {
				case []interface{}:
					parent[n] = append(e, val)
				default:
					parent[n] = []interface{}{e, val}
				}
			} else {
				parent[n] = val
			}

			if top >= len(stack) {
				stack = append(stack, val)
				names = append(names, n)
			} else {
				stack[top] = val
				names[top] = n
			}
			top++
		case xml.EndElement:
			exp := names[top-1]
			got := name("", t.Name)
			if exp != got {
				return nil, nil, fmt.Errorf("XML syntax error: element <%v> closed by </%v>", exp, got)
			}
			top--
		case xml.CharData:
			parent := stack[top-1]
			parent["text()"] = parent["text()"].(string) + string(t)
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
	case nil:
		return ScalarType(0), Bool(false), nil
	case map[string]interface{}:
		elems := v.(map[string]interface{})
		val := make(Object, len(elems))

		var head ObjectType
		switch h.(type) {
		case ObjectType:
			head = h.(ObjectType)
			if len(head) != len(elems) { /* very strict */
				return nil, nil, fmt.Errorf("invalid object type, expected %v got %v", head, elems)
			}
		case nil:
			head = make(ObjectType, len(elems))
		default:
			return nil, nil, fmt.Errorf("expected object, got %v (%v)", h.Name(), v)
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
				return nil, nil, fmt.Errorf("cannot find field %v in %v (%v)", name, head, v)
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
			return nil, nil, fmt.Errorf("expected list, got %v (%v)", h.Name(), v)
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
			return nil, nil, fmt.Errorf("expected bool, got %v (%v)", h.Name(), v)
		}
	case float64:
		switch h.(type) {
		case nil, ScalarType:
			return ScalarType(0), Number(v.(float64)), nil
		default:
			return nil, nil, fmt.Errorf("expected number, got %v (%v)", h.Name(), v)
		}
	default:
		switch h.(type) {
		case nil, ScalarType:
			return ScalarType(0), String(v.(string)), nil
		default:
			return nil, nil, fmt.Errorf("expected string, got %v (%v)", h.Name(), v)
		}
	}
}

func readJSON(r io.Reader) (Type, Value, error) {
	dec := json.NewDecoder(r)

	var data interface{}
	err := dec.Decode(&data) /* reading a single valid JSON value */
	if err != nil {
		return nil, nil, err
	}

	return traverse(nil, data)
}

func csvReader(in io.Reader) LineReader {
	r := csv.NewReader(in)
	r.Comma = ','
	r.LazyQuotes = true
	r.TrailingComma = true
	r.FieldsPerRecord = -1

	return r
}

func tsvReader(in io.Reader) LineReader {
	return &TabLineReader{reader: bufio.NewReader(in)}
}

func readText(r LineReader, fileName string) (Type, Value, error) {
	rec, err := r.Read()
	if err != nil {
		return nil, nil, err
	}

	head := make(ObjectType, len(rec))
	for i, f := range rec {
		head[i].Name = strings.Trim(f, " \r\n")
		head[i].Type = ScalarType(0)
	}

	t := ListType{Elem: head}
	return t, readBody(t, fileName, r), nil
}

func readBody(t ListType, fileName string, r LineReader) List {
	lines := make(chan line, 1024)
	go func() {
		for lineNo := 0; ; lineNo++ {
			rec, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("failed to parse %v, %v", fileName, err)
				break
			}
			lines <- line{lineNo, rec}
		}
		close(lines)
	}()

	tuples := make(Body, 1024)
	ctl := make(chan int)

	ot := t.Elem.(ObjectType)
	for i := 0; i < runtime.NumCPU(); i++ {
		go processLine(i, ot, lines, tuples, ctl)
	}
	go func() {
		for i := 0; i < runtime.NumCPU(); i++ {
			<-ctl
		}
		close(tuples)
	}()

	ticker := time.NewTicker(3 * time.Second)
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

	return list
}

func processLine(id int, ot ObjectType, in chan line, out Body, ctl chan int) {
	count := 0
	for l := range in {
		fields := l.rec
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

	ctl <- 1
}
