package main

// 9ccgo's code generation is two-pass. In the first pass, abstract
// syntax trees are compiled to IT (intermediate representation).
//
// IR resembles the real x86-64 instruction set, but it has infinite
// number of registers. We don't try too hard to reuse registers in
// this pass. Instead, we "kill" registers to mark them as dead when
// we are done with them and use new registers.
//
// Such infinite number of registers are mapped to a finite registers
// in a later pass.

import (
	"fmt"
)

var (
	code         *Vector
	nreg         = 1
	nlabel       = 1
	return_label int
	return_reg   int
	break_label  int
)

const (
	IR_ADD = iota + 256
	IR_ADD_IMM
	IR_SUB
	IR_SUB_IMM
	IR_MUL
	IR_MUL_IMM
	IR_DIV
	IR_IMM
	IR_BPREL
	IR_MOV
	IR_RETURN
	IR_CALL
	IR_LABEL
	IR_LABEL_ADDR
	IR_EQ
	IR_NE
	IR_LE
	IR_LT
	IR_AND
	IR_OR
	IR_XOR
	IR_SHL
	IR_SHR
	IR_MOD
	IR_NEG
	IR_JMP
	IR_IF
	IR_UNLESS
	IR_LOAD
	IR_STORE
	IR_STORE8_ARG
	IR_STORE32_ARG
	IR_STORE64_ARG
	IR_KILL
	IR_NOP
)

const (
	IR_TY_NOARG = iota + 256
	IR_TY_REG
	IR_TY_IMM
	IR_TY_MEM
	IR_TY_JMP
	IR_TY_LABEL
	IR_TY_LABEL_ADDR
	IR_TY_REG_REG
	IR_TY_REG_IMM
	IR_TY_IMM_IMM
	IR_TY_REG_LABEL
	IR_TY_CALL
)

type IR struct {
	op  int
	lhs int
	rhs int

	// Load/Store size in bytes
	size int

	// Function call
	name  string
	nargs int
	args  [6]int
}

type Function struct {
	name      string
	stacksize int
	globals   *Vector
	ir        *Vector
}

func add(op, lhs, rhs int) *IR {
	ir := new(IR)
	ir.op = op
	ir.lhs = lhs
	ir.rhs = rhs
	vec_push(code, ir)
	return ir
}

func kill(r int) {
	add(IR_KILL, r, -1)
}

func label(x int) {
	add(IR_LABEL, x, -1)
}

func jmp(x int) {
	add(IR_JMP, x, -1)
}

func choose_insn(node *Node, op8, op32, op64 int) int {
	if node.ty.size == 1 {
		return op8
	}
	if node.ty.size == 4 {
		return op32
	}
	// assert(node.ty.size == 8)
	return op64
}

func load(node *Node, dst, src int) {
	ir := add(IR_LOAD, dst, src)
	ir.size = node.ty.size
}

func store(node *Node, dst, src int) {
	ir := add(IR_STORE, dst, src)
	ir.size = node.ty.size
}

func store_arg_insn(node *Node) int {
	return choose_insn(node, IR_STORE8_ARG, IR_STORE32_ARG, IR_STORE64_ARG)
}

// In C, all expressions that can be written on the left-hand side of
// the '=' operator must habe an address in memory. IN other words, if
// you can apply the '&' operator to take an address of some
// expression E, you can assign E to a new value.
//
// Other expressions, such as `1+2`, cannot be written on the lhs of
// '=', since they are just temporary values that don't have an address.
//
// The stuff that can be written on the lhs of '=' os called lvalue.
// Other values are called rvalue. An lvalue is essentially an address.
//
// When lvalues appear on the rvalue context, they are converted to
// rvalues by loading their values from their addresses. You can think
// '&' as an operator that suppresses such auutomatic lvalue-to-rvalue
// conversion.
//
// This function evaluates a given node as an lvalue.
func gen_lval(node *Node) int {
	if node.op == ND_DEREF {
		return gen_expr(node.expr)
	}

	if node.op == ND_DOT {
		r := gen_lval(node.expr)
		add(IR_ADD_IMM, r, node.offset)
		return r
	}

	if node.op == ND_LVAR {
		r := nreg
		nreg++
		add(IR_BPREL, r, node.offset)
		return r
	}
	// assert(node.op == ND_GVAR)
	r := nreg
	nreg++
	ir := add(IR_LABEL_ADDR, r, -1)
	ir.name = node.name
	return r
}

func gen_binop(ty int, node *Node) int {
	lhs, rhs := gen_expr(node.lhs), gen_expr(node.rhs)
	add(ty, lhs, rhs)
	kill(rhs)
	return lhs
}

