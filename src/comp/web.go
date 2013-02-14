package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type WebQuery Store
type RawQuery Store

type QueryReq struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type QueryResp struct {
	Total int     `json:"total"`
	Time  string  `json:"time"`
	Body  []Tuple `json:"body"`
}

const QueryPage = `<!doctype html>
<html>
  <head><title>Comp Query Panel</title></head>
  <body>
    <h1>Query</h1>
    <form method="POST" action="/console">
      <input name="query" type="text" spellcheck="false" value="{{html .Query}}" size="120"></input>
      <input name="run" type="submit" value="Run"></input>
    </form>

    {{if .Time}}
    <h1>Stats</h1>
    <p>{{.Time}} {{len .Body}} records</p>
    {{end}}

    {{if .Error}}
    <h1>Error</h1>
    <p style="color:red">{{.Error}}</p>
    {{end}}

    {{if len .Body}}
    <h1>Results</h1>
    <table width="100%">
      {{range .Body}}<tr>{{range .}}<td>{{.}}</td>{{end}}</tr>
      {{end}}
    </table>
    {{end}}
  </body>
</html>`

func webFail(w http.ResponseWriter, msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	http.Error(w, msg, http.StatusInternalServerError)
	log.Print(msg)
}

func readJSON(r *http.Request, data interface{}) *ParseError {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read json: %v", err)
		return NewError(-1, -1, "failed to read json")
	}

	err = json.Unmarshal(body, data)
	if err != nil {
		log.Printf("failed to unmarshal json: %v", err)
		return NewError(-1, -1, "failed to unmarshal json")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, data interface{}) *ParseError {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal json: %v", err)
		return NewError(-1, -1, "failed to marshal json")
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(msg)
	if err != nil {
		log.Printf("failed to write json response: %v", err)
		return NewError(-1, -1, "failed to write json response")
	}

	return nil
}

func (wq WebQuery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	store := Store(wq)
	var obj struct {
		Query string
		Error error
		Body  []Tuple
		Time  time.Duration
	}
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			webFail(w, "invalid form submission: %v", err)
			return
		}

		obj.Query = r.Form.Get("query")
		mem, load, comp, err := Parse(obj.Query, store)
		if err != nil {
			obj.Error = fmt.Errorf(err.Error)
		} else {
			t := time.Now()
			for t := range store.Run(mem, load, comp) {
				obj.Body = append(obj.Body, t)
			}
			obj.Time = time.Now().Sub(t)

			log.Printf("%v for %v", obj.Time, obj.Query)
		}
	}

	t := template.Must(template.New("QueryPage").Parse(QueryPage))
	if err := t.Execute(w, obj); err != nil {
		webFail(w, "failed to execute template: %v", err)
		return
	}
}

func (rq RawQuery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	store := Store(rq)
	if r.Method == "POST" {
		start := time.Now()
		var req QueryReq
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, err)
		}

		mem, load, comp, err := Parse(req.Query, store)
		if err != nil {
			log.Printf("failed to parse '%v': %v", req.Query, err)
			writeJSON(w, err)
			return
		}

		body := make([]Tuple, 0)
		count := 0
		for t := range store.Run(mem, load, comp) {
			if count < req.Limit {
				body = append(body, t)
			}
			count++
		}

		writeJSON(w, &QueryResp{Total: count, Time: time.Now().Sub(start).String(), Body: body})
	} else {
		webFail(w, "unsupported method %v", r.Method)
		return
	}
}

type Profiler struct {
}

// See pprof_remote_servers.html bundled with the gperftools.
func (p *Profiler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "cmdline":
		for _, arg := range os.Args {
			fmt.Fprintf(w, "%v\n", arg)
		}
	case "profile":
		sec := r.URL.Query()["seconds"]
		if len(sec) > 0 {
			dur, _ := strconv.Atoi(sec[0])
			buf := new(bytes.Buffer)
			pprof.StartCPUProfile(buf)
			time.Sleep(time.Duration(dur) * time.Second)
			pprof.StopCPUProfile()

			buf.WriteTo(w)
		} else {
			webFail(w, "invalid profile request, expected seconds=XX")
		}
	case "memstats":
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		buf, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			webFail(w, "failed to marshal object: %v", err)
		} else {
			w.Write(buf)
		}
	case "symbol":
		if r.Method == "GET" {
			fmt.Fprintf(w, "num_symbols: 1")
			return
		}

		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			webFail(w, "failed to read request body: %v", err)
			return
		}

		for _, strAddr := range strings.Split(string(buf), "+") {
			strAddr = strings.Trim(strAddr, " \r\n\t")
			desc := "unknownFunc"
			addr, err := strconv.ParseUint(strAddr, 0, 64)
			if err == nil {
				fn := runtime.FuncForPC(uintptr(addr))
				if fn != nil {
					file, line := fn.FileLine(uintptr(addr))
					desc = fmt.Sprintf("%v:%v:%v", path.Base(file), line, fn.Name())
				}
			}
			fmt.Fprintf(w, "%v\t%v\n", strAddr, desc)
		}
	case "":
		for _, p := range pprof.Profiles() {
			fmt.Fprintf(w, "%v\n", p.Name())
		}
	default:
		for _, p := range pprof.Profiles() {
			if p.Name() == r.URL.Path {
				p.WriteTo(w, 0)
				return
			}
		}
		webFail(w, "unknown profile: %v", r.URL.Path)
	}
}
