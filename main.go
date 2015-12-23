// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const usage = `comp [-f <files>] <expr>

examples
  cat file.json | comp -f @json '[ i | i <- in, i.name =~ \"hello\" ]'
  comp -f file1.json,file2.csv '[ {i, j} | i <- file1, j <- file2, i.id == j.id ]'

flags
`

func openFiles(files string) (map[string]io.Reader, error) {
	res := make(map[string]io.Reader)
	if files != "" {
		for _, f := range strings.Split(files, ",") {
			if f[0] == '@' {
				f = fmt.Sprintf("in.%v", f[1:])
				res[f] = os.Stdin
			} else {
				r, err := os.Open(f)
				if err != nil {
					return nil, err
				}
				res[f] = r
			}
		}
	}

	return res, nil
}

func Run(expr string, inputs map[string]io.Reader, output io.Writer) error {
	store := Store{make(map[string]Type), make(map[string]Value)}
	for k, v := range inputs {
		if err := store.Add(k, v); err != nil {
			return err
		}
	}

	decls := store.Decls()

	prg, rt, err := Compile(expr, decls)
	if err != nil {
		return err
	}

	res := prg.Run(new(Stack))
	if res != nil {
		if err := res.Quote(output, rt); err != nil {
			return err
		}
		fmt.Printf("\n")
	}

	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}

	files := flag.String("f", "", "comma separated list of files (@json @csv @txt @xml for stdin types)")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		return
	}

	inputs, err := openFiles(*files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	if err := Run(args[0], inputs, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
