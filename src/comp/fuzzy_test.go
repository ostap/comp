// Copyright (c) 2013 Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import "testing"

func TestFuzzy(t *testing.T) {
	var f Fuzzy
	if d := f.dist("", ""); d != 0 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := f.dist("", "a"); d != 1 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := f.dist("a", ""); d != 1 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := f.dist("Hello World!", "Hello World!"); d != 0 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := f.dist("Hello", "hEELO"); d != 5 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := f.dist("Zürich", "Zurich"); d != 1 {
		t.Errorf("failed (dist == %d)", d)
	}

	if r := f.Compare("", ""); r != 1 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := f.Compare("", "a"); r != 0 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := f.Compare("a", ""); r != 0 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := f.Compare("Hello World!", "Hello World!"); r != 1 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := f.Compare("Hello World!", "Hello World"); r == 0 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := f.Compare("Hello World!", "Hello wORLD?"); r != 0.5 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := f.Compare("Zürich", "Zurich"); r != 0.8333333333333334 {
		t.Errorf("failed (fuzzy == %v)", r)
	}
}

func BenchmarkFuzzyBasic(b *testing.B) {
	var f Fuzzy
	for i := 0; i < b.N; i++ {
		f.Compare("Hello World!", "Hello wORLD?")
	}
}