func gen_expr(node *Node) int {

	switch node.op {
	case ND_NUM:
		{
			r := nreg
			nreg++
			add(IR_IMM, r, node.val)
			return r
		}
	case ND_EQ:
		return gen_binop(IR_EQ, node)
	case ND_NE:
		return gen_binop(IR_NE, node)
	case ND_LOGAND:
		{
			x := nlabel
			nlabel++
			r1 := gen_expr(node.lhs)
			add(IR_UNLESS, r1, x)
			r2 := gen_expr(node.rhs)
			add(IR_MOV, r1, r2)
			kill(r2)
			add(IR_UNLESS, r1, x)
			add(IR_IMM, r1, 1)
			label(x)
			return r1
		}
	case ND_LOGOR:
		{
			x := nlabel
			nlabel++
			y := nlabel
			nlabel++
			r1 := gen_expr(node.lhs)
			add(IR_UNLESS, r1, x)
			add(IR_IMM, r1, 1)
			jmp(y)
			label(x)
			r2 := gen_expr(node.rhs)
			add(IR_MOV, r1, r2)
			kill(r2)
			add(IR_UNLESS, r1, y)
			add(IR_IMM, r1, 1)
			label(y)
			return r1
		}
	case ND_GVAR, ND_LVAR, ND_DOT:
		{
			r := gen_lval(node)
			load(node, r, r)
			return r
		}

	case ND_CALL:
		{
			var args [6]int
			for i := 0; i < node.args.len; i++ {
				args[i] = gen_expr(node.args.data[i].(*Node))
			}
			r := nreg
			nreg++

			ir := add(IR_CALL, r, -1)
			ir.name = node.name
			ir.nargs = node.args.len
			for i := 0; i < 6; i++ {
				ir.args[i] = args[i]
			}
			for i := 0; i < ir.nargs; i++ {
				kill(ir.args[i])
			}
			return r
		}
	case ND_ADDR:
		{
			return gen_lval(node.expr)
		}
	case ND_DEREF:
		{
			r := gen_expr(node.expr)
			load(node, r, r)
			return r
		}
	case ND_STMT_EXPR:
		{
			orig_label := return_label
			orig_reg := return_reg
			return_label = nlabel
			nlabel++
			r := nreg
			nreg++
			return_reg = r

			gen_stmt(node.body)
			label(return_label)

			return_label = orig_label
			return_reg = orig_reg
			return r
		}
	case '=':
		{
			rhs, lhs := gen_expr(node.rhs), gen_lval(node.lhs)
			store(node, lhs, rhs)
			kill(rhs)
			return lhs
		}
	case '+', '-':
		{
			insn := IR_SUB
			if node.op == '+' {
				insn = IR_ADD
			}
			if node.lhs.ty.ty != PTR {
				return gen_binop(insn, node)
			}

			rhs := gen_expr(node.rhs)
			add(IR_MUL_IMM, rhs, node.lhs.ty.ptr_to.size)

			lhs := gen_expr(node.lhs)
			add(insn, lhs, rhs)
			kill(rhs)
			return lhs
		}
	case '*':
		return gen_binop(IR_MUL, node)
	case '/':
		return gen_binop(IR_DIV, node)
	case '%':
		return gen_binop(IR_MOD, node)
	case '<':
		return gen_binop(IR_LT, node)
	case ND_LE:
		return gen_binop(IR_LE, node)
	case '&':
		return gen_binop(IR_AND, node)
	case '|':
		return gen_binop(IR_OR, node)
	case '^':
		return gen_binop(IR_XOR, node)
	case ND_SHL:
		return gen_binop(IR_SHL, node)
	case ND_SHR:
		return gen_binop(IR_SHR, node)
	case ND_NEG:
		{
			r := gen_expr(node.expr)
			add(IR_NEG, r, -1)
			return r
		}
	case ND_PRE_INC:
		return gen_pre_inc(node, 1)
	case ND_PRE_DEC:
		return gen_pre_inc(node, -1)
	case ND_POST_INC:
		return gen_post_inc(node, 1)
	case ND_POST_DEC:
		return gen_post_inc(node, -1)
	case ',':
		kill(gen_expr(node.lhs))
		return gen_expr(node.rhs)
	case '?':
		{
			x := nlabel
			nlabel++
			y := nlabel
			nlabel++
			r := gen_expr(node.cond)

			add(IR_UNLESS, r, x)
			r2 := gen_expr(node.then)
			add(IR_MOV, r, r2)
			kill(r2)
			jmp(y)

			label(x)
			r3 := gen_expr(node.els)
			add(IR_MOV, r, r3)
			kill(r2)
			label(y)
			return r
		}
	case '!':
		{
			lhs := gen_expr(node.expr)
			rhs := nreg
			nreg++
			add(IR_IMM, rhs, 0)
			add(IR_EQ, lhs, rhs)
			kill(rhs)
			return lhs
		}
	default:
		//assert(0 && "unknown AST type")
	}

	return 0
}

func get_inc_scale(node *Node) int {
	if node.ty.ty == PTR {
		return node.ty.ptr_to.size
	}
	return 1
}

func gen_pre_inc(node *Node, num int) int {
	addr := gen_lval(node.expr)
	val := nreg
	nreg++
	load(node, val, addr)
	add(IR_ADD_IMM, val, num*get_inc_scale(node))
	store(node, addr, val)
	kill(addr)
	return val
}

func gen_post_inc(node *Node, num int) int {
	val := gen_pre_inc(node, num)
	add(IR_SUB_IMM, val, num*get_inc_scale(node))
	return val
}

