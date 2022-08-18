package ast

type Imp uint8

const (
	ImpNone Imp = iota
	ImpStackManip
	ImpArithmetic
	ImpHeapAccess
	ImpFlowControl
	ImpIO
)

var impString = []string{
	"IMP:None",
	"IMP:Stack Manipulation",
	"IMP:Arithmetic",
	"IMP:Heap Access",
	"IMP:Flow Control",
	"IMP:I/O",
}

func (i Imp) String() string {
	if int(i) < len(impString) {
		return impString[i]
	}
	return "IMP:Unknown"
}
