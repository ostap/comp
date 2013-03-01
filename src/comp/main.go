package main

import (
	"flag"
	"log"
	"net/http"
	"path"
	"runtime"
	"strings"
)

type Tuple []Value
type Body chan Tuple
type Head map[string]int

func main() {
	bind := flag.String("bind", ":9090", "bind address")
	data := flag.String("data", "", "list of data files")
	peers := flag.String("peers", "", "list of peers (excluding self)")
	cores := flag.Int("cores", runtime.NumCPU(), "how many cores to use locally")
	flag.Parse()

	if *data == "" {
		flag.Usage()
		return
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Printf("running on %d core(s)", runtime.NumCPU())
	log.Printf("adjusting runtime to run on %d cores (old value %d)", *cores, runtime.GOMAXPROCS(*cores))

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

		parts, err := ReadBody(head, fileName, *cores)
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

	log.Printf("announcing /full /part /console /pprof")

	group := NewGroup(store, strings.Split(*peers, ","))
	http.Handle("/full", FullQuery(group))
	http.Handle("/part", PartQuery(group))
	http.Handle("/console", Console(0))
	http.Handle("/pprof/", http.StripPrefix("/pprof/", Profiler(0)))
	log.Fatal(http.ListenAndServe(*bind, nil))
}
