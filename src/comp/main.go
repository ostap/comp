package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Body chan Tuple
type Head map[string]int
type Tuple []string

func IsIdent(s string) bool {
	ident, _ := regexp.MatchString("^\\w+$", s)
	return ident
}

func ReadHead(fileName string) (Head, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	str, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}

	res := make(Head)
	for idx, attr := range strings.Split(str, "\t") {
		attr = strings.Trim(attr, " \r\n")
		if !IsIdent(attr) {
			return nil, fmt.Errorf("invalid attribute name: '%v'", attr)
		}
		res[attr] = idx
	}

	return res, nil
}

func ReadBody(head Head, fileName string) ([]Tuple, error) {
	log.Printf("loading file %v", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	body := make([]Tuple, 0)
	buf := bufio.NewReader(file)
	lineNo := 0
	for ; ; lineNo++ {
		line, _ := buf.ReadString('\n')
		if len(line) == 0 {
			break
		}
		if lineNo == 0 {
			continue
		}

		tuple := strings.Split(line[:len(line)-1], "\t")
		if len(tuple) > len(head) {
			tuple = tuple[:len(head)]
		}

		for len(tuple) < len(head) {
			tuple = append(tuple, "")
		}

		body = append(body, tuple)

		if lineNo%100000 == 0 {
			log.Printf("line: %d", lineNo)
		}
	}

	log.Printf("%d lines", lineNo)

	return body, nil
}

func (r Body) Return(exprs []Expr) Body {
	body := make(Body)
	go func() {
		for {
			t := <-r
			if t == nil {
				break
			}

			tuple := make(Tuple, len(exprs))
			for i, e := range exprs {
				tuple[i] = e(t).Str()
			}

			body <- tuple
		}
		body <- nil
	}()

	return body
}

func (r Body) Select(expr Expr) Body {
	body := make(Body)
	go func() {
		for {
			t := <-r
			if t == nil {
				break
			}

			if expr(t).Bool() {
				body <- t
			}
		}
		body <- nil
	}()

	return body
}

type Views struct {
	heads  map[string]Head
	bodies map[string][]Tuple
}

func NewViews() Views {
	return Views{heads: make(map[string]Head), bodies: make(map[string][]Tuple)}
}

func (v Views) Store(name string, h Head, b []Tuple) {
	v.heads[name] = h
	v.bodies[name] = b
}

func (v Views) Has(name string) bool {
	return v.heads[name] != nil && v.bodies[name] != nil
}

func (v Views) Load(name string) (Head, Body) {
	if !v.Has(name) {
		return nil, nil
	}

	body := make(Body)
	go func() {
		for _, t := range v.bodies[name] {
			body <- t
		}
		body <- nil
	}()

	return v.heads[name], body
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %v file1.txt ... fileN.txt\n", os.Args[0])
		return
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	// log.Printf("running on %d core(s)", runtime.NumCPU())
	// log.Printf("adjusting runtime (old value %d)", runtime.GOMAXPROCS(runtime.NumCPU()))

	views := NewViews()
	for _, fileName := range os.Args[1:] {
		name := path.Base(fileName)
		if dot := strings.Index(name, "."); dot > 0 {
			name = name[:dot]
		}

		if !IsIdent(name) {
			log.Printf("invalid file name: '%v' cannot be used as an identifier (ignoring)", name)
			continue
		}

		head, err := ReadHead(fileName)
		if err != nil {
			log.Printf("cannot load %v: %v", fileName, err)
			continue
		}

		body, err := ReadBody(head, fileName)
		if err != nil {
			log.Printf("cannot load %v: %v", fileName, err)
			continue
		}

		views.Store(name, head, body)
	}

	stdin := bufio.NewReader(os.NewFile(0, "stdin"))
	for {
		fmt.Printf("> ")
		line, _ := stdin.ReadString('\n')
		if len(line) == 0 {
			break
		}

		res := Parse(line, views)
		if res == nil {
			continue
		}

		t := time.Now()
		count := 0
		for t := <-res; t != nil; t = <-res {
			fmt.Printf("%v\n", t)
			count++
		}
		fmt.Printf("--- %d results (%v)\n", count, time.Now().Sub(t))
	}
}
