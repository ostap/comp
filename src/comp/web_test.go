// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
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
	run("!true")
	run("!false")
	run("true && true")
	run("true && false")
	run("false && true")
	run("false && false")
	run("true || true")
	run("true || false")
	run("false || true")
	run("false || false")

	// Output:
	// true
	// false
	// false
	// true
	// true
	// false
	// false
	// false
	// true
	// true
	// true
	// false
}

func ExampleNumbers() {
	run("1")
	run("1e7")
	run("3.1415")
	run("-3.1415")
	run("- 3.1415")
	run("+3.1415")
	run("+ 3.1415")
	run("1 + 2.1415")
	run("2.1415 + 1")
	run("3 - 1")
	run("1 - 3")
	run("3 * 4")
	run("4 * 3")
	run("8 / 2")
	run("2 / 8")
	run("1 + 2 * 3 - 10 / 2")

	// Output:
	// 1
	// 1e+07
	// 3.1415
	// -3.1415
	// -3.1415
	// 3.1415
	// 3.1415
	// 3.1415
	// 3.1415
	// 2
	// -2
	// 12
	// 12
	// 4
	// 0.25
	// 2
}

func ExampleStrings() {
	run(`"hello"`)
	run("`hello`")
	run("`hello` ++ ` world`")
	run("`hello` ++ 1")
	run("2 ++ `hello`")

	// Output:
	// "hello"
	// "hello"
	// "hello world"
	// "hello1"
	// "2hello"
}

func ExampleComparisons() {
	run("-2 < -1")
	run("-1 < 0")
	run("0 < 1")
	run("1 < 2")
	run("2 < 1")
	run("1 < 0")
	run("0 < -1")
	run("-1 < -2")
	run("0")

	run("2 > 1")
	run("1 > 0")
	run("0 > -1")
	run("-1 > -2")
	run("-2 > -1")
	run("-1 > 0")
	run("0 > 1")
	run("1 > 2")
	run("1")

	run("-2 <= -1")
	run("-1 <= 0")
	run("0 <= 1")
	run("1 <= 2")
	run("2 <= 1")
	run("1 <= 0")
	run("0 <= -1")
	run("-1 <= -2")
	run("2")

	run("2 >= 1")
	run("1 >= 0")
	run("0 >= -1")
	run("-1 >= -2")
	run("-2 >= -1")
	run("-1 >= 0")
	run("0 >= 1")
	run("1 >= 2")
	run("3")

	run("-2 <= -2")
	run("-2 >= -2")
	run("0 <= 0")
	run("0 >= 0")
	run("2 <= 2")
	run("2 >= 2")
	run("4")

	run("-1.24e10 < -1.23e10 && 0 <= 1.23e3 && 1.23e3 >= 1.23e3")
	run("-1.24e10 < -1.23e10 && 0 >= 1.23e3 && 1.23e3 >= 1.23e3")
	run("5")

	// Output:
	// true
	// true
	// true
	// true
	// false
	// false
	// false
	// false
	// 0
	// true
	// true
	// true
	// true
	// false
	// false
	// false
	// false
	// 1
	// true
	// true
	// true
	// true
	// false
	// false
	// false
	// false
	// 2
	// true
	// true
	// true
	// true
	// false
	// false
	// false
	// false
	// 3
	// true
	// true
	// true
	// true
	// true
	// true
	// 4
	// true
	// false
	// 5
}

func ExampleEqualityNumbers() {
	run("-1 == -1")
	run("-1 != -1")
	run("0 == 0")
	run("0 != 0")
	run("1 == 1")
	run("1 != 1")

	run("-1 != -2")
	run("-1 == -2")

	// Output:
	// true
	// false
	// true
	// false
	// true
	// false
	// true
	// false
}

func ExampleEqualityStrings() {
	run("`` == ``")
	run("`` != ``")
	run("`hello world` == `hello world`")
	run("`hello world` != `hello world`")

	run("`` != `hello world`")
	run("`` == `hello world`")

	// Output:
	// true
	// false
	// true
	// false
	// true
	// false
}

func ExampleEqualityReflexivityNumbers() {
	run("2 - 1 == 2 - 1")
	run("2 - 1 != 2 - 1")

	// Output:
	// true
	// false
}

