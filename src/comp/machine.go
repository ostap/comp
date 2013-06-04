package main

import (
	"fmt"
	"regexp"
)

// instructions
const (
	opList   int8 = iota // allocate a new list on the stack
	opAppend             // append a value from the stack to the list on the stack
	opNot                // logical not of the value from the stack
	opNeg
	opPos
	opMul
	opDiv
	opAdd
	opSub
	opCat
	opLT
	opLTE
	opGT
	opGTE
	opEq
	opNEq
	opAnd
	opOr
	opLoad   // load a value from address addr (push a value on the stack)
	opStore  // store a value from the top of the stack into a memory address
	opObject // Allocate a new object on the stack with that many fields.
	opSet    // Set a field of an object to a value from the stack.
	opGet    // Get a field of an object and push it on the stack.
	opLoop   /* Prepare for iteration over a list from the stack. Pushes the first element
	   from the list on the stack (if any) and pushes a boolean value indicating
	   whether the iteration is over. */
	opNext /* Push the next element from the list on the stack and push a boolean value
	   indicating success or failure (the same as in OpLoop). */
	opTest  // Jump if the top of the stack is false.
	opMatch // Match a regular expression re with the top of the stack.
	opCall  // Call a function. Takes arguments from the stack and puts a result back on the stack.
)

type Op struct {
	Code int8
	Arg  int
}

type Program struct {
	code    []Op
	data    []Value
	regexps []*regexp.Regexp
	funcs   []*Func
	loops   []*iterator
}

type Stack struct {
	data [4096]Value
	top  int
}

type iterator struct {
	pos  int
	list List
}

func (p *Program) Run() Value {
	s := new(Stack)
	i := 0
	for i > -1 && i < len(p.code) {
		op := p.code[i]
		jump := false

		switch op.Code {
		case opList:
			s.PushList(make(List, 0))
		case opAppend:
			val := s.Pop()
			list := s.PopList()
			list = append(list, val)
			s.PushList(list)
		case opNot:
			s.PushBool(!s.PopBool())
		case opNeg:
			s.PushNum(-s.PopNum())
		case opPos:
			s.PushNum(+s.PopNum())
		case opAnd:
			l := s.PopBool()
			r := s.PopBool()
			s.PushBool(l && r)
		case opOr:
			l := s.PopBool()
			r := s.PopBool()
			s.PushBool(l || r)
		// TODO: test LT, LTE, GT, GTE
		case opLT:
			l := s.PopNum()
			r := s.PopNum()
			s.PushBool(l < r)
		case opLTE:
			l := s.PopNum()
			r := s.PopNum()
			s.PushBool(l <= r)
		case opGT:
			l := s.PopNum()
			r := s.PopNum()
			s.PushBool(l > r)
		case opGTE:
			l := s.PopNum()
			r := s.PopNum()
			s.PushBool(l >= r)
		case opAdd:
			l := s.PopNum()
			r := s.PopNum()
			s.PushNum(l + r)
		case opSub:
			l := s.PopNum()
			r := s.PopNum()
			s.PushNum(l - r)
		case opMul:
			l := s.PopNum()
			r := s.PopNum()
			s.PushNum(l * r)
		case opDiv:
			l := s.PopNum()
			r := s.PopNum()
			s.PushNum(l / r)
		case opCat:
			l := s.PopStr()
			r := s.PopStr()
			s.PushStr(l + r)
		// TODO: test Eq, NEQ
		case opEq:
			l := s.Pop()
			r := s.Pop()
			s.Push(l.Equals(r))
		case opNEq:
			l := s.Pop()
			r := s.Pop()
			s.PushBool(!bool(l.Equals(r)))
		case opLoad:
			s.Push(p.data[op.Arg])
		case opStore:
			p.data[op.Arg] = s.Pop()
		case opObject:
			s.PushObj(make(Object, op.Arg))
		case opSet:
			val := s.Pop()
			obj := s.PopObj()
			obj[op.Arg] = val
			s.PushObj(obj)
		case opGet:
			obj := s.PopObj()
			val := obj[op.Arg]
			s.Push(val)
		case opLoop:
			list := s.PopList()
			if len(list) > 0 {
				p.loops[op.Arg] = &iterator{1, list}
				s.Push(list[0])
				s.Push(True)
			} else {
				s.Push(False)
			}
		case opNext:
			i := p.loops[op.Arg]
			if i.pos > -1 && i.pos < len(i.list) {
				s.Push(i.list[i.pos])
				s.Push(False)
				i.pos++
			} else {
				s.Push(True)
			}
		case opTest:
			if !s.PopBool() {
				i += op.Arg
				jump = true
			}
		case opMatch:
			str := s.PopStr()
			val := p.regexps[op.Arg].MatchString(str)
			s.PushBool(val)
		case opCall:
			p.funcs[op.Arg].Eval(s)
		default:
			msg := fmt.Sprintf("unknown operation %v", op)
			panic(msg)
		}

		if !jump {
			i++
		}
	}

	return s.Pop()
}

