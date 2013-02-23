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
	mem   *Mem
	comp  Comp
	out   Body
	stats chan Stats
}

type Stats struct {
	Total int
	Found int
}

var StatsFailed = Stats{-1, -1}

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

func (s Store) Run(mem *Mem, load string, comp Comp, out Body) (total, found int) {
	workers := s.workers[load]
	stats := make(chan Stats, len(workers))

	for _, wq := range workers {
		wq <- WorkUnit{mem.Clone(), comp, out, stats}
	}

	total = 0
	found = 0
	for i := 0; i < len(workers); i++ {
		info := <-stats
		total += info.Total
		found += info.Found
	}

	return
}

func (s Store) Declare(m *Mem, prefix, name string) {
	m.Declare(prefix, s.heads[name])
}

func worker(wq WorkQueue, part []Tuple) {
	for w := range wq {
		total, found := 0, 0
		for _, t := range part {
			if t = w.comp(w.mem, t); t != nil {
				w.out <- t
				found++
			}
			total++
		}
		w.stats <- Stats{total, found}
	}
}
