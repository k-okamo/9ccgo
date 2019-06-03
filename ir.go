package main

import (
	"fmt"
)

var (
	regno   int
	basereg int
	bpoff   int
	code    *Vector
	vars    *Map
)

const (
	IR_IMM = iota
	IR_ADD_IMM
	IR_MOV
	IR_RETURN
	IR_ALLOCA
	IR_LOAD
	IR_STORE
	IR_KILL
	IR_NOP
)

type IR struct {
	op  int
	lhs int
	rhs int
}

func add(op, lhs, rhs int) *IR {
	ir := new(IR)
	ir.op = op
	ir.lhs = lhs
	ir.rhs = rhs
	vec_push(code, ir)
	return ir
}

func gen_lval(node *Node) int {
	if node.ty != ND_IDENT {
		error("not a lvalue")
	}

	if !map_exists(vars, node.name) {
		map_put(vars, node.name, bpoff)
		bpoff += 8
	}

	r := regno
	regno++
	off := map_get(vars, node.name).(int)
	add(IR_MOV, r, basereg)
	add(IR_ADD_IMM, r, off)
	return r
}

func gen_expr(node *Node) int {

	if node.ty == ND_NUM {
		r := regno
		regno++
		add(IR_IMM, r, node.val)
		return r
	}

	if node.ty == ND_IDENT {
		r := gen_lval(node)
		add(IR_LOAD, r, r)
		return r
	}

	if node.ty == '=' {
		rhs, lhs := gen_expr(node.rhs), gen_lval(node.lhs)
		add(IR_STORE, lhs, rhs)
		add(IR_KILL, rhs, -1)
		return lhs
	}
	// assert(strche("+-*/", node.ty))

	lhs, rhs := gen_expr(node.lhs), gen_expr(node.rhs)

	add(node.ty, lhs, rhs)
	add(IR_KILL, rhs, -1)
	return lhs
}

func gen_stmt(node *Node) {
	if node.ty == ND_RETURN {
		r := gen_expr(node.expr)
		add(IR_RETURN, r, -1)
		add(IR_KILL, r, -1)
		return
	}
	if node.ty == ND_EXPR_STMT {
		r := gen_expr(node.expr)
		add(IR_KILL, r, -1)
		return
	}
	if node.ty == ND_COMP_STMT {
		for i := 0; i < node.stmts.len; i++ {
			gen_stmt((node.stmts.data[i]).(*Node))
		}
		return
	}
	error("unknown node: %d", node.ty)
}

func gen_ir(node *Node) *Vector {
	// assert(node.ty == ND_COMP_STMT)
	code = new_vec()
	regno = 1
	basereg = 0
	vars = new_map()
	bpoff = 0

	alloca := add(IR_ALLOCA, basereg, -1)
	gen_stmt(node)
	alloca.rhs = bpoff
	add(IR_KILL, basereg, -1)
	return code
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
			op = "IR_IMM    "
		case IR_ADD_IMM:
			op = "IR_ADD_IMM"
		case IR_MOV:
			op = "IR_MOV    "
		case IR_RETURN:
			op = "IR_RETURN "
		case IR_ALLOCA:
			op = "IR_ALLOCA "
		case IR_LOAD:
			op = "IR_LOAD   "
		case IR_STORE:
			op = "IR_STORE  "
		case IR_KILL:
			op = "IR_KILL   "
		case IR_NOP:
			op = "IR_NOP    "
		case '+':
			op = "+         "
		case '-':
			op = "-         "
		default:
			op = "          "
		}
		fmt.Printf("[%02d] op: %s, lhs: %d, rhs: %d\n", i, op, ir.lhs, ir.rhs)
	}
	fmt.Println("")
}
