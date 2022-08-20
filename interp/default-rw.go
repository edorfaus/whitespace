package interp

import (
	"fmt"
)

// This file holds the default implementations for reading and writing
// characters and numbers.
// For simplicity, these use stdin/stdout and decimal numbers.

func DefaultWriteChar(r rune) error {
	_, err := fmt.Printf("%c", r)
	return err
}

func DefaultWriteNumber(n int64) error {
	_, err := fmt.Printf("%d\n", n)
	return err
}

func DefaultReadChar() (rune, error) {
	var r rune
	_, err := fmt.Scanf("%c", &r)
	return r, err
}

func DefaultReadNumber() (int64, error) {
	var v int64
	_, err := fmt.Scanf("%d\n", &v)
	return v, err
}
