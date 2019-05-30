package main

import (
	"fmt"
)

var (
	regno int
	code  *Vector
)

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

func add(op, lhs, rhs int) *IR {
	ir := new(IR)
	ir.op = op
	ir.lhs = lhs
	ir.rhs = rhs
	vec_push(code, ir)
	return ir
}

func gen_expr(node *Node) int {

	if node.ty == ND_NUM {
		r := regno
		regno++
		add(IR_IMM, r, node.val)
		return r
	}
	// assert(strche("+-*/", node.ty))

	lhs, rhs := gen_expr(node.lhs), gen_expr(node.rhs)

	add(node.ty, lhs, rhs)
	add(IR_KILL, rhs, 0)
	return lhs
}

func gen_stmt(node *Node) {
	if node.ty == ND_RETURN {
		r := gen_expr(node.expr)
		add(IR_RETURN, r, 0)
		add(IR_KILL, r, 0)
		return
	}
	if node.ty == ND_EXPR_STMT {
		r := gen_expr(node.expr)
		add(IR_KILL, r, 0)
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
	gen_stmt(node)
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
