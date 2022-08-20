package interp

import (
	"errors"

	"github.com/edorfaus/whitespace/ast"
)

type Instr struct {
	Op  func(*VM, int64)
	Arg int64
}

// errProgramExit is a sentinel value for when the program executed the
// end-the-program operation (the exit instruction).
var errProgramExit = errors.New("program exited")

var vmOps = [ast.CountCmds]func(*VM, int64){
	(*VM).opInvalid, // CmdNone
	// IMP: Stack Manipulation
	(*VM).opPush,
	(*VM).opDup,
	(*VM).opCopy,
	(*VM).opSwap,
	(*VM).opDiscard,
	(*VM).opSlide,
	// IMP: Arithmetic
	(*VM).opAdd,
	(*VM).opSub,
	(*VM).opMul,
	(*VM).opDiv,
	(*VM).opMod,
	// IMP: Heap Access
	(*VM).opStore,
	(*VM).opRetrieve,
	// IMP: Flow Control
	(*VM).opInvalid, // CmdMark
	(*VM).opCall,
	(*VM).opJump,
	(*VM).opJumpIfZero,
	(*VM).opJumpIfNeg,
	(*VM).opReturn,
	(*VM).opExit,
	// IMP: I/O
	(*VM).opOutChar,
	(*VM).opOutNumber,
	(*VM).opReadChar,
	(*VM).opReadNumber,
}

func (vm *VM) opInvalid(_ int64) {
	vm.fail("invalid opcode")
}

func (vm *VM) opPush(arg int64) {
	vm.Stack = append(vm.Stack, arg)
}

func (vm *VM) opDup(_ int64) {
	if vm.stackSize(1) {
		vm.Stack = append(vm.Stack, vm.Stack[len(vm.Stack)-1])
	}
}

func (vm *VM) opCopy(arg int64) {
	// TODO validate arg is within int range
	n := int(arg)
	if vm.stackSize(n + 1) {
		vm.Stack = append(vm.Stack, vm.Stack[len(vm.Stack)-1-n])
	}
}

func (vm *VM) opSwap(_ int64) {
	if vm.stackSize(2) {
		p := len(vm.Stack) - 2
		vm.Stack[p], vm.Stack[p+1] = vm.Stack[p+1], vm.Stack[p]
	}
}

func (vm *VM) opDiscard(_ int64) {
	if vm.stackSize(1) {
		vm.pop()
	}
}

func (vm *VM) opSlide(arg int64) {
	// TODO validate arg is within int range
	n := int(arg)
	if vm.stackSize(n + 1) {
		p := len(vm.Stack) - 1 - n
		vm.Stack[p] = vm.Stack[len(vm.Stack)-1]
		vm.Stack = vm.Stack[:p+1]
	}
}

func (vm *VM) opAdd(_ int64) {
	if vm.stackSize(2) {
		b := vm.pop()
		vm.Stack[len(vm.Stack)-1] += b
	}
}

func (vm *VM) opSub(_ int64) {
	if vm.stackSize(2) {
		b := vm.pop()
		vm.Stack[len(vm.Stack)-1] -= b
	}
}

func (vm *VM) opMul(_ int64) {
	if vm.stackSize(2) {
		b := vm.pop()
		vm.Stack[len(vm.Stack)-1] *= b
	}
}

func (vm *VM) opDiv(_ int64) {
	if vm.stackSize(2) {
		b := vm.pop()
		vm.Stack[len(vm.Stack)-1] /= b
	}
}

func (vm *VM) opMod(_ int64) {
	if vm.stackSize(2) {
		b := vm.pop()
		vm.Stack[len(vm.Stack)-1] %= b
	}
}

func (vm *VM) opStore(_ int64) {
	if vm.stackSize(2) {
		adr, val := vm.Stack[len(vm.Stack)-2], vm.Stack[len(vm.Stack)-1]
		vm.Stack = vm.Stack[:len(vm.Stack)-2]
		vm.storeHeap(adr, val)
	}
}

func (vm *VM) opRetrieve(_ int64) {
	if vm.stackSize(1) {
		p := len(vm.Stack) - 1
		adr := vm.Stack[p]
		switch {
		case adr < 0:
			vm.fail("retrieve from negative heap address: %v", adr)
			return
		case int64(len(vm.Heap)) > adr:
			vm.Stack[p] = vm.Heap[adr]
		default:
			vm.Stack[p] = 0
		}
	}
}

func (vm *VM) opCall(arg int64) {
	vm.RetTo = append(vm.RetTo, vm.PC)
	vm.PC = int(arg)
}

func (vm *VM) opReturn(_ int64) {
	if len(vm.RetTo) < 1 {
		vm.fail("return with empty call stack")
		return
	}
	p := len(vm.RetTo) - 1
	vm.PC = vm.RetTo[p]
	vm.RetTo = vm.RetTo[:p]
}

func (vm *VM) opJump(arg int64) {
	vm.PC = int(arg)
}

func (vm *VM) opJumpIfZero(arg int64) {
	if vm.stackSize(1) && vm.pop() == 0 {
		vm.PC = int(arg)
	}
}

func (vm *VM) opJumpIfNeg(arg int64) {
	if vm.stackSize(1) && vm.pop() < 0 {
		vm.PC = int(arg)
	}
}

func (vm *VM) opExit(_ int64) {
	if vm.Err == nil {
		vm.Err = errProgramExit
	}
}

func (vm *VM) opOutChar(_ int64) {
	if vm.stackSize(1) {
		err := vm.WriteChar(rune(vm.pop()))
		if err != nil && vm.Err == nil {
			vm.Err = err
		}
	}
}

func (vm *VM) opOutNumber(_ int64) {
	if vm.stackSize(1) {
		err := vm.WriteNumber(vm.pop())
		if err != nil && vm.Err == nil {
			vm.Err = err
		}
	}
}

func (vm *VM) opReadChar(_ int64) {
	if vm.stackSize(1) {
		adr := vm.pop()
		ch, err := vm.ReadChar()
		if err != nil && vm.Err == nil {
			vm.Err = err
			return
		}
		vm.storeHeap(adr, int64(ch))
	}
}

func (vm *VM) opReadNumber(_ int64) {
	if vm.stackSize(1) {
		adr := vm.pop()
		val, err := vm.ReadNumber()
		if err != nil && vm.Err == nil {
			vm.Err = err
			return
		}
		vm.storeHeap(adr, val)
	}
}
