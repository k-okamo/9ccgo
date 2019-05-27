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

	for i := range reg_map {
		reg_map[i] = -1
	}

	// Tokenize and parse.
	tokenize(os.Args[1])
	node := expr()

	gen_ir(node)
	alloc_regs()

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	gen_X86()
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

// Intermediate reperentation

const (
	IR_IMM = iota
	IR_MOV
	IR_RETURN
	IR_KILL
	IR_NOP
)

type IR struct {
	op  int
	lhs int
	rhs int
}

func new_ir(op, lhs, rhs int) *IR {
	ir := new(IR)
	ir.op = op
	ir.lhs = lhs
	ir.rhs = rhs
	return ir
}

var ins = make([]*IR, 100)
var inp int
var regno int

func gen_ir_sub(node *Node) int {

	if node.ty == ND_NUM {
		r := regno
		regno++
		ins[inp] = new_ir(IR_IMM, r, node.val)
		inp++
		return r
	}
	// asset(node->ty == '+' || node-> == '-')

	lhs, rhs := gen_ir_sub(node.lhs), gen_ir_sub(node.rhs)

	ins[inp] = new_ir(node.ty, lhs, rhs)
	inp++
	ins[inp] = new_ir(IR_KILL, rhs, 0)
	inp++
	return lhs
}

func gen_ir(node *Node) {
	r := gen_ir_sub(node)
	ins[inp] = new_ir(IR_RETURN, r, 0)
	inp++
}

// Register allocator

var regs = []string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15", ""}
var used [8]bool
var reg_map [1000]int

func alloc(ir_reg int) int {
	if reg_map[ir_reg] != -1 {
		r := reg_map[ir_reg]
		//de
		if !used[r] {
			fmt.Printf("used[%d] is not true.\n", r)
		}
		//de
		//assert(used[r])
		return r
	}

	for i := 0; i < len(regs); i++ {
		if used[i] == true {
			continue
		}
		used[i] = true
		reg_map[ir_reg] = i
		return i
	}
	error("register exhausted")
	return -1
}

func kill(r int) {
	//asset(used[r])
	used[r] = false
}

func alloc_regs() {
	for i := 0; i < inp; i++ {
		ir := ins[i]

		switch ir.op {
		case IR_IMM:
			ir.lhs = alloc(ir.lhs)
		case IR_MOV, '+', '-':
			ir.lhs = alloc(ir.lhs)
			ir.rhs = alloc(ir.rhs)
		case IR_RETURN:
			kill(reg_map[ir.lhs])
		case IR_KILL:
			kill(reg_map[ir.lhs])
			ir.op = IR_NOP
		default:
			//asset(0&& "unknown operator")
		}
	}
}

// Code generator

func gen_X86() {
	for i := 0; i < inp; i++ {
		ir := ins[i]

		switch ir.op {
		case IR_IMM:
			fmt.Printf("\tmov %s, %d\n", regs[ir.lhs], ir.rhs)
		case IR_MOV:
			fmt.Printf("\tmov %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_RETURN:
			fmt.Printf("\tmov rax, %s\n", regs[ir.lhs])
			fmt.Printf("\tret\n")
		case '+':
			fmt.Printf("\tadd %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case '-':
			fmt.Printf("\tsub %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_NOP:
			break
		default:
			//asset(0 && "unknown operator")
		}
	}
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
