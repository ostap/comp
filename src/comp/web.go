package main

import (
	"bytes"
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

const QueryPage = `<!doctype html>
<html>
  <head><title>Comp Query Panel</title></head>
  <body>
    <h1>Query</h1>
    <form method="POST" action="/">
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

func (v Views) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		comp, err := Parse(obj.Query, v)
		if err != nil {
			obj.Error = err
		} else {
			t := time.Now()
			for t := range comp.Run(v) {
				obj.Body = append(obj.Body, t)
			}
			obj.Time = time.Now().Sub(t)
		}
	}

	t := template.Must(template.New("QueryPage").Parse(QueryPage))
	if err := t.Execute(w, obj); err != nil {
		webFail(w, "failed to execute template: %v", err)
		return
	}
}

func (rv RawViews) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			webFail(w, "invalid form submission: %v", err)
			return
		}

		query := r.Form.Get("query")
		comp, err := Parse(query, Views(rv))
		if err != nil {
			webFail(w, "failed to parse the query: %v", err)
		}

		for t := range comp.Run(Views(rv)) {
			tab := ""
			for _, v := range t {
				fmt.Fprintf(w, "%v%v", tab, v)
				tab = "\t"
			}
			fmt.Fprintf(w, "\n")
		}
	} else {
		webFail(w, "unsupported method %v", r.Method)
		return
	}
}

type Profiler struct {
}

// See pprof_remote_servers.html bundled with the gperftools.
func (p *Profiler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v", r.URL)

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
