package main

type Loop struct {
	inner *Loop
	addr  int
	list  Expr
	sel   []Expr
	ret   Expr
}

func ForEach(addr int, list Expr) *Loop {
	return &Loop{nil, addr, list, nil, BadExpr}
}

func (l *Loop) Eval(mem *Mem) List {
	res := make(List, 0)

	for _, v := range ToList(l.list, mem) {
		mem.Store(l.addr, v)

		ok := true
		for _, e := range l.sel {
			if ok = ToBool(e, mem); !ok {
				break
			}
		}

		if ok {
			if l.inner != nil {
				for _, i := range l.inner.Eval(mem) {
					res = append(res, i)
				}
			} else {
				res = append(res, l.ret.Eval(mem))
			}
		}
	}

	return res
}

func (l *Loop) Nest(addr int, list Expr) *Loop {
	l.innermost().inner = &Loop{nil, addr, list, nil, BadExpr}
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
