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
	lid     int
	inner   *Loop
	resAddr int
	varAddr int
	list    Expr
	sel     []Expr
	ret     Expr
}

func ForEach(lid int, varAddr int, list Expr) *Loop {
	return &Loop{lid, nil, -1, varAddr, list, nil, BadExpr}
}

func (l *Loop) Code() []Op {
	code := l.list.Code()
	clen := l.codeLen(-1)

	// jump to the first instruction _after_ the loop
	loopJump := clen + 1 /* OpTest */ + 1
	// jump back to the first instruction of the loop
	nextJump := -clen

	code = append(code, OpLoop(l.lid))
	code = append(code, OpTest(loopJump))
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
		code = append(code, OpAppend)
		code = append(code, OpStore(l.resAddr))
	}

	code = append(code, OpNext(l.lid))
	return append(code, OpTest(nextJump))
}

func (l *Loop) Nest(lid int, varAddr int, list Expr) *Loop {
	l.innermost().inner = &Loop{lid, nil, -1, varAddr, list, nil, BadExpr}
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

// codeLen calculates the length of the code after a test instruction. selPos
// is the index of the select statement (left to right). passing -1 will
// produce the length of the whole loop minus one (trailing OpTest).
func (l *Loop) codeLen(selPos int) int {
	jump := 0
	if selPos < 0 {
		jump++ /* OpStore */
	}

	for i := selPos + 1; i < len(l.sel); i++ {
		jump += len(l.sel[i].Code()) + 1 /* OpTest */
	}

	if l.inner != nil {
		jump += len(l.inner.Code())
	} else {
		jump += 1 /* OpLoad */ + len(l.ret.Code()) + 1 /* OpAppend */ + 1 /* OpStore */
	}

	return jump + 1 /* OpNext */
}
