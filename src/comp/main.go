package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
)

type Body chan Tuple
type Head map[string]int
type Tuple []Value
type Expr func(m *Mem, t Tuple) Value

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

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("garbage collecting (heap ~%vMB)", m.Alloc/1024/1024)
	runtime.GC()
	runtime.ReadMemStats(&m)
	log.Printf("done (heap ~%vMB)", m.Alloc/1024/1024)

	http.Handle("/", WebQuery(store))
	http.Handle("/raw", RawQuery(store))
	http.Handle("/pprof/", http.StripPrefix("/pprof/", new(Profiler)))
	http.ListenAndServe(*bind, nil)
}
