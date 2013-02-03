package main

import "testing"

func TestPos(t *testing.T) {
	mem := NewMem()

	aa := mem.PosPtr("a.a")
	ab1 := mem.PosPtr("a.b")
	ca := mem.PosPtr("c.a")
	ab2 := mem.PosPtr("a.b")

	head := make(Head)
	head["a"] = 0
	head["b"] = 1

	mem.Decl("a", head)
	if bad := mem.BadAttrs(); len(bad) != 1 || bad[0] != "c.a" {
		t.Errorf("invalid BadAttrs")
	}

	if *aa != 0 || *ab1 != 1 || *ab2 != 1 {
		t.Errorf("invalid positions")
	}

	if *ca != 0 {
		t.Errorf("c.a should be 0")
	}
}