func (op Op) String() string {
	switch op.Code {
	case opList:
		return "list"
	case opAppend:
		return "append"
	case opNot:
		return "not"
	case opNeg:
		return "neg"
	case opPos:
		return "pos"
	case opMul:
		return "mul"
	case opDiv:
		return "div"
	case opAdd:
		return "add"
	case opSub:
		return "sub"
	case opCat:
		return "cat"
	case opLT:
		return "lt"
	case opLTE:
		return "lte"
	case opGT:
		return "gt"
	case opGTE:
		return "gte"
	case opEq:
		return "eq"
	case opNEq:
		return "neq"
	case opAnd:
		return "and"
	case opOr:
		return "or"
	case opLoad:
		return fmt.Sprintf("load %d", op.Arg)
	case opStore:
		return fmt.Sprintf("store %d", op.Arg)
	case opObject:
		return fmt.Sprintf("object %d", op.Arg)
	case opSet:
		return fmt.Sprintf("set %d", op.Arg)
	case opGet:
		return fmt.Sprintf("get %d", op.Arg)
	case opLoop:
		return fmt.Sprintf("loop %d", op.Arg)
	case opNext:
		return fmt.Sprintf("next %d", op.Arg)
	case opTest:
		return fmt.Sprintf("test %d", op.Arg)
	case opMatch:
		return fmt.Sprintf("match %d", op.Arg)
	case opCall:
		return fmt.Sprintf("call %d", op.Arg)
	}

	return fmt.Sprintf("unknown op=%x arg=%x (raw=%x)", op.Code, op.Arg, op)
}

func OpList() Op {
	return Op{opList, 0}
}

func OpAppend() Op {
	return Op{opAppend, 0}
}

func OpNot() Op {
	return Op{opNot, 0}
}

func OpNeg() Op {
	return Op{opNeg, 0}
}

func OpPos() Op {
	return Op{opPos, 0}
}

func OpMul() Op {
	return Op{opMul, 0}
}

func OpDiv() Op {
	return Op{opDiv, 0}
}

func OpAdd() Op {
	return Op{opAdd, 0}
}

func OpSub() Op {
	return Op{opSub, 0}
}

func OpCat() Op {
	return Op{opCat, 0}
}

func OpLT() Op {
	return Op{opLT, 0}
}

func OpLTE() Op {
	return Op{opLTE, 0}
}

func OpGT() Op {
	return Op{opGT, 0}
}

func OpGTE() Op {
	return Op{opGTE, 0}
}

func OpEq() Op {
	return Op{opEq, 0}
}

func OpNEq() Op {
	return Op{opNEq, 0}
}

func OpAnd() Op {
	return Op{opAnd, 0}
}

func OpOr() Op {
	return Op{opOr, 0}
}

func OpLoad(addr int) Op {
	return Op{opLoad, addr}
}

func OpStore(addr int) Op {
	return Op{opStore, addr}
}

func OpObject(fields int) Op {
	return Op{opObject, fields}
}

func OpSet(field int) Op {
	return Op{opSet, field}
}

func OpGet(field int) Op {
	return Op{opGet, field}
}

func OpLoop(lid int) Op {
	return Op{opLoop, lid}
}

func OpNext(lid int) Op {
	return Op{opNext, lid}
}

func OpTest(jump int) Op {
	return Op{opTest, jump}
}

func OpCall(fn int) Op {
	return Op{opCall, fn}
}

func OpMatch(re int) Op {
	return Op{opMatch, re}
}

func (s *Stack) Push(v Value) {
	s.data[s.top] = v
	s.top++
}

func (s *Stack) Pop() Value {
	s.top--
	return s.data[s.top]
}

func (s *Stack) PushBool(b bool) {
	s.data[s.top] = Bool(b)
	s.top++
}

func (s *Stack) PopBool() bool {
	s.top--
	return bool(s.data[s.top].Bool())
}

func (s *Stack) PushNum(n float64) {
	s.data[s.top] = Number(n)
	s.top++
}

func (s *Stack) PopNum() float64 {
	s.top--
	return float64(s.data[s.top].Number())
}

func (s *Stack) PushStr(str string) {
	s.data[s.top] = String(str)
	s.top++
}

func (s *Stack) PopStr() string {
	s.top--
	return string(s.data[s.top].String())
}

func (s *Stack) PushList(l List) {
	s.data[s.top] = l
	s.top++
}

func (s *Stack) PopList() List {
	s.top--
	v := s.data[s.top]
	if v == nil {
		return nil
	}
	return v.List()
}

func (s *Stack) PushObj(o Object) {
	s.data[s.top] = o
	s.top++
}

func (s *Stack) PopObj() Object {
	s.top--
	return s.data[s.top].Object()
}
