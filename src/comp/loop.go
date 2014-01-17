// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

/*
The following comprehension:
  [i * j | i <- [1, 2, 3], j <- [10, 20], i == j / 10]

Will be compiled as follows:
  // list [1, 2, 3]
  loop +L1
  // list [10, 20]
  loop +L2
  // i == j / 10
  test -T1 // back to L2
  // R: append(res, i * j)
  next -N1 // back to L2
  next -N2 // back to L1

Where L1, L2, T1, N1, N2 are relative code offsets (jumps):
  +-------->----------->----------+
  |     +-->----------->-------+  |
  |     |     +-------->-+     |  |
  L1 -> L2 -> T1 -> R -> N2 -> N1 v
  |      |               |     |
  |      +-<----------->-+     |
  +--------<----------->-------+
*/
type Loop struct {
	lid      int
	inner    *Loop
	resAddr  int
	varAddr  int
	iterAddr int
	list     Expr
	sel      []Expr
	ret      Expr
	parallel bool
}

func ForEach(lid int, varAddr, iterAddr int, list Expr, parallel bool) *Loop {
	return &Loop{lid, nil, -1, varAddr, iterAddr, list, nil, BadExpr, parallel}
}

func (l *Loop) Code() []Op {
	code := l.list.Code()
	clen := l.codeLen(-1)

	// jump over the loop
	loopJump := clen + 1 /* OpNext */ + 1 /* next instruction after the loop */
	// jump back to beginning
	nextJump := -clen

	parallel := 0
	if l.parallel {
		parallel = 1
	}

	code = append(code, OpArg(loopJump))
	code = append(code, OpArg(parallel))
	code = append(code, OpLoop(l.lid))
	code = append(code, OpStore(l.iterAddr))
	code = append(code, OpStore(l.varAddr))

	for i, s := range l.sel { // select(s)
		for _, c := range s.Code() {
			code = append(code, c)
		}
		code = append(code, OpTest(l.codeLen(i)))
	}

	if l.inner != nil { // nested loop(s)
		for _, c := range l.inner.Code() {
			code = append(code, c)
		}
	} else { // return
		code = append(code, OpLoad(l.resAddr))
		for _, c := range l.ret.Code() {
			code = append(code, c)
		}
		code = append(code, OpAppend())
		code = append(code, OpStore(l.resAddr))
	}

	code = append(code, OpArg(nextJump))
	return append(code, OpNext(l.lid))
}

func (l *Loop) Nest(lid int, varAddr, iterAddr int, list Expr, parallel bool) *Loop {
	l.innermost().inner = &Loop{lid, nil, -1, varAddr, iterAddr, list, nil, BadExpr, parallel}
	return l
}

func (l *Loop) Select(expr Expr) *Loop {
	i := l.innermost()
	i.sel = append(l.sel, expr)
	return l
}

func (l *Loop) Return(expr Expr, resAddr int) *Loop {
	l.innermost().ret = expr
	l.innermost().resAddr = resAddr
	return l
}

func (l *Loop) innermost() *Loop {
	i := l
	for i.inner != nil {
		i = i.inner
	}

	return i
}

// codeLen calculates the jump from a test instruction. selPos is the index
// of a select statement (left to right). passing -1 will calculate the length
// of the whole loop.
func (l *Loop) codeLen(selPos int) int {
	jump := 0
	if selPos < 0 {
		jump += 2 /* OpStore - iterAddr, varAddr */
	}

	for i := selPos + 1; i < len(l.sel); i++ {
		jump += len(l.sel[i].Code()) + 1 /* OpTest */
	}

	if l.inner != nil {
		jump += len(l.inner.Code())
	} else {
		jump += 1 /* OpLoad */ + len(l.ret.Code()) + 1 /* OpAppend */ + 1 /* OpStore */
	}

	return jump + 1 /* OpArg */
}
