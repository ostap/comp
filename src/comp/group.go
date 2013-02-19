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

func (g Group) FullRun(w io.Writer, query string, limit int) {
	start := time.Now()

	mem, load, comp, err := Parse(query, g.local)
	if err != nil {
		msg := strconv.Quote(err.Error)
		fmt.Fprintf(w, `{"error": %v, "line": %v, "column": %v}`, msg, err.Line, err.Column)
		log.Printf("parse error %+v: %v", err, query)
		return
	}

	out := make(Body, 1024)
	ctl := make(chan int, len(g.peers)+1)

	go g.local.Run(mem, load, comp, out, ctl)
	for _, p := range g.peers {
		go p.PartRun(query, limit, out, ctl)
	}
	go func() {
		for i := 0; i < len(g.peers)+1; i++ {
			<-ctl
		}
		close(out)
	}()

	fmt.Fprintf(w, `{"body": [ `)
	total := 0
	for t := range out {
		if limit < 0 || total < limit {
			if total == 0 {
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
		total++
	}

	duration := time.Now().Sub(start)
	fmt.Fprintf(w, ` ], "total": %v, "time": "%v"}`, total, duration)
	log.Printf("full run %v for %v", duration, query)
}

func (g Group) PartRun(w io.Writer, query string, limit int) {
	start := time.Now()

	mem, load, comp, err := Parse(query, g.local)
	if err != nil {
		msg := strconv.Quote(err.Error)
		fmt.Fprintf(w, `{"error": %v, "line": %v, "column": %v}`, msg, err.Line, err.Column)
		log.Printf("parse error %+v: %v", err, query)
		return
	}

	enc := gob.NewEncoder(w)
	out := make(Body, 1024)
	ctl := make(chan int)
	go g.local.Run(mem, load, comp, out, ctl)
	for {
		select {
		case t := <-out:
			enc.Encode(t)
		case <-ctl:
			goto end
		}
	}

end:
	duration := time.Now().Sub(start)
	log.Printf("part run %v for %v", duration, query)
}

func (p Peer) PartRun(query string, limit int, out Body, ctl chan int) {
	resp, err := http.Post(string(p), "application/x-comp-query", strings.NewReader(query))
	if err != nil {
		log.Printf("remote call failed: %v", err)
		ctl <- -1
	} else {
		defer resp.Body.Close()
		dec := gob.NewDecoder(resp.Body)
		for total := 0; ; total++ {
			var t Tuple
			if err := dec.Decode(&t); err != nil {
				ctl <- total
				break
			}

			out <- t
		}
	}
}
