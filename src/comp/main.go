// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func Start(bind string, cores int, load func() *Store) error {
	log.Printf("running on %d core(s)", runtime.NumCPU())
	log.Printf("adjusting runtime to run on %d cores (old value %d)", cores, runtime.GOMAXPROCS(cores))

	store := load()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("garbage collecting (heap ~%vMB)", m.Alloc/1024/1024)
	runtime.GC()
	runtime.ReadMemStats(&m)
	log.Printf("done (heap ~%vMB)", m.Alloc/1024/1024)

	log.Printf("announcing /full /console /pprof")

	http.Handle("/full", FullQuery(*store))
	http.Handle("/console", Console(0))
	http.Handle("/pprof/", http.StripPrefix("/pprof/", Profiler(0)))

	return http.ListenAndServe(bind, nil)
}

func main() {
	bind := flag.String("bind", ":9090", "bind address")
	data := flag.String("data", "", "list of data files")
	cores := flag.Int("cores", runtime.NumCPU(), "how many cores to use for computation")
	flag.Parse()

	load := func() *Store {
		store := NewStore()
		if *data != "" {
			for _, fileName := range strings.Split(*data, ",") {
				file, err := os.Open(fileName)
				if err != nil {
					log.Printf("%v", err)
					continue
				}
				r := bufio.NewReader(file)

				if err := store.Add(fileName, r); err != nil {
					log.Printf("%v", err)
				}

				file.Close()
			}
		}
		return &store
	}

	log.Fatal(Start(*bind, *cores, load))
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