func gen_stmt(node *Node) {
	switch node.op {
	case ND_NULL:
		return

	case ND_VARDEF:
		{
			if node.init == nil {
				return
			}
			rhs := gen_expr(node.init)
			lhs := nreg
			nreg++
			add(IR_BPREL, lhs, node.offset)
			store(node, lhs, rhs)
			kill(lhs)
			kill(rhs)
			return
		}
	case ND_IF:
		{
			if node.els != nil {
				x := nlabel
				nlabel++
				y := nlabel
				nlabel++
				r := gen_expr(node.cond)
				add(IR_UNLESS, r, x)
				kill(r)
				gen_stmt(node.then)
				jmp(y)
				label(x)
				gen_stmt(node.els)
				label(y)
				return
			}
			x := nlabel
			nlabel++
			r := gen_expr(node.cond)
			add(IR_UNLESS, r, x)
			kill(r)
			gen_stmt(node.then)
			label(x)
			return
		}
	case ND_FOR:
		{
			x := nlabel
			nlabel++
			y := nlabel
			nlabel++
			orig := break_label
			break_label = nlabel
			nlabel++

			gen_stmt(node.init)
			label(x)
			if node.cond != nil {
				r := gen_expr(node.cond)
				add(IR_UNLESS, r, y)
				kill(r)
			}
			gen_stmt(node.body)
			if node.inc != nil {
				gen_stmt(node.inc)
			}
			jmp(x)
			label(y)
			label(break_label)
			break_label = orig
			return
		}
	case ND_DO_WHILE:
		{
			x := nlabel
			nlabel++
			orig := break_label
			break_label = nlabel
			nlabel++
			label(x)
			gen_stmt(node.body)
			r := gen_expr(node.cond)
			add(IR_IF, r, x)
			kill(r)
			label(break_label)
			break_label = orig
			return
		}
	case ND_BREAK:
		if break_label == 0 {
			error("stray 'break' statement")
		}
		jmp(break_label)
	case ND_RETURN:
		{
			r := gen_expr(node.expr)

			// Statement expression (GNU extension)
			if return_label != 0 {
				add(IR_MOV, return_reg, r)
				kill(r)
				jmp(return_label)
				return
			}

			add(IR_RETURN, r, -1)
			kill(r)
			return
		}
	case ND_EXPR_STMT:
		{
			kill(gen_expr(node.expr))
			return
		}
	case ND_COMP_STMT:
		{
			for i := 0; i < node.stmts.len; i++ {
				gen_stmt((node.stmts.data[i]).(*Node))
			}
			return
		}
	default:
		error("unknown node: %d", node.op)
	}
}

func gen_ir(nodes *Vector) *Vector {
	v := new_vec()
	nlabel = 1

	for i := 0; i < nodes.len; i++ {
		node := nodes.data[i].(*Node)

		if node.op == ND_VARDEF {
			continue
		}

		//assert(node.op == ND_FUNC)
		code = new_vec()

		for i := 0; i < node.args.len; i++ {
			arg := node.args.data[i].(*Node)
			add(store_arg_insn(arg), arg.offset, i)
		}

		gen_stmt(node.body)

		fn := new(Function)
		fn.name = node.name
		fn.stacksize = node.stacksize
		fn.ir = code
		fn.globals = node.globals
		vec_push(v, fn)
	}
	return v
}

// [Debug] intermediate reprensations
func print_irs(fns *Vector) {
	if !debug {
		return
	}
	fmt.Println("-- intermediate reprensetations --")
	for i := 0; i < fns.len; i++ {
		fn := fns.data[i].(*Function)
		for j := 0; j < fn.ir.len; j++ {
			ir := fn.ir.data[j].(*IR)
			op := ""
			switch ir.op {
			case IR_IMM:
				op = "IR_IMM      "
			case IR_BPREL:
				op = "IR_BPREL    "
			case IR_MOV:
				op = "IR_MOV      "
			case IR_RETURN:
				op = "IR_RETURN   "
			case IR_LABEL:
				op = "IR_LABEL    "
			case IR_LT:
				op = "IR_LT       "
			case IR_JMP:
				op = "IR_JMP      "
			case IR_UNLESS:
				op = "IR_UNLESS   "
			case IR_LOAD:
				op = "IR_LOAD     "
			case IR_STORE:
				op = "IR_STORE    "
			case IR_STORE32_ARG:
				op = "IR_STORE32_ARG  "
			case IR_STORE64_ARG:
				op = "IR_STORE64_ARG  "
			case IR_KILL:
				op = "IR_KILL     "
			case IR_NOP:
				op = "IR_NOP      "
			case IR_ADD:
				op = "IR_ADD      "
			case IR_SUB:
				op = "IR_SUB      "
			case IR_MUL:
				op = "IR_MUL      "
			case IR_DIV:
				op = "IR_DIV      "
			default:
				op = "            "
			}
			fmt.Printf("[%02d:%02d] op: %s, lhs: %d, rhs: %d\n", i, j, op, ir.lhs, ir.rhs)
		}
	}
	fmt.Println("")
}
