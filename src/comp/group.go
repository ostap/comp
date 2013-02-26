package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Peer string

type Group struct {
	local Store
	peers []Peer
}

func NewGroup(local Store, addrs []string) Group {
	peers := make([]Peer, 0, len(addrs))
	for _, p := range addrs {
		if p = strings.Trim(p, " \t\r\n"); p != "" {
			peers = append(peers, Peer(p))
		}
	}

	return Group{local, peers}
}

func (g Group) FullRun(w io.Writer, query string, limit int) error {
	start := time.Now()

	mem, load, comp, err := Parse(query, g.local)
	if err != nil {
		return fmt.Errorf(`{"error": %v, "line": %v, "column": %v}`, strconv.Quote(err.Error), err.Line, err.Column)
	}

	out := make(Body, 1024)
	stats := make(chan Stats, len(g.peers)+1)
	result := make(chan Stats, 1)

	go func() {
		total, found := g.local.Run(mem, load, comp, out)
		stats <- Stats{total, found}
	}()
	for _, p := range g.peers {
		go p.PartRun(query, limit, out, stats)
	}
	go func() {
		total, found := 0, 0
		for i := 0; i < len(g.peers)+1; i++ {
			if s := <-stats; s != StatsFailed {
				total += s.Total
				found += s.Found
			}
		}
		close(out)
		result <- Stats{total, found}
	}()

	fmt.Fprintf(w, `{"body": [ `)
	found := 0
	for t := range out {
		if limit < 0 || found < limit {
			if found == 0 {
				fmt.Fprintf(w, "[ ")
			} else {
				fmt.Fprintf(w, ", [ ")
			}

			for i, v := range t {
				if i == 0 {
					fmt.Fprintf(w, Quote(v))
				} else {
					fmt.Fprintf(w, ", %v", Quote(v))
				}
			}

			fmt.Fprintf(w, " ]")
		}
		found++
	}

	duration := time.Now().Sub(start)
	millis := duration.Nanoseconds() / 1000 / 1000
	info := <-result

	fmt.Fprintf(w, ` ], "total": %v, "found": %v, "time": "%vms"}%v`, info.Total, info.Found, millis, "\n")
	log.Printf("full run %v, limit %v, %+v, query %v", duration, limit, info, query)
	return nil
}

func (g Group) PartRun(w io.Writer, query string, limit int) error {
	start := time.Now()

	mem, load, comp, err := Parse(query, g.local)
	if err != nil {
		return fmt.Errorf(`{"error": %v, "line": %v, "column": %v}`, strconv.Quote(err.Error), err.Line, err.Column)
	}

	enc := gob.NewEncoder(w)
	out := make(Body, 1024)
	stats := make(chan Stats, 1)
	go func() {
		total, found := g.local.Run(mem, load, comp, out)
		close(out)
		stats <- Stats{total, found}
	}()

	sent := 0
	for t := range out {
		if limit < 0 || sent < limit {
			enc.Encode(t)
			sent++
		}
	}

	info := <-stats
	enc.Encode(Tuple{float64(info.Total), float64(info.Found)})

	duration := time.Now().Sub(start)
	log.Printf("part run %v, limit %v, %+v, query %v", duration, limit, info, query)
	return nil
}

func (p Peer) PartRun(query string, limit int, out Body, stats chan Stats) {
	url := fmt.Sprintf("%v?limit=%d", p, limit)
	resp, err := http.Post(url, "application/x-comp-query", strings.NewReader(query))
	if err != nil {
		log.Printf("remote call failed: %v", err)
		stats <- StatsFailed
	} else {
		defer resp.Body.Close()
		dec := gob.NewDecoder(resp.Body)
		var prev Tuple = nil
		for {
			var t Tuple
			if err := dec.Decode(&t); err != nil {
				total, found := 0, 0
				if err == io.EOF && len(prev) == 2 {
					total = int(Num(prev[0]))
					found = int(Num(prev[1]))
				}
				stats <- Stats{total, found}
				break
			}

			if prev != nil {
				out <- prev
			}

			prev = t
		}
	}
}
