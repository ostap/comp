package main

import (
	"fmt"
	"log"
)

type Store struct {
	heads map[string]Head
	parts map[string][][]Tuple
}

type Stats struct {
	Total int
	Found int
}

var StatsFailed = Stats{-1, -1}

func NewStore() Store {
	return Store{make(map[string]Head), make(map[string][][]Tuple)}
}

func (s Store) IsDef(name string) bool {
	return s.heads[name] != nil
}

func (s Store) Add(name string, head Head, parts [][]Tuple) {
	recs := 0
	info := ""
	for i, p := range parts {
		if i == 0 {
			info = fmt.Sprintf("%v", len(p))
		} else {
			info = fmt.Sprintf("%v %v", info, len(p))
		}
		recs += len(p)
	}

	s.heads[name] = head
	s.parts[name] = parts

	log.Printf("stored %v (recs %v | parts %v)", name, recs, info)
}

func (s Store) Run(mem *Mem, load string, comp Comp, out Body) (total, found int) {
	// FIXME: concurrent map access
	parts := s.parts[load]
	stats := make(chan Stats, len(parts))

	for _, part := range parts {
		go worker(part, mem.Clone(), comp, out, stats)
	}

	total = 0
	found = 0
	for i := 0; i < len(parts); i++ {
		info := <-stats
		total += info.Total
		found += info.Found
	}

	return
}

func (s Store) Declare(m *Mem, prefix, name string) {
	m.Declare(prefix, s.heads[name])
}

func worker(part []Tuple, mem *Mem, comp Comp, out Body, stats chan Stats) {
	total, found := 0, 0
	for _, t := range part {
		if t = comp(mem, t); t != nil {
			out <- t
			found++
		}
		total++
	}

	stats <- Stats{total, found}
}
