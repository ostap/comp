package main

type Decls struct {
	used    map[string]bool
	defined map[string]bool
	ordered []string
}

func NewDecls() *Decls {
	return &Decls{make(map[string]bool), make(map[string]bool), nil}
}

func (d *Decls) Use(name string) {
	if !d.used[name] {
		d.used[name] = true
		d.ordered = append(d.ordered, name)
	}
}

func (d *Decls) Reset(name string) {
	ordered := make([]string, 0, len(d.ordered))
	for _, n := range d.ordered {
		if n != name {
			ordered = append(ordered, n)
		}
	}

	d.ordered = ordered
	delete(d.used, name)
}

func (d *Decls) Declare(name string) {
	d.defined[name] = true
}

func (d *Decls) Unknown() []string {
	var res []string
	for _, name := range d.ordered {
		if !d.defined[name] {
			res = append(res, name)
		}
	}

	return res
}
