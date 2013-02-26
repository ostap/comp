package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

type Console int
type Profiler int
type FullQuery Group
type PartQuery Group

const QueryPage = `<!doctype html>
<html>
  <head>
    <title>Comp Console</title>
    <script type="text/javascript">
    function $(id) { return document.getElementById(id); }
    function info(msg) { $("info").innerHTML = msg; }
    function query() {
      info("processing ...");
      var req = new XMLHttpRequest();
      req.open("POST", "/full", false);
      req.send(JSON.stringify({ query: $("query").value, limit: -1 }));

      info("parsing ...");
      var resp = JSON.parse(req.responseText);
      if (resp.error) {
        info(req.responseText);
      } else {
        var msg = "processed " + resp.total + " records, found " + resp.found + " (" + resp.time + ")";
        info(msg + " rendering ...");

        var html = "<h1>Result</h1><table style='width:100%'>";
        for (var i = 0; i < resp.body.length; i++) {
          var t = resp.body[i];

          html += "<tr>";
          for (var j = 0; j < t.length; j++) {
            html += "<td>" + t[j] + "</td>";
          }
          html += "</tr>";
        }
        html += "</table>";
        $("table").innerHTML = html;
        info(msg);
      }
    }
    </script>
  </head>
  <body>
    <h1>Query</h1>
    <input id="query" type="text" spellcheck="false" size="120"></input>
    <input type="button" value="Run" onclick="query();"></input>
    <div id="info"></div>
    <div id="table"></div>
  </body>
</html>`

func webFail(w http.ResponseWriter, msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	msg = fmt.Sprintf(`{"error": %v}`, strconv.Quote(msg))
	http.Error(w, msg, http.StatusInternalServerError)
	log.Print(msg)
}

func (c Console) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, QueryPage)
}

func (fq FullQuery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		dec := json.NewDecoder(r.Body)

		var req struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if err := dec.Decode(&req); err != nil {
			webFail(w, "invalid request object: %v", err)
			return
		}

		// TODO: increase the limits after testing
		if req.Limit < 0 || req.Limit > 100 {
			req.Limit = 100
		}

		Group(fq).FullRun(w, req.Query, req.Limit)
	} else {
		webFail(w, "%v unsupported method %v", r.URL, r.Method)
	}
}

func (pq PartQuery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		query, err := ioutil.ReadAll(r.Body)
		if err != nil {
			webFail(w, "failed to read query: %v", err)
			return
		}

		limit := 0
		if str := r.URL.Query().Get("limit"); str != "" {
			num, err := strconv.ParseInt(str, 10, 64)
			if err == nil {
				limit = int(num)
			}
		}

		Group(pq).PartRun(w, string(query), limit)
	} else {
		webFail(w, "%v unsupported method %v", r.URL, r.Method)
	}
}

// See pprof_remote_servers.html bundled with the gperftools.
func (p Profiler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
