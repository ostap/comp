package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type line struct {
	lineNo  int
	lineStr string
}

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

	for i := 0; i < split; i++ {
		go TabDelimParser(i, len(head), lines, tuples, ctl)
	}
	go func() {
		for i := 0; i < split; i++ {
			<-ctl
		}
		close(tuples)
	}()

	ticker := time.NewTicker(1 * time.Second)
	parts := make([][]Tuple, split)

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

			p := count % split
			parts[p] = append(parts[p], t)
			count++
		}
	}
	ticker.Stop()

	return parts, nil
}

func TabDelimParser(id, numAttrs int, in chan line, out Body, ctl chan int) {
	count := 0
	for l := range in {
		attrs := strings.Split(l.lineStr[:len(l.lineStr)-1], "\t")
		if len(attrs) > numAttrs {
			log.Printf("line %d: truncating tuple (-%d attrs)", l.lineNo, len(attrs)-numAttrs)
			attrs = attrs[:numAttrs]
		} else if len(attrs) < numAttrs {
			log.Printf("line %d: missing attributes, appending blank strings", l.lineNo)
			for len(attrs) < numAttrs {
				attrs = append(attrs, "")
			}
		}

		tuple := make(Tuple, numAttrs)
		for i, s := range attrs {
			num, err := strconv.ParseFloat(s, 64)
			if err != nil {
				tuple[i] = s
			} else {
				tuple[i] = num
				count++
			}
		}

		out <- tuple
	}

	log.Printf("parser %d found %d numbers\n", id, count)
	ctl <- 1
}
