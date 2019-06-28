package main

import (
	"fmt"
	"log"
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

	var filename string
	dump_ir1 := false
	dump_ir2 := false

	if len(os.Args) == 3 && os.Args[1] == "-dump-ir1" {
		dump_ir1 = true
		filename = os.Args[2]
	} else if len(os.Args) == 3 && os.Args[1] == "-dump-ir2" {
		dump_ir2 = true
		filename = os.Args[2]
	} else {
		if len(os.Args) != 2 {
			fmt.Fprintf(os.Stderr, "Usage: 9ccgo [-test] [-dump-ir1] [-dump-ir2] <file>\n")
			os.Exit(0)
		}
		filename = os.Args[1]
	}

	// Tokenize and parse.
	input := read_file(filename)
	tokens = tokenize(input)
	print_tokens(tokens) // Debug
	nodes := parse(tokens)
	globals := sema(nodes)
	fns := gen_ir(nodes)

	print_irs(fns) // Debug
	if dump_ir1 {
		dump_ir(fns)
	}

	alloc_regs(fns)
	if dump_ir2 {
		dump_ir(fns)
	}

	gen_x86(globals, fns)
}

func read_file(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	buf := make([]byte, 1024)
	sb := new_sb()
	for {
		n, err := f.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
		sb_lappend(sb, string(buf[:n]), n)
	}
	return sb_get(sb)
}
