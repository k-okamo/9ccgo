package main

import (
	"log"
	"os"
)

var (
	debug    bool
	filename string
)

func main() {

	//debug = true
	if len(os.Args) == 1 {
		usage()
	}
	if len(os.Args) == 2 && os.Args[1] == "-test" {
		util_test()
		os.Exit(0)
	}

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
			usage()
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
	f := os.Stdin
	if filename != "-" {
		f2, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		f = f2
		defer f2.Close()
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
		sb_append_n(sb, string(buf[:n]), n)

	}

	if sb.data[sb.len-1] != '\n' {
		sb_add(sb, "\n")
	}
	return sb_get(sb)
}

func usage() {
	error("Usage: 9ccgo [-test] [-dump-ir1] [-dump-ir2] <file>")
}
