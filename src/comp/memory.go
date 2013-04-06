package main

import "regexp"

type Mem struct {
	Decls *Decls

	cells   [1024]Value
	regexps []*regexp.Regexp
	globals map[string]Type
}

func NewMem() *Mem {
	return &Mem{Decls: NewDecls()}
}

// RegExp returns the regular expression id (to be used in consequent m.MatchString() calls).
func (m *Mem) RegExp(pattern string) (int, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return -1, err
	}

	pos := len(m.regexps)
	m.regexps = append(m.regexps, re)

	return pos, nil
}

// MatchString checks if the regular expression re matches the string s.
func (m *Mem) MatchString(re int, s string) bool {
	return m.regexps[re].MatchString(s)
}

func (m *Mem) Global(name string, value Value, t Type) {
	addr := m.Decls.Declare(name, t)
	m.cells[addr] = value
}

func (m *Mem) Store(addr int, value Value) {
	m.cells[addr] = value
}

func (m *Mem) Load(addr int) Value {
	return m.cells[addr]
}
