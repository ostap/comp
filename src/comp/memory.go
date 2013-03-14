package main

import (
	"regexp"
)

type Mem struct {
	vars     map[string]Value
	regexps  []*regexp.Regexp // regular expressions used by the query
	patterns []string         // patterns of the regular expressions

}

func NewMem() *Mem {
	return &Mem{vars: make(map[string]Value)}
}

// RegExp returns the regular expression id (to be used in consequent m.MatchString() calls).
func (m *Mem) RegExp(pattern string) (int, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return -1, err
	}

	pos := len(m.regexps)
	m.regexps = append(m.regexps, re)
	m.patterns = append(m.patterns, pattern)

	return pos, nil
}

// MatchString checks if the regular expression re matches the string s.
func (m *Mem) MatchString(re int, s string) bool {
	return m.regexps[re].MatchString(s)
}

func (m *Mem) Decls() Decls {
	decls := NewDecls()
	for name, _ := range m.vars {
		decls.Declare(name)
	}

	return decls
}

func (m *Mem) Store(name string, v Value) {
	m.vars[name] = v
}

func (m *Mem) Load(name string) Value {
	return m.vars[name]
}
