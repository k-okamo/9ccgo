package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: 9ccgo <code>\n")
		os.Exit(1)
	}

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	n, s := strtol(os.Args[1], 10)
	fmt.Printf("\tmov rax, %d\n", n)

	for len(s) != 0 {
		if []rune(s)[0] == '+' {
			s = s[1:]
			n, s = strtol(s, 10)
			fmt.Printf("\tadd rax, %d\n", n)
			continue
		}
		if []rune(s)[0] == '-' {
			s = s[1:]
			n, s = strtol(s, 10)
			fmt.Printf("\tsub rax, %d\n", n)
			continue
		}

		fmt.Fprintf(os.Stderr, "unexpected character: %c\n", []rune(s)[0])
		os.Exit(1)
	}

	fmt.Printf("\tret\n")
}

func strtol(s string, b int) (int64, string) {
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
	n, _ := strconv.ParseInt(s[:j], b, 64)
	return n, s[j:]

}