func ExampleEqualitySymmetryNumbers() {
	run("1 == 3 - 2")
	run("1 != 3 - 2")
	run("3 - 2 == 1")
	run("3 - 2 != 1")

	// Output:
	// true
	// false
	// true
	// false
}

func ExampleEqualitySymmetryWithCoercions() {
	run("1 == `1`")
	run("1 != `1`")
	run("`1` == 1")
	run("`1` != 1")

	run("0 == ``")
	run("`` == 0")

	// Output:
	// true
	// false
	// true
	// false
	// false
	// false
}

func ExampleEqualityTransitivityNumbers() {
	run("2 - 1 == 3 - 2")
	run("2 - 1 != 3 - 2")
	run("3 - 2 == 4 - 3")
	run("3 - 2 != 4 - 3")
	run("2 - 1 == 4 - 3")
	run("2 - 1 != 4 - 3")

	// Output:
	// true
	// false
	// true
	// false
	// true
	// false
}

func ExampleRegexps() {
	run("`catdog` =~ `dog`")
	run("`catdog` =~ `dogma`")
	run("`catdog` =~ `c.....`")

	// Output:
	// true
	// false
	// true
}

func ExampleLists() {
	run("[true, false]")
	run("[1,2,3]")
	run(`["a","b","c"]`)
	run(`["a","b","c"][0]`)
	run(`["a","b","c"][3]`)
	run(`["a","b","c"][1.999]`)
	run(`[{id:0},{id:1},{id:2}][1]`)
	run(`[{a: "a"}, {"b"}, {"c"}]`)
	run(`[{"a"}, {"b"}, {"c"}]`)

	// Output:
	// [true,false]
	// [1,2,3]
	// ["a","b","c"]
	// "a"
	// ""
	// "b"
	// {"id":1}
	// [{"a":"a"},{"a":"b"},{"a":"c"}]
	// [{"\"a\"":"a"},{"\"a\"":"b"},{"\"a\"":"c"}]
}

func ExampleObjects() {
	run(`{"foo"}`)
	run(`{"foo"}["\"foo\""]`)
	run(`{1}`)
	run(`{1}["1"]`)
	run(`{id: 1, name: "foo"}`)
	run(`{id: 1, children: [2, 3]}`)
	run(`{id: 1, name: "foo"}.id`)
	run(`{id: 1, name: "foo"}["id"]`)
	run(`{id: 1, name: "foo"}.name`)
	run(`{id: 1, name: "foo"}["name"]`)
	run(`{id: 1, children: [2,3]}.children`)
	run(`{id: 1, obj: {parent: 1, value: "hello"}}.obj`)
	run(`{id: 1, obj: {parent: 1, value: "hello"}}.obj.value`)

	// Output:
	// {"\"foo\"":"foo"}
	// "foo"
	// {"1":1}
	// 1
	// {"id":1,"name":"foo"}
	// {"children":[2,3],"id":1}
	// 1
	// 1
	// "foo"
	// "foo"
	// [2,3]
	// {"parent":1,"value":"hello"}
	// "hello"
}

func ExampleComps() {
	run("[i | i <- [1, 2, 3]]")
	run("[i | i <- [1, 2, 3], i != 2]")
	run("[i | i <- [1, 2, 3], i != 0, i != 2, i != 100, i != 3, i != 200]")
	run("[i + j + k + l | i <- [1], j <- [3], k <- [5], l <- [7]]")
	run("[i - 3 | i <- [1, 2, 3, 4, 5]]")
	run("[i + 1 | i <- [j - 1 | j <- [1, 2, 3]]]")
	run("[{i: i + 1, j: i} | i <- [j - 1 | j <- [1, 2, 3]]]")
	run("[i * j | i <- [1, 2, 3], j <- [10, 20]]")
	run("[i * j | i <- [1, 2, 3], j <- [10, 20], i == j / 10]")
	run("[i * j | i <- [1, 2, 3], trunc(i), j <- [10, 20]]")
	run(`[ i["a"] | i <- [{a: "a"}, {"b"}, {"c"}]]`)
	run(`[ i["\"a\""] | i <- [{"a"}, {"b"}, {"c"}]]`)

	// Output:
	// [1,2,3]
	// [1,3]
	// [1]
	// [16]
	// [-2,-1,0,1,2]
	// [1,2,3]
	// [{"i":1,"j":0},{"i":2,"j":1},{"i":3,"j":2}]
	// [10,20,20,40,30,60]
	// [10,40]
	// [10,20,20,40,30,60]
	// ["a","b","c"]
	// ["a","b","c"]
}

