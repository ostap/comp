package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	Port = ":9090"
	Addr = "http://localhost" + Port + "/full"
)

func ExampleBools() {
	run("true")
	run("false")

	// Output:
	// true
	// false
}

func ExampleNumbers() {
	run("1")
	run("1e7")
	run("3.1415")

	// Output:
	// 1
	// 1e+07
	// 3.1415
}

func ExampleStrings() {
	run(`"hello"`)
	run("`hello`")

	// Output:
	// "hello"
	// "hello"
}

func ExampleLists() {
	run("[true, false]")
	run("[1,2,3]")
	run(`["a","b","c"]`)

	// Output:
	// [true,false]
	// [1,2,3]
	// ["a","b","c"]
}

func ExampleObjects() {
	run(`{id: 1, name: "foo"}`)
	run(`{id: 1, children: [2, 3]}`)
	run(`{id: 1, name: "foo"}.id`)
	run(`{id: 1, name: "foo"}.name`)
	run(`{id: 1, children: [2,3]}.children`)
	run(`{id: 1, obj: {parent: 1, value: "hello"}}.obj`)
	run(`{id: 1, obj: {parent: 1, value: "hello"}}.obj.value`)
	run(`{id: 1, obj: {parent: 1, value: "hello"}}.obj.value.unknown`)

	// Output:
	// {"id":1,"name":"foo"}
	// {"children":[2,3],"id":1}
	// 1
	// "foo"
	// [2,3]
	// {"parent":1,"value":"hello"}
	// "hello"
	// ""
}

func ExampleFuncs() {
	run("lower(`HELLO`)")
	run("upper(`hello`)")
	run("trim(`  hello  `)")
	run("trunc(1.234)")

	// Output:
	// "hello"
	// "HELLO"
	// "hello"
	// 1
}

func ExampleComps() {
	run("[ i | i <- [ 1, 2, 3 ]]")
	run("[ i - 3 | i <- [ 1, 2, 3, 4, 5 ]]")

	// Output:
	// [{"i":1},{"i":2},{"i":3}]
	// [{"":-2},{"":-1},{"":0},{"":1},{"":2}]
}

func ExampleErrors() {
	run("a")
	run("a + b")

	// Output:
	// unknown identifier(s): [a]
	// unknown identifier(s): [a b]
}

func run(query string) {
	req := fmt.Sprintf(`{"expr": %v}`, strconv.Quote(query))
	resp, err := http.Post(Addr, "application/json", strings.NewReader(req))
	if err != nil {
		log.Fatalf("post failed: %v", err)
		return
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %v", err)
		return
	}

	if resp.StatusCode == 200 {
		// verfiy that we get a valid json response
		var res struct {
			Time   string      `json:"time"`
			Result interface{} `json:"result"`
		}
		if err := json.Unmarshal(buf, &res); err != nil {
			log.Fatalf("failed to unmarshal json: %v", err)
			return
		}

		buf, err = json.Marshal(res.Result)
		if err != nil {
			log.Fatalf("failed to marshal json: %v", err)
			return
		}

		fmt.Printf("%v\n", string(buf))
	} else {
		var res struct {
			Error  string `json:"error"`
			Line   int    `json:"line"`
			Column int    `json:"column"`
		}
		if err := json.Unmarshal(buf, &res); err != nil {
			log.Fatalf("failed to unmarshal json: %v", err)
			return
		}

		fmt.Printf("%v\n", res.Error)
	}
}

func init() {
	go func() {
		if err := Start(Port, "", 4); err != nil {
			log.Fatalf("failed to start comp: %v", err)
		}
	}()
}
