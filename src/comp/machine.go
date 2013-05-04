package main

import (
	"fmt"
)

type Op int64

func (op Op) Code() Op {
	return op & opMASK
}

func (op Op) Arg() int {
	return int(op & ^opMASK)
}

func (op Op) String() string {
	switch op.Code() {
	case OpList:
		return "list"
	case OpAppend:
		return "append"
	case OpNot:
		return "not"
	case OpNeg:
		return "neg"
	case OpPos:
		return "pos"
	case OpMul:
		return "mul"
	case OpDiv:
		return "div"
	case OpAdd:
		return "add"
	case OpSub:
		return "sub"
	case OpCat:
		return "cat"
	case OpLT:
		return "lt"
	case OpLTE:
		return "lte"
	case OpGT:
		return "gt"
	case OpGTE:
		return "gte"
	case OpEq:
		return "eq"
	case OpNEq:
		return "neq"
	case OpAnd:
		return "and"
	case OpOr:
		return "or"

	case opLoad:
		return fmt.Sprintf("load %d", op.Arg())
	case opStore:
		return fmt.Sprintf("store %d", op.Arg())
	case opObject:
		return fmt.Sprintf("object %d", op.Arg())
	case opSet:
		return fmt.Sprintf("set %d", op.Arg())
	case opGet:
		return fmt.Sprintf("get %d", op.Arg())
	case opLoop:
		return fmt.Sprintf("loop %d", op.Arg())
	case opNext:
		return fmt.Sprintf("next %d", op.Arg())
	case opTest:
		return fmt.Sprintf("test %d", op.Arg())
	case opMatch:
		return fmt.Sprintf("match %d", op.Arg())
	case opCall:
		return fmt.Sprintf("call %d", op.Arg())
	}

	return fmt.Sprintf("UNKNOWN %x", op)
}

// Load a value from address addr (push a value on the stack).
func OpLoad(addr int) Op {
	return opLoad | Op(uint32(addr))
}

// Store the top of the stack into address addr.
func OpStore(addr int) Op {
	return opStore | Op(uint32(addr))
}

// Allocate a new object on the stack with that many fields.
func OpObject(fields int) Op {
	return opObject | Op(uint32(fields))
}

// Set a field of an object to a value from the stack.
func OpSet(field int) Op {
	return opSet | Op(uint32(field))
}

// Get a field of an object and push it on the stack.
func OpGet(field int) Op {
	return opGet | Op(uint32(field))
}

// Prepare for an iteration over a list from the stack.
// Puts the first element from the list on the stack.
func OpLoop(jump int) Op {
	return opLoop | Op(uint32(jump))
}

// Put the next element from a list (see OpLoop) on the stack and continue
// with the iteration (jump to start).
func OpNext(jump int) Op {
	return opNext | Op(uint32(jump))
}

// Jump if the top of the stack is false.
func OpTest(jump int) Op {
	return opTest | Op(uint32(jump))
}

// Call a function. Takes arguments from the stack and puts a result back on the stack.
func OpCall(fn int) Op {
	return opCall | Op(uint32(fn))
}

// Match a regular expression re with the top of the stack.
func OpMatch(re int) Op {
	return opMatch | Op(uint32(re))
}

const (
	opMASK = 0x7FFF000000000000

	OpList   Op = iota << 48 // Allocate a new list on the stack.
	OpAppend                 // Append a value from the stack to the list on the stack.
	OpNot
	OpNeg
	OpPos
	OpMul
	OpDiv
	OpAdd
	OpSub
	OpCat
	OpLT
	OpLTE
	OpGT
	OpGTE
	OpEq
	OpNEq
	OpAnd
	OpOr

	opLoad
	opStore
	opObject
	opSet
	opGet
	opLoop
	opNext
	opTest
	opMatch
	opCall
)

type Stack struct {
	data [4096]Value
	top  int
}

func (s *Stack) Pop() Value {
	s.top--
	return s.data[s.top]
}

func (s *Stack) PopBool() bool {
	s.top--
	return bool(s.data[s.top].Bool())
}

func (s *Stack) PopNum() float64 {
	s.top--
	return float64(s.data[s.top].Number())
}