func ExampleFuncs() {
	run("lower(`HELLO`)")
	run("upper(`hello`)")
	run("trim(`  hello  `)")
	run("trunc(1.234)")
	run(`replace(" 123 456", " ", "")`)

	// Output:
	// "hello"
	// "HELLO"
	// "hello"
	// 1
	// "123456"
}

func ExampleErrors() {
	run("a")
	run("b + a")
	run("[i | j <- [1, 2, 3]]")
	run("[i * j | i <- [0, 1, 2, 3], trunc(j), j <- [10, 20]]")
	run(`{id: 1, obj: {parent: 1, value: "hello"}}.obj.unknown`)
	run(`{id: 1, obj: {parent: 1, value: "hello"}}.obj.value.unknown`)
	run(`[i | i <- [1, 2, 3], i <- [1, 2, 3]]`)
	run(`[i | i <- 3 + 5]`)
	run(`{3, 3}`)

	// Output:
	// unknown identifier 'a'
	// unknown identifier 'b'
	// unknown identifier 'i'
	// unknown identifier 'j'
	// object '{id, obj}.obj' does not have field 'unknown'
	// '{id, obj}.obj.value' is not an object
	// 'i' is already declared
	// '3 + 5' is not a list
	// duplicate attribute '3' in object literal
}

func ExampleArguments() {
	args := make(map[string]interface{})
	args["num"] = 1
	args["str"] = "hello"
	args["list"] = []int{1, 2, 3}
	args["obj"] = map[string]interface{}{"id": 153, "name": "hello"}

	runWithArgs("1 + num", args)
	runWithArgs("str ++ ` world`", args)
	runWithArgs("[i | i <- list, i != 2]", args)
	runWithArgs("obj.id", args)

	delete(args, "expr")
	runWithArgs("", args)
	args["expr"] = 357
	runWithArgs("", args)

	// Output:
	// 2
	// "hello world"
	// [1,3]
	// 153
	// missing expr
	// expr is not of type string
}

const xmlData = `
<?xml version="1.0" encoding="UTF-8"?>
<!-- comment -->
<name>xmlData</name>
<items xmlns:m="https://mingle.io">
    <m:item id="1">
        <name>Just character data</name>
    </m:item>
    <m:item id="2">
        <name>Second name</name>
    </m:item>
</items>`

func ExampleXML() {
	run(`xmlData.name`)
	run(`xmlData.name["text()"]`)
	run(`xmlData.items["@xmlns:m"]`)
	run(`[ a.name | a <- xmlData.items["m:item"]]`)
	run(`[ a.name["text()"] | a <- xmlData.items["m:item"]]`)
	run(`[ a["@id"] | a <- xmlData.items["m:item"]]`)

	// Output:
	// {"text()":"xmlData"}
	// "xmlData"
	// "https://mingle.io"
	// [{"text()":"Just character data"},{"text()":"Second name"}]
	// ["Just character data","Second name"]
	// [1,2]
}

func run(expr string) {
	req := fmt.Sprintf(`{"expr": %v}`, strconv.Quote(expr))
	_run(req)
}

func runWithArgs(expr string, args map[string]interface{}) {
	if expr != "" {
		args["expr"] = expr
	}

	req, err := json.Marshal(args)
	if err != nil {
		log.Fatalf("failed to marshal json: %v", err)
		return
	}

	_run(string(req))
}

func _run(req string) {
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

func addVars(store Store) {
	store.Add("xmlData.xml", strings.NewReader(xmlData))
}

func init() {
	go func() {
		log.SetOutput(ioutil.Discard)
		if err := Server(Port, "", runtime.NumCPU(), addVars); err != nil {
			log.Fatalf("failed to start comp: %v", err)
		}
	}()
}
