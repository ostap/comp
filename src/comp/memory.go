package main

import (
	"fmt"
)

type Mem struct {
	data   map[string]*Value
	posIdx map[string]int  // attribute positions (index in posVal)
	posVal [1024]int       // the actual attribute positions
	posOk  map[string]bool // status of attribute declarations
}

func NewMem() *Mem {
	return &Mem{
		data:   make(map[string]*Value),
		posIdx: make(map[string]int),
		posOk:  make(map[string]bool)}
}

// Returns a pointer to the attribute position within a tuple. The pointer value
// is only valid after the corresponding Decl() call.
func (m *Mem) PosPtr(name string) *int {
	val := m.posOk[name]
	m.posOk[name] = val

	idx, ok := m.posIdx[name]
	if !ok {
		idx = len(m.posIdx)
		m.posIdx[name] = idx
	}

	if idx >= len(m.posVal) {
		return nil
	}

	return &m.posVal[idx]
}

// Declare a collection. To get an list of bad attribute identifiers call BadAttrs().
func (m *Mem) Decl(name string, head Head) {
	for attr, idx := range head {
		ident := fmt.Sprintf("%v.%v", name, attr)
		posIdx, ok := m.posIdx[ident]
		if !ok {
			posIdx = len(m.posIdx)
			m.posIdx[ident] = posIdx
		}

		m.posVal[posIdx] = idx
		m.posOk[ident] = true
	}
}

// Returns a list of invalid identifiers.
func (m *Mem) BadAttrs() []string {
	var bad []string
	for ident, ok := range m.posOk {
		if !ok {
			bad = append(bad, ident)
		}
	}

	return bad
}

func (m *Mem) Load(name string) *Value {
	return nil
}

func (m Mem) Store(name string, value *Value) {
}
