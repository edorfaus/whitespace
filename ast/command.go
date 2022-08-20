package ast

type Command struct {
	Cmd Cmd
	Arg interface{}
}

type Cmd uint8

const (
	CmdNone Cmd = iota
	// IMP: Stack Manipulation: [Space]
	CmdPush
	CmdDup
	CmdCopy
	CmdSwap
	CmdDiscard
	CmdSlide
	// IMP: Arithmetic: [Tab][Space]
	CmdAdd
	CmdSub
	CmdMul
	CmdDiv
	CmdMod
	// IMP: Heap Access: [Tab][Tab]
	CmdStore
	CmdRetrieve
	// IMP: Flow Control: [LF]
	CmdMark
	CmdCall
	CmdJump
	CmdJumpIfZero
	CmdJumpIfNeg
	CmdReturn
	CmdExit
	// IMP: I/O: [Tab][LF]
	CmdOutChar
	CmdOutNumber
	CmdReadChar
	CmdReadNumber

	// Total count of commands
	CountCmds
)
