package interp

import (
	"fmt"

	"github.com/edorfaus/whitespace/ast"
)

type VM struct {
	WriteChar   func(rune) error
	WriteNumber func(int64) error
	ReadChar    func() (rune, error)
	ReadNumber  func() (int64, error)

	Code  []Instr
	Stack []int64
	Heap  []int64
	RetTo []int
	PC    int
	Err   error
}

func NewVM(code []ast.Command) *VM {
	vm := &VM{
		WriteChar:   DefaultWriteChar,
		WriteNumber: DefaultWriteNumber,
		ReadChar:    DefaultReadChar,
		ReadNumber:  DefaultReadNumber,
	}
	vm.translate(code)
	return vm
}

func (vm *VM) Run() {
	if vm.Err != nil {
		return
	}
	for vm.Err == nil {
		i := vm.Code[vm.PC]
		vm.PC++
		i.Op(vm, i.Arg)
	}
	if vm.Err == errProgramExit {
		vm.Err = nil
	}
}

func (vm *VM) fail(format string, args ...interface{}) {
	if vm.Err == nil {
		vm.Err = fmt.Errorf(format, args...)
	}
}

func (vm *VM) translate(code []ast.Command) {
	labels := map[string]int64{}
	update := map[int]string{}
	maxLabel := -1
	var out []Instr
	for i, from := range code {
		if from.Cmd <= ast.CmdNone || from.Cmd >= ast.CountCmds {
			vm.fail("index %v: invalid command: %v", i, from.Cmd)
			return
		}
		inst := Instr{
			Op: vmOps[from.Cmd],
		}
		switch from.Cmd {
		case ast.CmdPush, ast.CmdCopy, ast.CmdSlide:
			v, ok := from.Arg.(int64)
			if !ok {
				vm.fail(
					"index %v: expected int64 argument, got %T",
					i, from.Arg,
				)
				return
			}
			inst.Arg = v
		case ast.CmdMark:
			v, ok := from.Arg.(string)
			if !ok {
				vm.fail(
					"index %v: expected string argument, got %T",
					i, from.Arg,
				)
				return
			}
			if pos, ok := labels[v]; ok {
				vm.fail(
					"index %v: duplicate label (from index %v): %q",
					i, pos, v,
				)
				return
			}
			labels[v] = int64(len(out))
			maxLabel = len(out)
			// Do not add the label definition to the output code
			continue
		case ast.CmdCall, ast.CmdJump, ast.CmdJumpIfZero, ast.CmdJumpIfNeg:
			v, ok := from.Arg.(string)
			if !ok {
				vm.fail(
					"index %v: expected string argument, got %T",
					i, from.Arg,
				)
				return
			}
			if pos, ok := labels[v]; ok {
				inst.Arg = pos
			} else {
				update[len(out)] = v
			}
		}

		out = append(out, inst)
	}

	if maxLabel >= len(out) {
		vm.fail("label points past end of code")
		return
	}

	for i, lbl := range update {
		if pos, ok := labels[lbl]; ok {
			out[i].Arg = pos
		} else {
			vm.fail("index %v: undefined label: %q", i, lbl)
			return
		}
	}

	vm.Code = out
}

func (vm *VM) stackSize(n int) bool {
	if len(vm.Stack) < n {
		vm.fail("stack underflow")
		return false
	}
	return vm.Err == nil
}

func (vm *VM) pop() int64 {
	p := len(vm.Stack) - 1
	v := vm.Stack[p]
	vm.Stack = vm.Stack[:p]
	return v
}

func (vm *VM) storeHeap(adr, val int64) {
	switch {
	case adr < 0:
		vm.fail("store to negative heap address: %v = %v", adr, val)
		return
	case int64(len(vm.Heap)) > adr:
		// Nothing to do
	case int64(cap(vm.Heap)) > adr:
		vm.Heap = vm.Heap[:adr+1]
	default:
		for int64(len(vm.Heap)) <= adr {
			vm.Heap = append(vm.Heap[:cap(vm.Heap)], 0)
		}
	}
	vm.Heap[adr] = val
}
