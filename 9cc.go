package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

const (
	TK_NUM = iota + 1 // Number literal
	TK_EOF            // End marker
)

// Token type
type Token struct {
	ty    int    // Token type
	val   int    // Number literal
	input string // Token string (for error reporting)
}

// Tokenized input is stored to this array.
var tokens = make([]Token, 100)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: 9ccgo <code>\n")
		os.Exit(1)
	}

	tokenize(os.Args[1])

	// Print the prologue
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// Verify that the given expression starts with a number,
	// and then emit the first `mov` instruction.
	if tokens[0].ty != TK_NUM {
		fail(0)
	}
	fmt.Printf("\tmov rax, %d\n", tokens[0].val)

	// Emit assembly as we consume the sequence of `+ <number>`
	// or `- <number>`
	i := 1
	for tokens[i].ty != TK_EOF {
		if tokens[i].ty == '+' {
			i++
			if tokens[i].ty != TK_NUM {
				fail(i)
			}
			fmt.Printf("\tadd rax, %d\n", tokens[i].val)
			i++
			continue
		}

		if tokens[i].ty == '-' {
			i++
			if tokens[i].ty != TK_NUM {
				fail(i)
			}
			fmt.Printf("\tsub rax, %d\n", tokens[i].val)
			i++
			continue
		}

		fail(i)
	}

	fmt.Printf("\tret\n")
}

func tokenize(s string) {
	i := 0
	for len(s) != 0 {
		c := []rune(s)[0]
		if unicode.IsSpace(c) {
			s = s[1:]
			continue
		}

		// + or -
		if c == '+' || c == '-' {
			tokens[i].ty = int(c)
			tokens[i].input = string(c)
			i++
			s = s[1:]
			continue
		}

		// Number
		if unicode.IsDigit(c) {
			tokens[i].ty = TK_NUM
			tokens[i].input = string(c)
			var val int
			val, s = strtol(s, 10)
			tokens[i].val = val
			i++
			continue
		}

		fmt.Fprintf(os.Stderr, "cannot tokenize: %s\n", string(c))
		os.Exit(1)
	}

	tokens[i].ty = TK_EOF
}

// An error reporting function
func fail(i int) {
	fmt.Fprintf(os.Stderr, "unexpected token: %s\n", tokens[i].input)
	os.Exit(1)
}

func strtol(s string, b int) (int, string) {
	if !unicode.IsDigit([]rune(s)[0]) {
		return 0, s
	}

	j := len(s)
	for i, c := range s {
		if !unicode.IsDigit(c) {
			j = i
			break
		}
	}
	n, _ := strconv.ParseInt(s[:j], b, 32)
	return int(n), s[j:]

}
