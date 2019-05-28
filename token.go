package main

import (
	"fmt"
	"os"
	"unicode"
)

var (
	tokens *Vector
)

const (
	TK_NUM = iota + 256 // Number literal
	TK_EOF              // End marker
)

// Token type
type Token struct {
	ty    int    // Token type
	val   int    // Number literal
	input string // Token string (for error reporting)
}

// Tokenized input is stored to this array.

func add_token(v *Vector, ty int, input string) *Token {
	t := new(Token)
	t.ty = ty
	t.input = input
	vec_push(v, t)
	return t
}

func tokenize(s string) *Vector {

	v := new_vec()
	i := 0
	for len(s) != 0 {
		c := []rune(s)[0]
		if unicode.IsSpace(c) {
			s = s[1:]
			continue
		}

		// + or -
		if c == '+' || c == '-' {
			add_token(v, int(c), string(c))
			i++
			s = s[1:]
			continue
		}

		// Number
		if unicode.IsDigit(c) {
			t := add_token(v, TK_NUM, string(c))
			val := 0
			val, s = strtol(s, 10)
			t.val = val
			i++
			continue
		}

		fmt.Fprintf(os.Stderr, "cannot tokenize: %s\n", string(c))
		os.Exit(1)
	}

	add_token(v, TK_EOF, s)
	return v
}

// [Debug] tokens print
func print_tokens(tokens *Vector) {
	if !debug {
		return
	}
	fmt.Println("-- tokens info --")
	for i := 0; i < tokens.len; i++ {
		t := tokens.data[i].(*Token)
		ty := ""
		switch t.ty {
		case TK_NUM:
			ty = "TK_NUM"
		case TK_EOF:
			ty = "TK_EOF"
		default:
			ty = "      "
		}
		fmt.Printf("[%02d] ty: %s, val: %d, input: %s\n", i, ty, t.val, t.input)
	}
	fmt.Println("")
}
