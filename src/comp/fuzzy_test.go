package main

import "testing"

func TestFuzzy(t *testing.T) {
	if d := dist("", ""); d != 0 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := dist("", "a"); d != 1 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := dist("a", ""); d != 1 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := dist("Hello World!", "Hello World!"); d != 0 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := dist("Hello", "hEELO"); d != 5 {
		t.Errorf("failed (dist == %d)", d)
	}

	if d := dist("Zürich", "Zurich"); d != 1 {
		t.Errorf("failed (dist == %d)", d)
	}

	if r := Fuzzy("", ""); r != 1 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := Fuzzy("", "a"); r != 0 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := Fuzzy("a", ""); r != 0 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := Fuzzy("Hello World!", "Hello World!"); r != 1 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := Fuzzy("Hello World!", "Hello World"); r == 0 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := Fuzzy("Hello World!", "Hello wORLD?"); r != 0.5 {
		t.Errorf("failed (fuzzy == %v)", r)
	}

	if r := Fuzzy("Zürich", "Zurich"); r != 0.8333333333333334 {
		t.Errorf("failed (fuzzy == %v)", r)
	}
}
