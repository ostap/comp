package main

import (
	"fmt"
	"log"
)

type Store struct {
	heads   map[string]Head
	workers map[string][]WorkQueue
}

type WorkQueue chan WorkUnit

type WorkUnit struct {
	comp   Comp
	out    Body
	closer chan int
}

func NewStore() Store {
	return Store{make(map[string]Head), make(map[string][]WorkQueue)}
}

func (s Store) IsDef(name string) bool {
	return s.heads[name] != nil
}

func (s Store) Add(name string, head Head, parts [][]Tuple) {
	workers := make([]WorkQueue, 0)

	recs := 0
	info := ""
	for i, p := range parts {
		wq := make(WorkQueue)
		workers = append(workers, wq)
		go worker(wq, p)

		if i == 0 {
			info = fmt.Sprintf("%v", len(p))
		} else {
			info = fmt.Sprintf("%v %v", info, len(p))
		}
		recs += len(p)
	}

	s.heads[name] = head
	s.workers[name] = workers

	log.Printf("stored %v (recs %v | parts %v)", name, recs, info)
}

func (s Store) Run(load string, comp Comp) Body {
	out := make(Body)
	closer := make(chan int)
	for _, wq := range s.workers[load] {
		wq <- WorkUnit{comp, out, closer}
	}

	workers := len(s.workers[load])
	go func() {
		for i := 0; i < workers; i++ {
			<-closer
		}
		close(out)
	}()

	return out
}

func (s Store) Decl(m *Mem, prefix, name string) {
	m.Decl(prefix, s.heads[name])
}

func worker(wq WorkQueue, part []Tuple) {
	for w := range wq {
		for _, t := range part {
			if t = w.comp(t); t != nil {
				w.out <- t
			}
		}
		w.closer <- 1
	}
}
