// Copyright (c) 2013 Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"testing"
)

func _readJSON(jsonBlob string) (Type, Value, error) {
	return readJSON(bufio.NewReader(bytes.NewReader([]byte(jsonBlob))))
}

func _readXML(xmlBlob string) (Type, Value, error) {
	return readXML(bufio.NewReader(bytes.NewReader([]byte(xmlBlob))))
}

func ok(t *testing.T, blobType, blob string) {
	var rt Type
	var rv Value
	var err error

	if blobType == "xml" {
		rt, rv, err = _readXML(blob)
	} else if blobType == "json" {
		rt, rv, err = _readJSON(blob)
	}
	if err != nil || rt == nil || rv == nil {
		t.Log("error:", err)
		t.FailNow()
	}
}

func err(t *testing.T, blobType, blob string) {
	var rt Type
	var rv Value
	var err error

	if blobType == "xml" {
		rt, rv, err = _readXML(blob)
	} else if blobType == "json" {
		rt, rv, err = _readJSON(blob)
	}
	if err == nil || rt != nil || rv != nil {
		t.Log("error:", err)
		t.FailNow()
	}
}

func TestJSONBasic(t *testing.T) {
	ok(t, "json", `[1,2,3,4]`)

	ok(t, "json", `{"Name": "Platypus"}`)

	ok(t, "json", `
        {"Name": "Platypus"}, {"Name": "Quoll"}
    ]`)

	ok(t, "json", `[
        {"Name": "Platypus"}, {"Name": 1}
    ]`)

	ok(t, "json", `[
        {"Name": "Platypus"}, {"Name": true}
    ]`)

	ok(t, "json", `
        [1,"hello"]
    `)

	err(t, "json", `
        [{},"hello"]
    `)

	err(t, "json", `[
        {"Name": "Platypus"}, {"Name": []}
    ]`)

	err(t, "json", `[
        {"Name": "Platypus"}, {"Name": {}}
    ]`)

	err(t, "json", `[
        {"Name": "Platypus"}, {"Id": "Quoll"}
    ]`)

	err(t, "json", `[
        {"Name": "Platypus"}, {"name": "Quoll"}
    ]`)
}

func TestJSONNested(t *testing.T) {
	ok(t, "json", `
        {"Order": [1,2,3,4]}
    `)

	ok(t, "json", `
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]}
    `)

	ok(t, "json", `[
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [{"Id": 1}]}
    ]`)

	ok(t, "json", `[
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [{"Id": "hello"}]}
    ]`)

	err(t, "json", `[
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [1, 2, 3]}
    ]`)

	err(t, "json", `[
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [[]]}
    ]`)

	err(t, "json", `[
        {"Order": [{"Id": 1}, {"Id": 2}, {"Id": 3}]},
        {"Order": [{}]}
    ]`)
}

func TestXML(t *testing.T) {
	ok(t, "xml", `Just Character Data`)

	ok(t, "xml", `0.123456`)

	ok(t, "xml", `<?xml version="1.0" encoding="UTF-8"?>`)

	ok(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        <!-- comment -->
        <item></item>
    `)

	ok(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        <!-- comment -->
        <item>
            <id>1</id>
        </item>
    `)

	ok(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        <!-- comment -->
        <item a="attribute" n="0.123456789">
            <id>1</id>
        </item>
    `)

	ok(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        <!-- comment -->
        <item a="attribute" n="0.123456789">
            <id>1</id>
            Just character data
        </item>
    `)

	ok(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        <!-- comment -->
        <item a="attribute" n="0.123456789">
            <id>1</id>
            <id>2</id>
            Just character data
        </item>
        <item a="attribute" n="0.123456789">
            <id>3</id>
            <id>4</id>
            Second character data
        </item>
    `)

	ok(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        <!-- comment -->
        <global:item a="attribute" n="0.123456789">
            <id>1</id>
            Just character data
        </global:item>
        <local:item>
            <name>Some Name</name>
        </local:item>
    `)

	err(t, "xml", `<?xml version="1.0" encoding="ISO-8859-2"?>`)

	err(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        </item>
    `)

	err(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        <item>
    `)

	err(t, "xml", `
        <?xml version="1.0" encoding="UTF-8"?>
        <!-- comment -->
        <item a="attribute">
            <id>1</id>
            Just character data
        </item>
        <item a="attribute">
            <name>Some Name</name>
        </item>
    `)
}
