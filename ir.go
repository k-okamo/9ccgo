package main

import (
	"fmt"
)

var (
	regno int
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

func new_ir(op, lhs, rhs int) *IR {
	ir := new(IR)
	ir.op = op
	ir.lhs = lhs
	ir.rhs = rhs
	return ir
}

func gen_ir_sub(v *Vector, node *Node) int {

	if node.ty == ND_NUM {
		r := regno
		regno++
		vec_push(v, new_ir(IR_IMM, r, node.val))
		return r
	}
	// assert(strche("+-*/", node.ty))

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
