// Copyright (c) 2013 Ostap Cherkashin. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

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
	io.WriteString(w, ConsolePage)
}

func (fq FullQuery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	if r.Method == "POST" {
		dec := json.NewDecoder(r.Body)

		var req struct {
			Expr string `json:"expr"`
		}
		if err := dec.Decode(&req); err != nil {
			badReq(w, `{"error": %v}`, strconv.Quote("invalid request object: "+err.Error()))
			return
		}

		start := time.Now()

		decls := Store(fq).Decls()
		prg, rt, err := Compile(req.Expr, decls)
		if err != nil {
			info, _ := json.Marshal(err)
			log.Printf("compilation error '%v'", req.Expr)
			badReq(w, string(info))
			return
		}

		res := prg.Run(new(Stack))
		fmt.Fprintf(w, `{"result": `)
		if res != nil {
			if err := res.Quote(w, rt); err != nil {
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

const ConsolePage = `<!DOCTYPE html>
<html>
  <head>
    <title>comp console</title>
    <style>
    html, body {
        background-color: #333;
        color: white;
        font-family: monospace;
        margin: 0;
        padding: 0;
    }
    #console {
        background-color: black;
        margin: 0 auto;
    }
    .jqconsole {
        padding: 10px;
    }
    .jqconsole-cursor {
        background-color: gray;
    }
    /* when the console looses focus */
    .jqconsole-blurred .jqconsole-cursor {
        background-color: #666;
    }
    .jqconsole-prompt {
        color: #0d0;
    }
    /* command history */
    .jqconsole-old-prompt {
        color: #0b0;
        font-weight: normal;
    }
    /* text color when in input mode */
    .jqconsole-input {
        color: #dd0;
    }
    /* previously entered input */
    .jqconsole-old-input {
        color: #bb0;
        font-weight: normal;
    }
    .jqconsole-output {
        color: white;
    }
    .jqconsole-info {
        color: green;
    }
    </style>
  </head>
  <body>
    <div id="console"></div>
    <script src="//ajax.googleapis.com/ajax/libs/jquery/2.0.3/jquery.min.js"
            type="text/javascript" charset="utf-8"></script>
    <script src="//code.jquery.com/jquery-migrate-1.2.1.min.js"
            type="text/javascript" charset="utf-8"></script>
    <script src="//cdnjs.cloudflare.com/ajax/libs/jq-console/2.7.7/jqconsole.min.js"
            type="text/javascript" charset="utf-8"></script>
    <script>
    $(function () {
        var jqconsole = $('#console').jqconsole('', '> ');
        var startPrompt = function () {
            jqconsole.Prompt(true, function (input) {
                if (input.trim() != '') {
                    $.ajax({
                        type:        'POST',
                        url:         '/full',
                        data:        JSON.stringify({ expr: input }),
                        processData: false,
                        contentType: 'application/json; charset=UTF-8',
                    }).done(function(data) {
                        jqconsole.Write(JSON.stringify(data.result, null, '  ') + '\n\n', 'jqconsole-output');
                        jqconsole.Write(data.time + '\n', 'jqconsole-info');
                    }).fail(function(xhr, type, err) {
                        jqconsole.Write(type + ': ' + err + ': ' + xhr.responseText + '\n');
                    });
                }

                startPrompt();
            });
        };
        startPrompt();
    });
    </script>
  </body>
</html>`
