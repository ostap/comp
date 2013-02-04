package main

import (
	"fmt"
	"log"
	"net/http"
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
		body, err := Parse(obj.Query, v)
		if err != nil {
			obj.Error = err
		} else {
			t := time.Now()
			for t := <-body; t != nil; t = <-body {
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
