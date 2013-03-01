package main

type Comp func(m *Mem, t Tuple) Tuple

func Reflect(m *Mem, t Tuple) Tuple {
	return t
}

func Select(c Comp, e Expr) Comp {
	return func(m *Mem, t Tuple) Tuple {
		if t != nil {
			if t = c(m, t); t != nil && BoolEval(e, m, t) {
				return t
			}
		}

		return nil
	}
}

func Return(c Comp, es []Expr) Comp {
	return func(m *Mem, t Tuple) Tuple {
		if t != nil {
			if t = c(m, t); t != nil {
				tuple := make(Tuple, len(es))
				for i, e := range es {
					tuple[i] = e.Eval(m, t)
				}
				return tuple
			}
		}

		return nil
	}
}