func (s *Stack) PopList() List {
	s.top--
	v := s.data[s.top]
	if v == nil {
		return nil
	}
	return v.List()
}

func (s *Stack) PopStr() string {
	s.top--
	return string(s.data[s.top].String())
}

func (s *Stack) PopObject() Object {
	s.top--
	return s.data[s.top].Object()
}

func (s *Stack) Push(v Value) {
	s.data[s.top] = v
	s.top++
}

func (s *Stack) PushList(l List) {
	s.data[s.top] = l
	s.top++
}

func (s *Stack) PushNum(n float64) {
	s.data[s.top] = Number(n)
	s.top++
}

func (s *Stack) PushBool(b bool) {
	s.data[s.top] = Bool(b)
	s.top++
}

func (s *Stack) PushStr(str string) {
	s.data[s.top] = String(str)
	s.top++
}

type Program struct {
	code []Op
	data []Value
	// funcs []Func
}

func (p *Program) Run() Value {
	s := new(Stack)
	i := 0
	for i > -1 && i < len(p.code) {
		op := p.code[i]
		jump := false

		switch op.Code() {
		case opLoad:
			s.Push(p.data[op.Arg()])
		case opStore:
			p.data[op.Arg()] = s.Pop()
		case OpList:
			s.Push(make(List, 0))
		case OpAppend:
			val := s.Pop()
			list := s.PopList()
			list = append(list, val)
			s.Push(list)
		case opObject:
			s.Push(make(Object, op.Arg()))
		case opSet:
			val := s.Pop()
			obj := s.PopObject()
			obj[op.Arg()] = val
			s.Push(obj)
		case opGet:
			obj := s.PopObject()
			s.Push(obj[op.Arg()])
		case opLoop:
			list := s.PopList()
			if len(list) > 0 {
				s.Push(list)
				s.PushNum(1)
				s.Push(list[0])
			} else {
				i += op.Arg()
				jump = true
			}
		case opNext:
			idx := s.PopNum()
			list := s.PopList()
			if int(idx) > -1 && int(idx) < len(list) {
				s.PushList(list)
				s.PushNum(idx + 1)
				s.Push(list[int(idx)])

				i += op.Arg()
				jump = true
			}
		case opTest:
			if !s.PopBool() {
				i += op.Arg()
				jump = true
			}
		case OpNot:
			s.PushBool(!s.PopBool())
		case OpNeg:
			s.PushNum(-s.PopNum())
		case OpPos:
			s.PushNum(+s.PopNum())
		case OpAnd:
			l := s.PopBool()
			r := s.PopBool()
			s.PushBool(l && r)
		case OpOr:
			l := s.PopBool()
			r := s.PopBool()
			s.PushBool(l || r)
		// TODO: test LT, LTE, GT, GTE
		case OpLT:
			l := s.PopNum()
			r := s.PopNum()
			s.PushBool(l < r)
		case OpLTE:
			l := s.PopNum()
			r := s.PopNum()
			s.PushBool(l <= r)
		case OpGT:
			l := s.PopNum()
			r := s.PopNum()
			s.PushBool(l > r)
		case OpGTE:
			l := s.PopNum()
			r := s.PopNum()
			s.PushBool(l >= r)
		case OpAdd:
			l := s.PopNum()
			r := s.PopNum()
			s.PushNum(l + r)
		case OpSub:
			l := s.PopNum()
			r := s.PopNum()
			s.PushNum(l - r)
		case OpMul:
			l := s.PopNum()
			r := s.PopNum()
			s.PushNum(l * r)
		case OpDiv:
			l := s.PopNum()
			r := s.PopNum()
			s.PushNum(l / r)
		case OpCat:
			l := s.PopStr()
			r := s.PopStr()
			s.PushStr(l + r)
		// TODO: test Eq, NEQ
		case OpEq:
			l := s.Pop()
			r := s.Pop()
			s.Push(l.Equals(r))
		case OpNEq:
			l := s.Pop()
			r := s.Pop()
			s.PushBool(!bool(l.Equals(r)))
		default:
			msg := fmt.Sprintf("unknown operation %x", op)
			panic(msg)
		}

		if !jump {
			i++
		}
	}

	return s.Pop()
}
