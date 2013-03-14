package main

type Decls struct {
	used    map[string]bool
	defined map[string]bool
}

func NewDecls() Decls {
	return Decls{make(map[string]bool), make(map[string]bool)}
}

func (d Decls) Use(name string) {
	d.used[name] = true
}

func (d Decls) Reset(name string) {
	delete(d.used, name)
}

func (d Decls) Declare(name string) {
	d.defined[name] = true
}

func (d Decls) Unknown() []string {
	var res []string
	for name, _ := range d.used {
		if !d.defined[name] {
			res = append(res, name)
		}
	}

	return res
}
