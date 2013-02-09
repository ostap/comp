package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

type Body chan Tuple
type Head map[string]int
type Tuple []string
type Expr func(m *Mem, t Tuple) Value

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

func ReadBody(head Head, fileName string, split int) ([][]Tuple, error) {
	log.Printf("loading file %v", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	parts := make([][]Tuple, split)
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

		pos := lineNo % split
		parts[pos] = append(parts[pos], tuple)

		if lineNo%100000 == 0 {
			log.Printf("line: %d", lineNo)
		}
	}

	log.Printf("%d lines", lineNo)

	return parts, nil
}

func main() {
	bind := flag.String("bind", "", "bind address, e.g. localhost:9090")
	data := flag.String("data", "", "coma separated list of data files")
	flag.Parse()

	if *bind == "" || *data == "" {
		fmt.Printf("usage: %v -bind localhost:9090 -data file1.txt,file2.txt\n", os.Args[0])
		return
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Printf("running on %d core(s)", runtime.NumCPU())
	log.Printf("adjusting runtime (old value %d)", runtime.GOMAXPROCS(runtime.NumCPU()))

	store := NewStore()
	for _, fileName := range strings.Split(*data, ",") {
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

		parts, err := ReadBody(head, fileName, runtime.NumCPU())
		if err != nil {
			log.Printf("cannot load %v: %v", fileName, err)
			continue
		}

		store.Add(name, head, parts)
	}

	http.Handle("/", WebQuery(store))
	http.Handle("/raw", RawQuery(store))
	http.Handle("/pprof/", http.StripPrefix("/pprof/", new(Profiler)))
	http.ListenAndServe(*bind, nil)
}
