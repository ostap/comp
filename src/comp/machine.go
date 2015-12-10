// Copyright (c) 2013 Ostap Cherkashin, Julius Chrobak. You can use this source code
// under the terms of the MIT License found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
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
	opObject // allocate a new object on the stack with that many fields
	opSet    // set a field of an object to a value from the stack
	opGet    // get a field of an object and push it on the stack
	opIndex  // get an element of a list and push it on the stack
	opLoop   // prepare for iteration over a list from the stack
	opNext   // push the next element from the list on the stack and jump to op.Arg
	opTest   // jump to op.Arg if the top of the stack is false
	opMatch  // match a regular expression re with the top of the stack.
	opCall   // call a function taking arguments from the stack and pushing the result back
	opArg    // pass an integer value (op.Arg) to the next instruction (push)
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
	step int
	list List
}

func (p *Program) Run(s *Stack) Value {
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
		case opIndex:
			list := s.PopList()
			if op.Arg > -1 && op.Arg < len(list) {
				s.Push(list[op.Arg])
			} else {
				s.Push(String(""))
			}
		case opArg:
			s.Push(Number(op.Arg))
		case opLoop:
			parallel := s.PopBool()
			offset := int(s.PopNum())
			list := s.PopList()
			lid := op.Arg

			if len(list) > 0 {
				if parallel {
					cores := runtime.NumCPU()
					if cores > len(list) {
						cores = len(list)
					}

					ch := make(chan Value)
					for c := 0; c < cores; c++ {
						pc := p.Clone(i+1, i+1+offset)
						pc.loops[lid] = &iterator{cores + c, cores, list}

						sc := s.Clone()
						sc.Push(list[c])

						go func(_p *Program, _s *Stack) {
							ch <- _p.Run(_s)
						}(pc, sc)
					}

					var res List
					for c := 0; c < cores; c++ {
						part := <-ch
						if part == nil {
							continue
						}

						for _, v := range part.List() {
							res = append(res, v)
						}
					}

					s.PushList(res)

					i += offset + 1
					jump = true
				} else {
					p.loops[lid] = &iterator{1, 1, list}
					s.Push(list[0])
				}
			} else {
				i += offset
				jump = true
			}
		case opNext:
			offset := int(s.PopNum())
			loop := p.loops[op.Arg]
			if loop.pos > -1 && loop.pos < len(loop.list) {
				s.Push(loop.list[loop.pos])
				loop.pos += loop.step

				i += offset
				jump = true
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
			fn := p.funcs[op.Arg]
			fn.Eval(fn.State, s)
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

func (p *Program) log() {
	log.Printf("program %p", p)
	for x := 0; x < len(p.code); x++ {
		log.Printf("%2d %v", x, p.code[x])
	}
}

func (p *Program) Clone(from, to int) *Program {
	// TODO: deep copy
	res := new(Program)
	res.code = p.code[from:to]
	res.data = make([]Value, len(p.data))
	res.regexps = make([]*regexp.Regexp, len(p.regexps))
	res.funcs = make([]*Func, len(p.funcs))
	res.loops = make([]*iterator, len(p.loops))

	copy(res.data, p.data)
	for i, re := range p.regexps {
		res.regexps[i] = regexp.MustCompile(re.String())
	}
	copy(res.funcs, p.funcs)
	for i, fn := range p.funcs {
		res.funcs[i].State = fn.InitState()
	}

	return res
}

func (s *Stack) Clone() *Stack {
	res := new(Stack)
	for i := 0; i < s.top; i++ {
		// TODO: deep copy
		res.data[i] = s.data[i]
	}
	res.top = s.top

	return res
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
	case opIndex:
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
	case opArg:
		return fmt.Sprintf("arg %d", op.Arg)
	}

	return fmt.Sprintf("unknown op=%d arg=%d", op.Code, op.Arg)
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

func OpIndex(field int) Op {
	return Op{opIndex, field}
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

func OpArg(arg int) Op {
	return Op{opArg, arg}
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
