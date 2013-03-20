package main

import (
	"bufio"
	"fmt"
	"log"
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
	head map[string]Head
	body map[string]List
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
	return Store{make(map[string]Head), make(map[string]List)}
}

func (s Store) IsDef(name string) bool {
	return s.head[name] != nil
}

func (s Store) Add(fileName string) error {
	name := path.Base(fileName)
	if dot := strings.Index(name, "."); dot > 0 {
		name = name[:dot]
	}

	if !IsIdent(name) {
		return fmt.Errorf("invalid file name: '%v' cannot be used as an identifier (ignoring)", name)
	}

	head, err := readHead(fileName)
	if err != nil {
		return fmt.Errorf("failed to load %v: %v", fileName, err)
	}

	body, err := readBody(head, fileName)
	if err != nil {
		return fmt.Errorf("failed to load %v: %v", fileName, err)
	}

	s.head[name] = head
	s.body[name] = body

	log.Printf("stored %v (recs %v)", name, len(body))
	return nil
}

func (s Store) Alloc() *Mem {
	mem := NewMem()
	for k, v := range s.body {
		mem.Store(k, v, s.head[k])
	}

	return mem
}

func IsIdent(s string) bool {
	ident, _ := regexp.MatchString("^\\w+$", s)
	return ident
}

func readHead(fileName string) (Head, error) {
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

func readBody(head Head, fileName string) (List, error) {
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

	for i := 0; i < runtime.NumCPU(); i++ {
		go tabDelimParser(i, head, lines, tuples, ctl)
	}
	go func() {
		for i := 0; i < runtime.NumCPU(); i++ {
			<-ctl
		}
		close(tuples)
	}()

	ticker := time.NewTicker(1 * time.Second)
	body := make(List, 0)

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

			body = append(body, t)
			count++
		}
	}
	ticker.Stop()

	return body, nil
}

func tabDelimParser(id int, h Head, in chan line, out Body, ctl chan int) {
	head := make([]string, len(h))
	for f, i := range h {
		head[i] = f
	}

	count := 0
	for l := range in {
		fields := strings.Split(l.lineStr[:len(l.lineStr)-1], "\t")
		if len(fields) > len(head) {
			log.Printf("line %d: truncating object (-%d fields)", l.lineNo, len(fields)-len(head))
			fields = fields[:len(head)]
		} else if len(fields) < len(head) {
			log.Printf("line %d: missing fields, appending blank strings", l.lineNo)
			for len(fields) < len(head) {
				fields = append(fields, "")
			}
		}

		obj := make(Object, len(head))
		for i, s := range fields {
			num, err := strconv.ParseFloat(s, 64)
			if err != nil {
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
