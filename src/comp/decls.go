package main

type Decls struct {
	used    map[string]bool
	defined map[string]bool
	ordered []string
	unknown []string
	strict  bool
}

func NewDecls() *Decls {
	return &Decls{make(map[string]bool), make(map[string]bool), nil, nil, false}
}

func (d *Decls) Strict(on bool) {
	d.strict = on
}

func (d *Decls) Use(name string) {
	if !d.used[name] {
		d.used[name] = true
		d.ordered = append(d.ordered, name)
	}

	if d.strict && !d.defined[name] {
		d.unknown = append(d.unknown, name)
	}
}

func (d *Decls) Reset(name string) {
	ordered := make([]string, 0, len(d.ordered))
	for _, n := range d.ordered {
		if n != name {
			ordered = append(ordered, n)
		}
	}
	unknown := make([]string, 0, len(d.unknown))
	for _, n := range d.unknown {
		if n != name {
			unknown = append(unknown, n)
		}
	}

	d.ordered = ordered
	d.unknown = unknown
	delete(d.used, name)
}

func (d *Decls) Declare(name string) {
	d.defined[name] = true
}

func (d *Decls) Unknown() []string {
	res := make([]string, len(d.unknown))
	copy(res, d.unknown)
	for _, name := range d.ordered {
		if !d.defined[name] {
			res = append(res, name)
		}
	}

	return res
}
