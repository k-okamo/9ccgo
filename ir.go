package main

import (
	"fmt"
	"os"
)

var (
	code      *Vector
	vars      *Map
	regno     int
	stacksize int
	label     int
)

var irinfo = []IRInfo{
	{op: IR_ADD, name: "ADD", ty: IR_TY_REG_REG},
	{op: IR_SUB, name: "SUB", ty: IR_TY_REG_REG},
	{op: IR_MUL, name: "MUL", ty: IR_TY_REG_REG},
	{op: IR_DIV, name: "DIV", ty: IR_TY_REG_REG},
	{op: IR_IMM, name: "MOV", ty: IR_TY_REG_IMM},
	{op: IR_SUB_IMM, name: "SUB", ty: IR_TY_REG_IMM},
	{op: IR_MOV, name: "MOV", ty: IR_TY_REG_REG},
	{op: IR_LABEL, name: "", ty: IR_TY_LABEL},
	{op: IR_JMP, name: "JMP", ty: IR_TY_LABEL},
	{op: IR_UNLESS, name: "UNLESS", ty: IR_TY_REG_LABEL},
	{op: IR_CALL, name: "CALL", ty: IR_TY_CALL},
	{op: IR_RETURN, name: "RET", ty: IR_TY_REG},
	{op: IR_LOAD, name: "LOAD", ty: IR_TY_REG_REG},
	{op: IR_STORE, name: "STORE", ty: IR_TY_REG_REG},
	{op: IR_KILL, name: "KILL", ty: IR_TY_REG},
	{op: IR_SAVE_ARGS, name: "SAVE_ARGS", ty: IR_TY_IMM},
	{op: IR_NOP, name: "NOP", ty: IR_TY_NOARG},
	{op: 0, name: "", ty: 0},
}

const (
	IR_ADD = iota + 256
	IR_SUB
	IR_MUL
	IR_DIV
	IR_IMM
	IR_SUB_IMM
	IR_MOV
	IR_RETURN
	IR_CALL
	IR_LABEL
	IR_JMP
	IR_UNLESS
	IR_LOAD
	IR_STORE
	IR_KILL
	IR_SAVE_ARGS
	IR_NOP
)

const (
	IR_TY_NOARG = iota + 256
	IR_TY_REG
	IR_TY_IMM
	IR_TY_LABEL
	IR_TY_REG_REG
	IR_TY_REG_IMM
	IR_TY_REG_LABEL
	IR_TY_CALL
)

type IR struct {
	op  int
	lhs int
	rhs int

	// Function call
	name  string
	nargs int
	args  [6]int
}

type IRInfo struct {
	op   int
	name string
	ty   int
}

type Function struct {
	name      string
	stacksize int
	ir        *Vector
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
		return format(".L%d:\n", ir.lhs)
	case IR_TY_IMM:
		return format("%s %d\n", info.name, ir.lhs)
	case IR_TY_REG:
		return format("%s r%d\n", info.name, ir.lhs)
	case IR_TY_REG_REG:
		return format("%s r%d, r%d\n", info.name, ir.lhs, ir.rhs)
	case IR_TY_REG_IMM:
		return format("%s r%d, %d\n", info.name, ir.lhs, ir.rhs)
	case IR_TY_REG_LABEL:
		return format("%s r%d, .L%d\n", info.name, ir.lhs, ir.rhs)
	case IR_TY_CALL:
		{
			sb := new_sb()
			sb_append(sb, format("r%d = %s(", ir.lhs, ir.name))
			for i := 0; i < ir.nargs; i++ {
				sb_append(sb, format(", r%d", ir.args))
			}
			sb_append(sb, ")\n")
			return sb_get(sb)
		}
	default:
		//asset(info.ty == IR_TY_NOARG)
		return format("%s\n", info.name)
	}
	return ""
}

func dump_ir(irv *Vector) {
	for i := 0; i < irv.len; i++ {
		fn := irv.data[i].(*Function)
		fmt.Fprintf(os.Stderr, "%s():\n", fn.name)
		for j := 0; j < fn.ir.len; j++ {
			fmt.Fprintf(os.Stderr, " %s", tostr(fn.ir.data[j].(*IR)))
		}
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
		stacksize += 8
		map_put(vars, node.name, stacksize)
	}

	r := regno
	regno++
	off := map_get(vars, node.name).(int)
	add(IR_MOV, r, 0)
	add(IR_SUB_IMM, r, off)
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

	if node.ty == ND_CALL {
		var args [6]int
		for i := 0; i < node.args.len; i++ {
			args[i] = gen_expr(node.args.data[i].(*Node))
		}
		r := regno
		regno++

		ir := add(IR_CALL, r, -1)
		ir.name = node.name
		ir.nargs = node.args.len
		for i := 0; i < 6; i++ {
			ir.args[i] = args[i]
		}
		for i := 0; i < ir.nargs; i++ {
			add(IR_KILL, ir.args[i], -1)
		}
		return r
	}

	if node.ty == '=' {
		rhs, lhs := gen_expr(node.rhs), gen_lval(node.lhs)
		add(IR_STORE, lhs, rhs)
		add(IR_KILL, rhs, -1)
		return lhs
	}
	// assert(strche("+-*/", node.ty))

	var ty int
	if node.ty == '+' {
		ty = IR_ADD
	} else if node.ty == '-' {
		ty = IR_SUB
	} else if node.ty == '*' {
		ty = IR_MUL
	} else {
		ty = IR_DIV
	}

	lhs, rhs := gen_expr(node.lhs), gen_expr(node.rhs)

	add(ty, lhs, rhs)
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

		if node.els == nil {
			add(IR_LABEL, x, -1)
			return
		}

		y := label
		label++
		add(IR_JMP, y, -1)
		add(IR_LABEL, x, -1)
		gen_stmt(node.els)
		add(IR_LABEL, y, -1)
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

func gen_args(nodes *Vector) {
	if nodes.len == 0 {
		return
	}

	add(IR_SAVE_ARGS, nodes.len, -1)

	for i := 0; i < nodes.len; i++ {
		node := nodes.data[i].(*Node)
		if node.ty != ND_IDENT {
			error("bad parameter")
		}

		stacksize += 8
		map_put(vars, node.name, stacksize)
	}
}

func gen_ir(nodes *Vector) *Vector {
	v := new_vec()

	for i := 0; i < nodes.len; i++ {
		node := nodes.data[i].(*Node)
		//assert(node.ty == ND_FUNC)

		code = new_vec()
		vars = new_map()
		regno = 1
		stacksize = 0

		gen_args(node.args)
		gen_stmt(node.body)

		fn := new(Function)
		fn.name = node.name
		fn.stacksize = stacksize
		fn.ir = code
		vec_push(v, fn)
	}
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
			op = "IR_IMM      "
		case IR_SUB_IMM:
			op = "IR_SUB_IMM  "
		case IR_MOV:
			op = "IR_MOV      "
		case IR_RETURN:
			op = "IR_RETURN   "
		case IR_LABEL:
			op = "IR_LABEL    "
		case IR_JMP:
			op = "IR_JMP      "
		case IR_UNLESS:
			op = "IR_UNLESS   "
		case IR_LOAD:
			op = "IR_LOAD     "
		case IR_STORE:
			op = "IR_STORE    "
		case IR_KILL:
			op = "IR_KILL     "
		case IR_SAVE_ARGS:
			op = "IR_SAVE_ARGS"
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
		fmt.Printf("[%02d] op: %s, lhs: %d, rhs: %d\n", i, op, ir.lhs, ir.rhs)
	}
	fmt.Println("")
}
