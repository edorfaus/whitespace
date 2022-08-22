// This version of the code is designed to not load the entire program
// into memory, but instead interpret it directly from the source file.
//
// That enables it to handle both much larger programs (larger than the
// available memory), and programs that have trailing garbage.
//
// However, that also makes it slower for most reasonable programs, as
// it has to read the file several times to handle loops and such.
// (Also, the current implementation has horrible file handling,
// reading the bytes one at a time without doing any buffering.)
//
// Memory is still used to keep the target locations of any labels it
// has seen, to avoid having to re-scan the file on every jump or call,
// so a program with a lot of labels can still be a problem.

package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var Err error

var s *source

var stack, callStack []int64
var heap = map[int64]int64{}

var labels = map[string]int64{}

func runSource() {
	for Err == nil {
		a, b := s.Next(), s.Next()
		if a == ' ' && b == ' ' {
			// Stack Manipulation: Push
			push(readNumber())
			continue
		}
		c := s.Next()
		is := func(A, B, C byte) bool {
			return a == A && b == B && c == C
		}
		switch {
		// Stack Manipulation
		case is(' ', '\n', ' '): // Duplicate
			push(stack[len(stack)-1])
		case is(' ', '\t', ' '): // Copy
			n := readNumber()
			push(stack[int64(len(stack))-1-n])
		case is(' ', '\n', '\t'): // Swap
			a, b := pop(), pop()
			push(a)
			push(b)
		case is(' ', '\n', '\n'): // Discard
			pop()
		case is(' ', '\t', '\n'): // Slide
			n := readNumber()
			v := pop()
			for ; n > 0; n-- {
				pop()
			}
			push(v)
		// Arithmetic
		case is('\t', ' ', ' '):
			b, a := pop(), pop()
			switch s.Next() {
			case ' ': // Add
				push(a + b)
			case '\t': // Subtract
				push(a - b)
			case '\n': // Multiply
				push(a * b)
			}
		case is('\t', ' ', '\t'):
			switch s.Next() {
			case ' ': // Division
				d := pop()
				push(pop() / d)
			case '\t': // Modulo
				d := pop()
				push(pop() % d)
			case '\n': // Undefined
				fail("unknown instruction: TSTL")
			}
		// Heap Access
		case is('\t', '\t', ' '): // Store
			val, adr := pop(), pop()
			heap[adr] = val
		case is('\t', '\t', '\t'): // Retrieve
			push(heap[pop()])
		// Flow Control
		case is('\n', ' ', ' '): // Mark
			l := readLabel()
			labels[l] = s.Pos()
		case is('\n', ' ', '\t'): // Call
			l := readLabel()
			callStack = append(callStack, s.Pos())
			jump(l)
		case is('\n', ' ', '\n'): // Jump
			jump(readLabel())
		case is('\n', '\t', ' '): // Jump if zero
			l := readLabel()
			if pop() == 0 {
				jump(l)
			}
		case is('\n', '\t', '\t'): // Jump if negative
			l := readLabel()
			if pop() < 0 {
				jump(l)
			}
		case is('\n', '\t', '\n'): // Return
			v := callStack[len(callStack)-1]
			callStack = callStack[:len(callStack)-1]
			s.Seek(v)
		case is('\n', '\n', '\n'): // End
			return
		// I/O
		case is('\t', '\n', ' '):
			switch s.Next() {
			case ' ': // Output character
				_, err := fmt.Printf("%c", pop())
				setErr(err)
			case '\t': // Output number
				_, err := fmt.Printf("%d", pop())
				setErr(err)
			case '\n': // Undefined
				fail("unknown instruction: TLSL")
			}
		case is('\t', '\n', '\t'):
			var v int64
			switch s.Next() {
			case ' ': // Read character
				_, err := fmt.Scanf("%c", &v)
				setErr(err)
				heap[pop()] = v
			case '\t': // Read number
				_, err := fmt.Scanf("%d\n", &v)
				setErr(err)
				heap[pop()] = v
			case '\n': // Undefined
				fail("unknown instruction: TLTL")
			}
		default:
			fail("unknown instruction: %q %q %q", a, b, c)
		}
	}
}

