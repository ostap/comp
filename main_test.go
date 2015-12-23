// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
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

	// FIXME: the following two tests should both return false (see #26)
	// run("0 == ``")
	// run("`` == 0")

	// Output:
	// true
	// false
	// true
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
	// {"id":1,"children":[2,3]}
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
	run(`[{g,c}|g <- [1], c <- [0], c-1 == 0 && c == 0]`)
	run(`[{g,c}|g <- [1], c <- [0], c-1 == 0, c == 0]`)

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
	//
	//
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

func ExampleJSON() {
	json := `
		{
			"num": 1,
			"str": "hello",
			"list": [1, 2, 3],
			"obj": {"id": 153, "name": "hello"}
		}`

	runWithInputs("1 + in.num", "in.json", json)
	runWithInputs("in.str ++ ` world`", "in.json", json)
	runWithInputs("[i | i <- in.list, i != 2]", "in.json", json)
	runWithInputs("in.obj.id", "in.json", json)

	// Output:
	// 2
	// "hello world"
	// [1,3]
	// 153
}

func ExampleXML() {
	const xml = `
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

	runWithInputs(`xmlData.name`, "xmlData.xml", xml)
	runWithInputs(`xmlData.name["text()"]`, "xmlData.xml", xml)
	runWithInputs(`xmlData.items["@xmlns:m"]`, "xmlData.xml", xml)
	runWithInputs(`[ a.name | a <- xmlData.items["m:item"]]`, "xmlData.xml", xml)
	runWithInputs(`[ a.name["text()"] | a <- xmlData.items["m:item"]]`, "xmlData.xml", xml)
	runWithInputs(`[ a["@id"] | a <- xmlData.items["m:item"]]`, "xmlData.xml", xml)

	// Output:
	// {"text()":"xmlData"}
	// "xmlData"
	// "https://mingle.io"
	// [{"text()":"Just character data"},{"text()":"Second name"}]
	// ["Just character data","Second name"]
	// [1,2]
}

func _run(expr string, inputs map[string]io.Reader) {
	buf := new(bytes.Buffer)
	if err := Run(expr, inputs, buf); err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Printf("%v", buf.String())
	}
}

func runWithInputs(expr, file, data string) {
	inputs := make(map[string]io.Reader)
	inputs[file] = strings.NewReader(data)

	_run(expr, inputs)
}

func run(expr string) {
	_run(expr, nil)
}
