package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type Body chan Tuple
type Head map[string]int
type Tuple []string

func ReadHead(fileName string) Head {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("failed to read the header: %v", err)
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	str, err := buf.ReadString('\n')
	if err != nil {
		log.Fatalf("failed to read the first line: %v", err)
	}

	re := regexp.MustCompile("^\\w+$")
	res := make(Head)
	for idx, attr := range strings.Split(str, "\t") {
		attr = strings.Trim(attr, " \r\n")
		if !re.MatchString(attr) {
			log.Fatalf("invalid attribute name: '%v'", attr)
		}
		res[attr] = idx
	}

	return res
}

func LoadFile(head Head, fileName string) ([]Tuple, error) {
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

		tuple := strings.Split(line, "\t")
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
				tuple[i] = e(gHead, t).Str()
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

			if expr(gHead, t).Bool() {
				body <- t
			}
		}
		body <- nil
	}()

	return body
}

func Load(ident string) Body {
	body := make(Body)
	go func() {
		for i := 0; i < len(gBody); i++ {
			body <- gBody[i]
		}
		body <- nil
	}()

	return body
}

var gBody []Tuple
var gHead Head

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %v <file>\n", os.Args[0])
		return
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	// log.Printf("running on %d core(s)", runtime.NumCPU())
	// log.Printf("adjusting runtime (old value %d)", runtime.GOMAXPROCS(runtime.NumCPU()))

	gHead = ReadHead(os.Args[1])

	var err error
	gBody, err = LoadFile(gHead, os.Args[1])
	if err != nil {
		log.Fatalf("failed to load file: %v", err)
	}

	stdin := bufio.NewReader(os.NewFile(0, "stdin"))
	for {
		fmt.Printf("> ")
		line, _ := stdin.ReadString('\n')
		if len(line) == 0 {
			break
		}

		res := Parse(line)
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
