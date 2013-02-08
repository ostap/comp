package main

type Comp func(t Tuple) Tuple

func Reflect(t Tuple) Tuple {
	return t
}

func Select(c Comp, e Expr) Comp {
	return func(t Tuple) Tuple {
		if t != nil {
			if t = c(t); t != nil && Bool(e(t)) {
				return t
			}
		}

		return nil
	}
}

func Return(c Comp, es []Expr) Comp {
	return func(t Tuple) Tuple {
		if t != nil {
			if t = c(t); t != nil {
				tuple := make(Tuple, len(es))
				for i, e := range es {
					tuple[i] = Str(e(t))
				}
				return tuple
			}
		}

		return nil
	}
}
