package main

import (
	"fmt"
	"os"
)

var (
	debug bool
)

func main() {

	//debug = true
	if len(os.Args) > 1 && os.Args[1] == "-test" {
		util_test()
		os.Exit(0)
	}

	var input string
	dump_ir1 := false
	dump_ir2 := false

	if len(os.Args) == 3 && os.Args[1] == "-dump-ir1" {
		dump_ir1 = true
		input = os.Args[2]
	} else if len(os.Args) == 3 && os.Args[1] == "-dump-ir2" {
		dump_ir2 = true
		input = os.Args[2]
	} else {
		if len(os.Args) != 2 {
			fmt.Fprintf(os.Stderr, "Usage: 9ccgo [-test] [-dump-ir1] [-dump-ir2] <code>\n")
			os.Exit(0)
		}
		input = os.Args[1]
	}

	// Tokenize and parse.
	tokens = tokenize(input)
	print_tokens(tokens) // Debug
	fns := gen_ir(parse(tokens))

	print_irs(fns) // Debug
	if dump_ir1 {
		dump_ir(fns)
	}

	alloc_regs(fns)
	if dump_ir2 {
		dump_ir(fns)
	}

	gen_x86(fns)
}
