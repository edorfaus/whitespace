package parser

import (
	"github.com/edorfaus/whitespace/ast"
)

type state struct {
	f func(*Parser, byte)
	n string
}

var stateStart state

func init() {
	// Cannot be in the variable initializer because it causes a loop
	stateStart = state{(*Parser).stateStart, "start"}
}

func (p *Parser) stateStart(b byte) {
	switch b {
	case ' ':
		p.state = stateStackManip
	case '\t':
		p.state = stateImpTab
	case '\n':
		p.state = stateFlowControl
	default:
		p.badByte(b)
	}
}

var stateImpTab = state{(*Parser).stateImpTab, "impTab"}

func (p *Parser) stateImpTab(b byte) {
	switch b {
	case ' ':
		p.state = stateArithmetic
	case '\t':
		p.state = stateHeapAccess
	case '\n':
		p.state = stateIO
	default:
		p.badByte(b)
	}
}

var stateArithmetic = state{(*Parser).stateArithmetic, "arithmetic"}

func (p *Parser) stateArithmetic(b byte) {
	switch b {
	case ' ':
		p.state = stateArithSpace
	case '\t':
		p.state = stateArithTab
	case '\n':
		p.badCommand("Arithmetic/LF")
	default:
		p.badByte(b)
	}
}

var stateArithSpace = state{(*Parser).stateArithSpace, "arithSpace"}

func (p *Parser) stateArithSpace(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdAdd, nil)
	case '\t':
		p.addCommand(ast.CmdSub, nil)
	case '\n':
		p.addCommand(ast.CmdMul, nil)
	default:
		p.badByte(b)
	}
}

var stateArithTab = state{(*Parser).stateArithTab, "arithTab"}

func (p *Parser) stateArithTab(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdDiv, nil)
	case '\t':
		p.addCommand(ast.CmdMod, nil)
	case '\n':
		p.badCommand("Arithmetic/Tab/LF")
	default:
		p.badByte(b)
	}
}

var stateStackManip = state{(*Parser).stateStackManip, "stackManip"}

func (p *Parser) stateStackManip(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdPush, p.parseNumber())
	case '\t':
		p.state = stateStackManipTab
	case '\n':
		p.state = stateStackManipLF
	default:
		p.badByte(b)
	}
}

var stateStackManipTab = state{(*Parser).stateStackManipTab, "stackManipTab"}

func (p *Parser) stateStackManipTab(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdCopy, p.parseNumber())
	case '\t':
		p.badCommand("Stack Manipulation/Tab/Tab")
	case '\n':
		p.addCommand(ast.CmdSlide, p.parseNumber())
	default:
		p.badByte(b)
	}
}

var stateStackManipLF = state{(*Parser).stateStackManipLF, "stackManipLF"}

func (p *Parser) stateStackManipLF(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdDup, nil)
	case '\t':
		p.addCommand(ast.CmdSwap, nil)
	case '\n':
		p.addCommand(ast.CmdDiscard, nil)
	default:
		p.badByte(b)
	}
}

var stateHeapAccess = state{(*Parser).stateHeapAccess, "heapAccess"}

func (p *Parser) stateHeapAccess(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdStore, nil)
	case '\t':
		p.addCommand(ast.CmdRetrieve, nil)
	case '\n':
		p.badCommand("Heap Access/LF")
	default:
		p.badByte(b)
	}
}

var stateIO = state{(*Parser).stateIO, "io"}

func (p *Parser) stateIO(b byte) {
	switch b {
	case ' ':
		p.state = stateIOSpace
	case '\t':
		p.state = stateIOTab
	case '\n':
		p.badCommand("IO/LF")
	default:
		p.badByte(b)
	}
}

var stateIOSpace = state{(*Parser).stateIOSpace, "ioSpace"}

func (p *Parser) stateIOSpace(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdOutChar, nil)
	case '\t':
		p.addCommand(ast.CmdOutNumber, nil)
	case '\n':
		p.badCommand("IO/Space/LF")
	default:
		p.badByte(b)
	}
}

var stateIOTab = state{(*Parser).stateIOTab, "ioTab"}

func (p *Parser) stateIOTab(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdReadChar, nil)
	case '\t':
		p.addCommand(ast.CmdReadNumber, nil)
	case '\n':
		p.badCommand("IO/Tab/LF")
	default:
		p.badByte(b)
	}
}

var stateFlowControl = state{(*Parser).stateFlowControl, "flowControl"}

func (p *Parser) stateFlowControl(b byte) {
	switch b {
	case ' ':
		p.state = stateFlowControlSpace
	case '\t':
		p.state = stateFlowControlTab
	case '\n':
		p.state = stateFlowControlLF
	default:
		p.badByte(b)
	}
}

var stateFlowControlSpace = state{(*Parser).stateFlowControlSpace, "flowControlSpace"}

func (p *Parser) stateFlowControlSpace(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdMark, p.parseLabel())
	case '\t':
		p.addCommand(ast.CmdCall, p.parseLabel())
	case '\n':
		p.addCommand(ast.CmdJump, p.parseLabel())
	default:
		p.badByte(b)
	}
}

var stateFlowControlTab = state{(*Parser).stateFlowControlTab, "flowControlTab"}

func (p *Parser) stateFlowControlTab(b byte) {
	switch b {
	case ' ':
		p.addCommand(ast.CmdJumpIfZero, p.parseLabel())
	case '\t':
		p.addCommand(ast.CmdJumpIfNeg, p.parseLabel())
	case '\n':
		p.addCommand(ast.CmdReturn, nil)
	default:
		p.badByte(b)
	}
}

var stateFlowControlLF = state{(*Parser).stateFlowControlLF, "flowControlLF"}

func (p *Parser) stateFlowControlLF(b byte) {
	switch b {
	case ' ':
		p.badCommand("Flow Control/LF/Space")
	case '\t':
		p.badCommand("Flow Control/LF/Tab")
	case '\n':
		p.addCommand(ast.CmdExit, nil)
	default:
		p.badByte(b)
	}
}
