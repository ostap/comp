package main

type Comp struct {
	name  string
	trans func(t Tuple) Tuple
}

func Load(name string) *Comp {
	return &Comp{name: name, trans: func(t Tuple) Tuple { return t }}
}

func (c *Comp) Select(e Expr) *Comp {
	if c == nil {
		return nil
	}

	trans := c.trans
	c.trans = func(t Tuple) Tuple {
		if t != nil {
			if t = trans(t); t != nil && Bool(e(t)) {
				return t
			}
		}

		return nil
	}

	return c
}

func (c *Comp) Return(es []Expr) *Comp {
	if c == nil {
		return nil
	}

	trans := c.trans
	c.trans = func(t Tuple) Tuple {
		if t != nil {
			if t = trans(t); t != nil {
				tuple := make(Tuple, len(es))
				for i, e := range es {
					tuple[i] = Str(e(t))
				}
				return tuple
			}
		}

		return nil
	}

	return c
}

func (c *Comp) Run(v Views) Body {
	in := v.Body(c.name)
	out := make(Body)
	go func() {
		for _, t := range in {
			if t = c.trans(t); t != nil {
				out <- t
			}
		}

		close(out)
	}()

	return out
}
