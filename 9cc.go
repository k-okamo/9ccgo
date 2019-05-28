package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// Vector
type Vector struct {
	data     []interface{}
	capacity int
	len      int
}

func new_vec() *Vector {
	v := new(Vector)
	v.data = make([]interface{}, 16)
	v.capacity = 16
	v.len = 0
	return v
}

func vec_push(v *Vector, elem interface{}) {
	if v.len == v.capacity {
		v.data = append(v.data, make([]interface{}, v.capacity)...)
		v.capacity *= 2
	}
	v.data[v.len] = elem
	v.len++
}

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

var tokens *Vector

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
	t := (tokens.data[pos]).(*Token)
	if t.ty != TK_NUM {
		error("number expected, but got %s", t.input)
		return nil
	}
	pos++
	return new_node_num(t.val)
}

func expr() *Node {
	lhs := number()
	for {
		t := (tokens.data[pos]).(*Token)
		op := t.ty
		if op != '+' && op != '-' {
			break
		}
		pos++
		lhs = new_node(op, lhs, number())
	}
	t := (tokens.data[pos]).(*Token)
	if t.ty != TK_EOF {
		error("stray token: %s", t.input)
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

var regno int

func gen_ir_sub(v *Vector, node *Node) int {

	if node.ty == ND_NUM {
		r := regno
		regno++
		vec_push(v, new_ir(IR_IMM, r, node.val))
		return r
	}
	// asset(node->ty == '+' || node-> == '-')

	lhs, rhs := gen_ir_sub(v, node.lhs), gen_ir_sub(v, node.rhs)

	vec_push(v, new_ir(node.ty, lhs, rhs))
	vec_push(v, new_ir(IR_KILL, rhs, 0))
	return lhs
}

func gen_ir(node *Node) *Vector {
	v := new_vec()
	r := gen_ir_sub(v, node)
	vec_push(v, new_ir(IR_RETURN, r, 0))
	return v
}

// Register allocator

var regs = []string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15"}
var used [8]bool
var reg_map []int

func alloc(ir_reg int) int {
	if reg_map[ir_reg] != -1 {
		r := reg_map[ir_reg]
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

func alloc_regs(irv *Vector) {

	reg_map = make([]int, irv.len)
	for i := range reg_map {
		reg_map[i] = -1
	}

	for i := 0; i < irv.len; i++ {
		ir := irv.data[i].(*IR)

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

func gen_X86(irv *Vector) {
	for i := 0; i < irv.len; i++ {
		ir := irv.data[i].(*IR)

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

// [Debug] intermediate reprensations
func print_irs(irs *Vector) {
	if !debug {
		return
	}
	fmt.Println("-- intermediate reprensetations --")
	for i := 0; i < irs.len; i++ {
		ir := irs.data[i].(*IR)
		op := ""
		switch ir.op {
		case IR_IMM:
			op = "IR_IMM   "
		case IR_MOV:
			op = "IR_MOV   "
		case IR_RETURN:
			op = "IR_RETURN"
		case IR_KILL:
			op = "IR_KILL  "
		case IR_NOP:
			op = "IR_NOP   "
		case '+':
			op = "+        "
		case '-':
			op = "-        "
		default:
			op = "         "
		}
		fmt.Printf("[%02d] op: %s, lhs: %d, rhs: %d\n", i, op, ir.lhs, ir.rhs)
	}
	fmt.Println("")
}

var debug bool

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: 9ccgo <code>\n")
		os.Exit(1)
	}
	for i := range reg_map {
		reg_map[i] = -1
	}

	// debug flag
	//debug = true

	// Tokenize and parse.
	tokens = tokenize(os.Args[1])
	print_tokens(tokens)
	node := expr()

	irv := gen_ir(node)
	print_irs(irv)
	alloc_regs(irv)

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")
	gen_X86(irv)
}
