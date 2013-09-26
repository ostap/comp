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
)

func Command(expr, files string) error {
	log.SetOutput(os.Stderr)

	store, e := BuildStore(files)
	if e != nil {
		return e
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

func Server(bind, files string, cores int, init func(Store)) error {
	log.Printf("running on %d core(s)", runtime.NumCPU())
	log.Printf("adjusting runtime to run on %d cores (old value %d)", cores, runtime.GOMAXPROCS(cores))

	store, _ := BuildStore(files)
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
		fmt.Fprintf(os.Stderr, "  read data from files/stdin and evaluate an expresion\n")
		fmt.Fprintf(os.Stderr, "    comp -f <files> <expr>\n")
		fmt.Fprintf(os.Stderr, "  start a server with the specified files\n")
		fmt.Fprintf(os.Stderr, "    comp -f <files> -l <host:port>\n\n")
		fmt.Fprintf(os.Stderr, "examples:\n")
		fmt.Fprintf(os.Stderr, "  cat file.json | comp -f @json '[ i | i <- in, i.name =~ \"hello\" ]'\n")
		fmt.Fprintf(os.Stderr, "  comp -f file1.json,file2.csv '[ {i, j} | i <- file1, j <- file2, i.id == j.id]'\n")
		fmt.Fprintf(os.Stderr, "  comp -f file1.txt,file2.xml -l :9090\n\n")
		fmt.Fprintf(os.Stderr, "flags:\n")
		flag.PrintDefaults()
	}

	bind := flag.String("l", "", "start a server listening on the specified host:port")
	files := flag.String("f", "", "comma separated list of files (@json @csv @txt @xml for stdin types)")
	cores := flag.Int("c", runtime.NumCPU(), "how many cores to use for processing")
	flag.Parse()

	if *bind == "" {
		args := flag.Args()
		if len(args) != 1 {
			flag.Usage()
		} else if err := Command(args[0], *files); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	} else {
		log.Fatal(Server(*bind, *files, *cores, nil))
	}
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
