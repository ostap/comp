// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func Command(expr, ctype string) error {
	log.SetOutput(os.Stderr)

	store := NewStore()
	if ctype != "" {
		if err := store.Add(fmt.Sprintf("in.%v", ctype), os.Stdin); err != nil {
			return err
		}
	}

	decls := store.Decls()
	prg, rt, err := Compile(expr, decls)
	if err != nil {
		return err
	}

	res := prg.Run(new(Stack))
	if err := res.Quote(os.Stdout, rt); err != nil {
		return err
	}

	return nil
}

func Server(bind, data string, cores int, init func(Store)) error {
	log.Printf("running on %d core(s)", runtime.NumCPU())
	log.Printf("adjusting runtime to run on %d cores (old value %d)", cores, runtime.GOMAXPROCS(cores))

	store := NewStore()
	if data != "" {
		for _, fileName := range strings.Split(data, ",") {
			file, err := os.Open(fileName)
			if err != nil {
				log.Printf("%v", err)
				continue
			}

			if err := store.Add(fileName, file); err != nil {
				log.Printf("%v", err)
			}

			file.Close()
		}
	}
	if init != nil {
		init(store)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("garbage collecting (heap ~%vMB)", m.Alloc/1024/1024)
	runtime.GC()
	runtime.ReadMemStats(&m)
	log.Printf("done (heap ~%vMB)", m.Alloc/1024/1024)

	log.Printf("announcing %v /full /console /pprof", bind)

	http.Handle("/full", FullQuery(store))
	http.Handle("/console", Console(0))
	http.Handle("/pprof/", http.StripPrefix("/pprof/", Profiler(0)))

	return http.ListenAndServe(bind, nil)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage:\n")
		fmt.Fprintf(os.Stderr, "  read data from the standard input and execute the expression\n")
		fmt.Fprintf(os.Stderr, "    comp -t <type> <expr>\n")
		fmt.Fprintf(os.Stderr, "  start a server with the specified files\n")
		fmt.Fprintf(os.Stderr, "    comp -l <host:port> -f <files>\n\n")
		fmt.Fprintf(os.Stderr, "examples:\n")
		fmt.Fprintf(os.Stderr, "  cat file.json | comp -t json '[ i | i <- in, i.name =~ \"hello\" ]'\n")
		fmt.Fprintf(os.Stderr, "  comp -l :9090 -f file1.json,file2.csv\n\n")
		fmt.Fprintf(os.Stderr, "flags:\n")
		flag.PrintDefaults()
	}

	bind := flag.String("l", "", "start a server listening on the specified address (host:port)")
	data := flag.String("f", "", "comma separated list of data files (.json, .xml, .csv, .txt)")
	ctype := flag.String("t", "", "content type of stdin (json, xml, csv, txt)")
	cores := flag.Int("c", runtime.NumCPU(), "how many cores to use for processing")
	flag.Parse()

	if *bind == "" {
		args := flag.Args()
		if len(args) != 1 {
			flag.Usage()
		} else if err := Command(args[0], *ctype); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	} else {
		log.Fatal(Server(*bind, *data, *cores, nil))
	}
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
