package main

import (
	"fmt"
	"regexp"
)

type Mem struct {
	Attrs []int // attribute positions within a tuple

	regexps  []*regexp.Regexp // regular expressions used by the query
	patterns []string         // patterns of the regular expressions
	heads    map[string]Head  // declared bindings, e.g. a <- geonames
	posIdx   map[string]int   // attribute positions (index in m.attrs)
	posOk    map[string]bool  // status of attribute declarationis
}

func NewMem() *Mem {
	return &Mem{heads: make(map[string]Head), posIdx: make(map[string]int), posOk: make(map[string]bool)}
}

// AttrPos returns the attribute position (an index in m.Attrs).
func (m *Mem) AttrPos(name string) int {
	val := m.posOk[name]
	m.posOk[name] = val

	return m.findAttr(name)
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

func (m *Mem) findAttr(name string) int {
	idx, ok := m.posIdx[name]
	if !ok {
		idx = len(m.Attrs)
		m.Attrs = append(m.Attrs, -1)
		m.posIdx[name] = idx
	}

	return idx
}

// Declare attributes. To get the list of undeclared attributes call m.BadAttrs().
func (m *Mem) Declare(name string, head Head) {
	for attr, idx := range head {
		ident := fmt.Sprintf("%v.%v", name, attr)
		pos := m.findAttr(ident)

		m.Attrs[pos] = idx
		m.posOk[ident] = true
	}

	pos := m.findAttr(name)
	m.Attrs[pos] = -1
	m.posOk[name] = true
	m.heads[name] = head
}

func (m *Mem) Head(name string) []string {
	head := m.heads[name]
	if head == nil {
		return nil
	}

	res := make([]string, len(head))
	for k, v := range head {
		res[v] = k
	}

	return res
}

// BadAttrs returns all attribute identifiers which are not known (e.g. were not m.Declare()).
func (m *Mem) BadAttrs() []string {
	var bad []string
	for ident, ok := range m.posOk {
		if !ok {
			bad = append(bad, ident)
		}
	}

	return bad
}

// Clone the memory to run in a separate goroutine.
func (m *Mem) Clone() *Mem {
	attrs := make([]int, len(m.Attrs))
	regexps := make([]*regexp.Regexp, len(m.patterns))
	patterns := make([]string, len(m.patterns))

	copy(attrs, m.Attrs)
	copy(patterns, m.patterns)

	for i, p := range patterns {
		regexps[i] = regexp.MustCompile(p)
	}

	return &Mem{Attrs: attrs, regexps: regexps, patterns: patterns}
}
