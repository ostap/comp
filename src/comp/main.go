package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"
	"strings"
)

type Head map[string]int

func Start(bind, data string, cores int) error {
	log.Printf("running on %d core(s)", runtime.NumCPU())
	log.Printf("adjusting runtime to run on %d cores (old value %d)", cores, runtime.GOMAXPROCS(cores))

	store := NewStore()
	if data != "" {
		for _, fileName := range strings.Split(data, ",") {
			if err := store.Add(fileName); err != nil {
				log.Printf("%v", err)
			}
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("garbage collecting (heap ~%vMB)", m.Alloc/1024/1024)
	runtime.GC()
	runtime.ReadMemStats(&m)
	log.Printf("done (heap ~%vMB)", m.Alloc/1024/1024)

	log.Printf("announcing /full /part /console /pprof")

	http.Handle("/full", FullQuery(store))
	http.Handle("/console", Console(0))
	http.Handle("/pprof/", http.StripPrefix("/pprof/", Profiler(0)))

	return http.ListenAndServe(bind, nil)
}

func main() {
	bind := flag.String("bind", ":9090", "bind address")
	data := flag.String("data", "", "list of data files")
	cores := flag.Int("cores", runtime.NumCPU(), "how many cores to use for computation")
	flag.Parse()

	log.Fatal(Start(*bind, *data, *cores))
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
