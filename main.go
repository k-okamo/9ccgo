package main

import (
	"fmt"
	"os"
)

var (
	debug bool
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: 9ccgo <code>\n")
		os.Exit(1)
	}

	//debug = true
	if os.Args[1] == "-test" {
		util_test()
		os.Exit(0)
	}

	// Tokenize and parse.
	tokens = tokenize(os.Args[1])
	print_tokens(tokens) // Debug
	node := parse(tokens)

	irv := gen_ir(node)
	print_irs(irv) // Debug
	alloc_regs(irv)

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	gen_X86(irv)
}
