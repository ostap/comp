package main

import "testing"

func TestMemAttrPos(t *testing.T) {
	t.Errorf("fixme")

	/*
		mem := NewMem()

		aa := mem.AttrPos("a.a")
		ab1 := mem.AttrPos("a.b")
		ca := mem.AttrPos("c.a")
		ab2 := mem.AttrPos("a.b")

		head := make(Head)
		head["a"] = 0
		head["b"] = 1

		mem.Declare("a", head)
		if bad := mem.BadAttrs(); len(bad) != 1 || bad[0] != "c.a" {
			t.Errorf("invalid BadAttrs")
		}

		if mem.Attrs[aa] != 0 || mem.Attrs[ab1] != 1 || mem.Attrs[ab2] != 1 {
			t.Errorf("invalid positions")
		}

		if mem.Attrs[ca] != -1 {
			t.Errorf("c.a should be -1")
		}
	*/
}
