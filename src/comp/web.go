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
type FullQuery Store

const QueryPage = `<!doctype html>
<html>
  <head>
    <title>Comp Console</title>
    <script type="text/javascript">
    function $(id) { return document.getElementById(id); }
    function info(msg) { $("info").innerHTML = msg; }
    function run() {
      info("processing ...");
      var req = new XMLHttpRequest();
      req.open("POST", "/full", false);
      req.send(JSON.stringify({ expr: $("expr").value, limit: -1 }));

      info("parsing ...");
      var resp = JSON.parse(req.responseText);
      if (resp.error) {
        info(req.responseText);
      } else {
        info(resp.time);
        var html = "<h1>Result</h1>";
        if (Array.isArray(resp.result)) {
          html += "<table style='width:100%'>";
          for (var i = 0; i < resp.result.length; i++) {
            var elem = resp.result[i];

            html += "<tr>";
            for (var k in elem) {
              html += "<td>" + elem[k] + "</td>";
            }
            html += "</tr>";
          }
          html += "</table>";
        } else {
          html += JSON.stringify(resp.result);
        }

        $("result").innerHTML = html;
      }
      return false;
    }
    </script>
  </head>
  <body>
    <h1>Expression</h1>
    <form name="expression" method="POST" onsubmit="return run();">
      <input id="expr" type="text" spellcheck="false" size="120" autofocus></input>
      <input type="submit" value="Run"></input>
    </form>
    <div id="info"></div>
    <div id="result"></div>
  </body>
</html>`

func webFail(w http.ResponseWriter, msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	http.Error(w, msg, http.StatusInternalServerError)
	log.Print(msg)
}

func badReq(w http.ResponseWriter, json string, args ...interface{}) {
	json = fmt.Sprintf(json, args...)
	http.Error(w, json, http.StatusBadRequest)
	log.Print(json)
}

func (c Console) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, QueryPage)
}

func (fq FullQuery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		dec := json.NewDecoder(r.Body)

		var req struct {
			Expr  string `json:"expr"`
			Limit int    `json:"limit"`
		}
		if err := dec.Decode(&req); err != nil {
			badReq(w, `{"error": %v}`, strconv.Quote("invalid request object: "+err.Error()))
			return
		}

		start := time.Now()

		mem := Store(fq).Alloc()
		expr, err := Compile(req.Expr, mem)
		if err != nil {
			info, _ := json.Marshal(err)
			badReq(w, string(info))
			return
		}

		res := expr.Eval(mem)
		fmt.Fprintf(w, `{"result": `)
		if res != nil {
			if err := res.Quote(w); err != nil {
				log.Printf("failed to marshal result: %v", err)
			}
		} else {
			fmt.Fprintf(w, "null")
		}

		dur := time.Now().Sub(start)
		fmt.Fprintf(w, `, "time": "%v"}`, dur)
		log.Printf("%v for '%v'", dur, req.Expr)
	} else {
		badReq(w, `{"error": "%v unsupported method %v"}`, r.URL, r.Method)
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