func jump(label string) {
	if pos, ok := labels[label]; ok {
		s.Seek(pos)
		return
	}
	for Err == nil {
		a, b := s.Next(), s.Next()
		if a == ' ' && b == ' ' {
			// Stack Manipulation: Push
			readNumber()
			continue
		}
		c := s.Next()
		is := func(A, B, C byte) bool {
			return a == A && b == B && c == C
		}
		switch {
		// Stack Manipulation
		case is(' ', '\n', ' '): // Duplicate
		case is(' ', '\t', ' '): // Copy
			readNumber()
		case is(' ', '\n', '\t'): // Swap
		case is(' ', '\n', '\n'): // Discard
		case is(' ', '\t', '\n'): // Slide
			readNumber()
		// Arithmetic
		case is('\t', ' ', ' '): // Add, Subtract, Multiply
			s.Next()
		case is('\t', ' ', '\t'):
			switch s.Next() {
			case ' ': // Division
			case '\t': // Modulo
			case '\n': // Undefined
				fail("unknown instruction: TSTL")
			}
		// Heap Access
		case is('\t', '\t', ' '): // Store
		case is('\t', '\t', '\t'): // Retrieve
		// Flow Control
		case is('\n', ' ', ' '): // Mark
			l := readLabel()
			labels[l] = s.Pos()
			if l == label {
				return
			}
		case is('\n', ' ', '\t'): // Call
			readLabel()
		case is('\n', ' ', '\n'): // Jump
			readLabel()
		case is('\n', '\t', ' '): // Jump if zero
			readLabel()
		case is('\n', '\t', '\t'): // Jump if negative
			readLabel()
		case is('\n', '\t', '\n'): // Return
		case is('\n', '\n', '\n'): // End
		// I/O
		case is('\t', '\n', ' '):
			switch s.Next() {
			case ' ': // Output character
			case '\t': // Output number
			case '\n': // Undefined
				fail("unknown instruction: TLSL")
			}
		case is('\t', '\n', '\t'):
			switch s.Next() {
			case ' ': // Read character
			case '\t': // Read number
			case '\n': // Undefined
				fail("unknown instruction: TLTL")
			}
		default:
			fail("unknown instruction: %q %q %q", a, b, c)
		}
	}
}

func readNumber() int64 {
	neg := false
	b := s.Next()
	if Err != nil {
		return 0
	}
	switch b {
	case ' ':
		// Positive
	case '\t':
		neg = true
	default:
		fail("expected sign for number, got %q", b)
		return 0
	}

	var value int64
	for {
		b := s.Next()
		if Err != nil {
			return 0
		}
		switch b {
		case ' ':
			value <<= 1
		case '\t':
			value = (value << 1) | 1
		case '\n':
			if neg {
				return -value
			}
			return value
		}
		if value < 0 {
			fail("read a number too large for this implementation")
			return 0
		}
	}
}

func readLabel() string {
	var sb strings.Builder
	for {
		b := s.Next()
		switch b {
		case ' ', '\t':
			sb.WriteByte(b)
		case '\n':
			return sb.String()
		default:
			fail("unexpected byte during label: %q", b)
			return ""
		}
	}
}

func push(v int64) {
	stack = append(stack, v)
}

func pop() int64 {
	v := stack[len(stack)-1]
	stack = stack[:len(stack)-1]
	return v
}

type source struct {
	file *os.File
}

func (s *source) Next() byte {
	if Err != nil {
		return 0
	}
	var buf [1]byte
	times := 0
	for {
		n, err := s.file.Read(buf[:])
		if setErr(err) {
			return 0
		}
		if n < 1 {
			times++
			if times >= 64 {
				setErr(io.ErrNoProgress)
				return 0
			}
		}
		times = 0
		switch buf[0] {
		case ' ', '\t', '\n':
			return buf[0]
		}
	}
}

func (s *source) Pos() int64 {
	if Err != nil {
		return -1
	}
	pos, err := s.file.Seek(0, io.SeekCurrent)
	setErr(err)
	return pos
}

func (s *source) Seek(pos int64) {
	if Err != nil {
		return
	}
	p, err := s.file.Seek(pos, io.SeekStart)
	setErr(err)
	if p != pos {
		fail("seek failed, %v != %v", p, pos)
	}
}

func run() {
	if Err != nil {
		return
	}
	var file *os.File
	if len(os.Args) > 1 {
		file, Err = os.Open(os.Args[1])
	} else {
		file, Err = os.Open("hello-world.ws")
	}
	if Err != nil {
		return
	}
	defer func() { setErr(file.Close()) }()

	s = &source{
		file: file,
	}

	runSource()
}

func fail(format string, args ...interface{}) {
	if Err == nil {
		Err = fmt.Errorf(format, args...)
	}
}

func setErr(err error) bool {
	if Err == nil {
		Err = err
	}
	return Err != nil
}

func main() {
	run()
	if Err != nil {
		fmt.Fprintln(os.Stderr, "Error:", Err)
		os.Exit(1)
	}
}
