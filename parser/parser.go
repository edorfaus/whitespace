package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/edorfaus/whitespace/ast"
)

type Parser struct {
	src *bufio.Scanner
	err error
}

func New(r io.Reader) *Parser {
	p := &Parser{
		src: bufio.NewScanner(r),
	}
	p.src.Split(splitFunc)
	return p
}

func splitFunc(data []byte, atEOF bool) (int, []byte, error) {
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '\n':
			return i, data[i : i+1], nil
		}
	}
	return len(data), nil, nil
}

func (p *Parser) Err() error {
	if p.err != nil {
		return p.err
	}
	return p.src.Err()
}

func (p *Parser) Parse() {
	imp := p.parseIMP()
	fmt.Println("IMP:", imp)
}

func (p *Parser) parseIMP() ast.Imp {
	b := p.next()
	switch b {
	case ' ':
		return ast.ImpStackManip
	case '\n':
		return ast.ImpFlowControl
	case '\t':
		b = p.next()
		switch b {
		case ' ':
			return ast.ImpArithmetic
		case '\t':
			return ast.ImpHeapAccess
		case '\n':
			return ast.ImpIO
		case 0:
			p.unexpectedEOF("IMP")
			return ast.ImpNone
		default:
			p.badByte(b)
			return ast.ImpNone
		}
	case 0:
		return ast.ImpNone
	default:
		p.badByte(b)
		return ast.ImpNone
	}
}

func (p *Parser) parseLabel() string {
	var sb strings.Builder
	for {
		b := p.next()
		switch b {
		case ' ', '\t':
			sb.WriteByte(b)
		case '\n':
			return sb.String()
		case 0:
			p.unexpectedEOF("label")
			return ""
		default:
			p.badByte(b)
			return ""
		}
	}
}

func (p *Parser) next() byte {
	if p.err == nil && p.src.Scan() {
		return p.src.Bytes()[0]
	}
	return 0
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

func (p *Parser) unexpectedEOF(during string) {
	p.fail("unexpected EOF while reading %s", during)
}
