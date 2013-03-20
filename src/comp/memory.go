package main

import "regexp"

type Mem struct {
	vars    map[string]Value
	heads   map[string]Head
	fields  map[int64]Head
	regexps []*regexp.Regexp
}

func NewMem() *Mem {
	return &Mem{make(map[string]Value), make(map[string]Head), make(map[int64]Head), nil}
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

func (m *Mem) Decls() *Decls {
	decls := NewDecls()
	for name, _ := range m.vars {
		decls.Declare(name)
	}

	return decls
}

func (m *Mem) Bind(eid int64, name string, v Value) {
	if head := m.fields[eid]; head != nil {
		m.heads[name] = head
	}

	m.vars[name] = v
}

func (m *Mem) Store(name string, v Value, h Head) {
	m.vars[name] = v
	m.heads[name] = h
}

func (m *Mem) SetHead(eid int64, h Head) {
	m.fields[eid] = h
}

func (m *Mem) Load(eid int64, name string) Value {
	if head := m.fields[eid]; head == nil {
		if head = m.heads[name]; head != nil {
			m.fields[eid] = head
		}
	}

	return m.vars[name]
}

func (m *Mem) Field(eid int64, name string) int {
	if head := m.fields[eid]; head != nil {
		pos, ok := head[name]
		if ok {
			return pos
		}
	}

	return -1
}
