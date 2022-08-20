package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/edorfaus/whitespace/ast"
)

type Parser struct {
	Commands []ast.Command

	state state

	src *bufio.Scanner
	err error
}

func New(r io.Reader) *Parser {
	p := &Parser{
		state: stateStart,
		src:   bufio.NewScanner(r),
	}
	p.src.Split(splitFunc)
	return p
}

func splitFunc(data []byte, atEOF bool) (int, []byte, error) {
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '\n':
			return i + 1, data[i : i+1], nil
		}
	}
	return len(data), nil, nil
}

func (p *Parser) Err() error {
	return p.err
}

func (p *Parser) Parse() {
	for p.err == nil && p.src.Scan() {
		b := p.src.Bytes()[0]
		p.state.f(p, b)
	}
	if p.err == nil {
		p.err = p.src.Err()
	}
	if p.err == nil && p.state.n != stateStart.n {
		p.fail("unexpected EOF in state %v", p.state.n)
	}
}

func (p *Parser) addCommand(c ast.Cmd, a interface{}) {
	p.Commands = append(p.Commands, ast.Command{Cmd: c, Arg: a})
	p.state = stateStart
}

func (p *Parser) parseLabel() string {
	if p.err != nil {
		return ""
	}
	var sb strings.Builder
	for {
		if !p.mustScan("a label") {
			return ""
		}
		b := p.src.Bytes()[0]
		switch b {
		case ' ', '\t':
			sb.WriteByte(b)
		case '\n':
			return sb.String()
		default:
			p.badByte(b)
			return ""
		}
	}
}

func (p *Parser) parseNumber() int64 {
	if p.err != nil {
		return 0
	}

	// First read the sign byte
	if !p.mustScan("sign for number") {
		return 0
	}
	neg := false
	b := p.src.Bytes()[0]
	switch b {
	case ' ':
		// positive
	case '\t':
		neg = true
	case '\n':
		p.fail("unexpected LF, expected a sign (space/tab)")
		return 0
	default:
		p.badByte(b)
		return 0
	}

	// Then skip leading zeroes
	for {
		if !p.mustScan("a number") {
			return 0
		}
		b = p.src.Bytes()[0]
		if b != ' ' {
			break
		}
	}
	switch b {
	case '\t':
		// non-0 value
	case '\n':
		// all-0 value
		return 0
	case ' ':
		// cannot happen
		p.fail("code error: should never get here")
		return 0
	default:
		p.badByte(b)
		return 0
	}

	// Next, read the number (as far as it fits)
	value := int64(1)
	bits := 1
	for {
		if !p.mustScan("a number") {
			return 0
		}
		b = p.src.Bytes()[0]
		switch b {
		case ' ':
			bits++
			value <<= 1
		case '\t':
			bits++
			value = (value << 1) | 1
		case '\n':
			if neg {
				return -value
			}
			return value
		default:
			p.badByte(b)
			return 0
		}
		if bits > 63 {
			p.fail("number too large for implementation (>63 bits)")
			return 0
		}
	}
}

func (p *Parser) mustScan(during string) bool {
	if p.err == nil && !p.src.Scan() {
		p.err = p.src.Err()
		p.unexpectedEOF(during)
	}
	return p.err == nil
}

func (p *Parser) fail(format string, args ...interface{}) {
	if p.err != nil {
		return
	}
	p.err = fmt.Errorf(format, args...)
}

func (p *Parser) badByte(b byte) {
	p.fail("unexpected byte from scanner: %02X '%c'", b, b)
}

func (p *Parser) badCommand(what string) {
	p.fail("invalid command: %v", what)
}

func (p *Parser) unexpectedEOF(during string) {
	p.fail("unexpected EOF while reading %s", during)
}
