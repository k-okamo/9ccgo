package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// Tokenizer

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
var tokens = make([]Token, 100)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: 9ccgo <code>\n")
		os.Exit(1)
	}

	tokenize(os.Args[1])
	node := expr()

	// Print the prologue.
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// Generate code while descending the parse tree.
	reg := gen(node)
	fmt.Printf("\tmov rax, %s\n", reg)
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

// Recursive-descendent parser
var pos = 0

const (
	ND_NUM = iota + 256 // Number literal
)

type Node struct {
	ty  int   // Node type
	lhs *Node // left-hand side
	rhs *Node // right-hand side
	val int   // Number literal
}

func new_node(op int, lhs, rhs *Node) *Node {
	node := new(Node)
	node.ty = op
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func new_node_num(val int) *Node {
	node := new(Node)
	node.ty = ND_NUM
	node.val = val
	return node
}

func number() *Node {
	if tokens[pos].ty == TK_NUM {
		node := new_node_num(tokens[pos].val)
		pos++
		return node
	}
	error("number expected, but got %s", tokens[pos].input)
	return nil
}

func expr() *Node {
	lhs := number()
	for {
		op := tokens[pos].ty
		if op != '+' && op != '-' {
			break
		}
		pos++
		lhs = new_node(op, lhs, number())
	}
	if tokens[pos].ty != TK_EOF {
		error("stray token: %s", tokens[pos].input)
	}
	return lhs
}

// Code generator

var regs = []string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15", ""}
var cur = 0

func gen(node *Node) string {
	if node.ty == ND_NUM {
		reg := regs[cur]
		if regs[cur] == "" {
			error("register exhausted")
		}
		cur++
		fmt.Printf("\tmov %s, %d\n", reg, node.val)
		return reg
	}

	dst, src := gen(node.lhs), gen(node.rhs)

	switch node.ty {
	case '+':
		fmt.Printf("\tadd %s, %s\n", dst, src)
		return dst
	case '-':
		fmt.Printf("\tsub %s, %s\n", dst, src)
		return dst
	default:
		error("unknown operator")
	}
	return ""
}

// An error reporting function
func error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintf(os.Stderr, "\n")
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
