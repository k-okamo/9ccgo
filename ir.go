package main

import (
	"fmt"
	"os"
)

var (
	regno   int
	basereg int
	bpoff   int
	label   int
	code    *Vector
	vars    *Map
)

var irinfo = []IRInfo{
	{op: '+', name: "+", ty: IR_TY_REG_REG},
	{op: '-', name: "-", ty: IR_TY_REG_REG},
	{op: '*', name: "*", ty: IR_TY_REG_REG},
	{op: '/', name: "/", ty: IR_TY_REG_REG},
	{op: IR_IMM, name: "MOV", ty: IR_TY_REG_IMM},
	{op: IR_ADD_IMM, name: "ADD", ty: IR_TY_REG_IMM},
	{op: IR_MOV, name: "MOV", ty: IR_TY_REG_REG},
	{op: IR_LABEL, name: "", ty: IR_LABEL},
	{op: IR_UNLESS, name: "UNLESS", ty: IR_TY_REG_LABEL},
	{op: IR_RETURN, name: "RET", ty: IR_TY_REG},
	{op: IR_ALLOCA, name: "ALLOCA", ty: IR_TY_REG_IMM},
	{op: IR_LOAD, name: "LOAD", ty: IR_TY_REG_REG},
	{op: IR_STORE, name: "STORE", ty: IR_TY_REG_REG},
	{op: IR_KILL, name: "KILL", ty: IR_TY_REG},
	{op: IR_NOP, name: "NOP", ty: IR_TY_NOARG},
	{op: 0, name: "", ty: 0},
}

const (
	IR_IMM = iota + 256
	IR_ADD_IMM
	IR_MOV
	IR_RETURN
	IR_LABEL
	IR_UNLESS
	IR_ALLOCA
	IR_LOAD
	IR_STORE
	IR_KILL
	IR_NOP
)

const (
	IR_TY_NOARG = iota + 256
	IR_TY_REG
	IR_TY_LABEL
	IR_TY_REG_REG
	IR_TY_REG_IMM
	IR_TY_REG_LABEL
)

type IR struct {
	op  int
	lhs int
	rhs int
}

type IRInfo struct {
	op   int
	name string
	ty   int
}

func get_irinfo(ir *IR) IRInfo {

	for _, info := range irinfo {
		if info.op == ir.op {
			return info
		}
	}
	// asset(0 && "invalid instruction")
	return IRInfo{op: 0, name: "", ty: 0}
}

func tostr(ir *IR) string {
	info := get_irinfo(ir)
	switch info.ty {
	case IR_TY_LABEL:
		return format("%s:\n", ir.lhs)
	case IR_TY_REG:
		return format("%s r%d\n", info.name, ir.lhs)
	case IR_TY_REG_REG:
		return format("%s r%d, r%d\n", info.name, ir.lhs, ir.rhs)
	case IR_TY_REG_IMM:
		return format("%s r%d, %d\n", info.name, ir.lhs, ir.rhs)
	case IR_TY_REG_LABEL:
		return format("%s r%d, .L%s\n", info.name, ir.lhs, ir.rhs)
	default:
		//asset(info.ty == IR_TY_NOARG)
		return format("%s\n", info.name)
	}
	return ""
}

func dump_ir(irv *Vector) {
	for i := 0; i < irv.len; i++ {
		fmt.Fprintf(os.Stderr, "%s", tostr(irv.data[i].(*IR)))
	}
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

	if node.ty == ND_IF {
		r := gen_expr(node.cond)
		x := label
		label++
		add(IR_UNLESS, r, x)
		add(IR_KILL, r, -1)
		gen_stmt(node.then)
		add(IR_LABEL, x, -1)
		return
	}
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
	label = 0

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
		case IR_LABEL:
			op = "IR_LABEL  "
		case IR_UNLESS:
			op = "IR_UNLESS "
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
