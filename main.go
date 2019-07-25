package main

import (
	"os"
)

func main() {

	if len(os.Args) == 1 {
		usage()
	}
	if len(os.Args) == 2 && os.Args[1] == "-test" {
		util_test()
		os.Exit(0)
	}

	path := ""
	dump_ir1 := false
	dump_ir2 := false

	if len(os.Args) == 3 && os.Args[1] == "-dump-ir1" {
		dump_ir1 = true
		path = os.Args[2]
	} else if len(os.Args) == 3 && os.Args[1] == "-dump-ir2" {
		dump_ir2 = true
		path = os.Args[2]
	} else {
		if len(os.Args) != 2 {
			usage()
		}
		path = os.Args[1]
	}

	// Tokenize and parse.
	tokens := tokenize(path, true)
	nodes := parse(tokens)
	globals := sema(nodes)
	fns := gen_ir(nodes)

	if dump_ir1 {
		dump_ir(fns)
	}

	alloc_regs(fns)
	if dump_ir2 {
		dump_ir(fns)
	}

	gen_x86(globals, fns)
}

func usage() { error("Usage: 9ccgo [-test] [-dump-ir1] [-dump-ir2] <file>") }
