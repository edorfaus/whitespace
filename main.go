package main

import (
	"fmt"
	"os"

	"github.com/edorfaus/whitespace/interp"
	"github.com/edorfaus/whitespace/parser"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	fn := "hello-world.ws"
	if len(os.Args) > 1 {
		fn = os.Args[1]
	}
	p, err := parseFile(fn)
	if err != nil {
		return err
	}

	vm := interp.NewVM(p.Commands)
	if vm.Err != nil {
		return vm.Err
	}

	vm.Run()
	if vm.Err != nil {
		return vm.Err
	}

	return nil
}

func parseFile(fn string) (par *parser.Parser, retErr error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	p := parser.New(f)
	p.Parse()

	return p, p.Err()
}
