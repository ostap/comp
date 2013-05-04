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
	inner   *Loop
	resAddr int
	varAddr int
	list    Expr
	sel     []Expr
	ret     Expr
}

func ForEach(resAddr, varAddr int, list Expr) *Loop {
	return &Loop{nil, resAddr, varAddr, list, nil, BadExpr}
}

func (l *Loop) Code() []Op {
	code := l.list.Code()

	// jump to the first instruction after the loop
	loopJump := l.codeLen(-1) + 1
	// jump back to the first instruction of the loop
	nextJump := -(loopJump - 2)

	code = append(code, OpLoop(loopJump))
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

	return append(code, OpNext(nextJump))
}

func (l *Loop) ResAddr() int {
	return l.innermost().resAddr
}

func (l *Loop) Nest(resAddr, varAddr int, list Expr) *Loop {
	l.innermost().inner = &Loop{nil, resAddr, varAddr, list, nil, BadExpr}
	return l
}

func (l *Loop) Select(expr Expr) *Loop {
	i := l.innermost()
	i.sel = append(l.sel, expr)
	return l
}

func (l *Loop) Return(expr Expr) *Loop {
	l.innermost().ret = expr
	return l
}

func (l *Loop) innermost() *Loop {
	i := l
	for i.inner != nil {
		i = i.inner
	}

	return i
}

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
