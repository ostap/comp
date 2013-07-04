// Copyright (c) 2013 Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"encoding/json"
	"testing"
)

func _traverse(t *testing.T, jsonBlob []byte) (Type, Value, error) {
	var data interface{}
	err := json.Unmarshal(jsonBlob, &data)
	if err != nil {
		t.Log("error:", err)
		t.FailNow()
	}
	return traverse(nil, data)
}

func ok(t *testing.T, jsonBlob []byte) {
	rt, rv, err := _traverse(t, jsonBlob)
	if err != nil || rt == nil || rv == nil {
		t.Log("error:", err)
		t.FailNow()
	}
}

func err(t *testing.T, jsonBlob []byte) {
	rt, rv, err := _traverse(t, jsonBlob)
	if err == nil || rt != nil || rv != nil {
		t.Log("error:", err)
		t.FailNow()
	}
}

func TestJSONBasic(t *testing.T) {
	ok(t, []byte(`
        [1,2,3,4]
    `))

	ok(t, []byte(`
        {"Name": "Platypus"}
    `))

	ok(t, []byte(`[
        {"Name": "Platypus"}, {"Name": "Quoll"}
    ]`))

	ok(t, []byte(`[
        {"Name": "Platypus"}, {"Name": 1}
    ]`))

	ok(t, []byte(`[
        {"Name": "Platypus"}, {"Name": true}
    ]`))

	ok(t, []byte(`
        [1,"hello"]
    `))

	err(t, []byte(`
        [{},"hello"]
    `))

	err(t, []byte(`[
        {"Name": "Platypus"}, {"Name": []}
    ]`))

	err(t, []byte(`[
        {"Name": "Platypus"}, {"Name": {}}
    ]`))

	err(t, []byte(`[
        {"Name": "Platypus"}, {"Id": "Quoll"}
    ]`))

	err(t, []byte(`[
        {"Name": "Platypus"}, {"name": "Quoll"}
    ]`))
}

func TestJSONNested(t *testing.T) {
	ok(t, []byte(`
        {"Order": [1,2,3,4]}
    `))

	ok(t, []byte(`
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]}
    `))

	ok(t, []byte(` [
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [{"Id": 1}]}
    ]`))

	ok(t, []byte(` [
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [{"Id": "hello"}]}
    ]`))

	err(t, []byte(` [
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [1, 2, 3]}
    ]`))

	err(t, []byte(` [
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [[]]}
    ]`))

	err(t, []byte(` [
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [{}]}
    ]`))
}
