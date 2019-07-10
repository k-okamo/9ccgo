package main

import (
	"fmt"
	"os"
)

var irinfo = map[int]IRInfo{
	IR_ADD:        {name: "ADD", ty: IR_TY_BINARY},
	IR_CALL:       {name: "CALL", ty: IR_TY_CALL},
	IR_DIV:        {name: "DIV", ty: IR_TY_REG_REG},
	IR_IMM:        {name: "IMM", ty: IR_TY_REG_IMM},
	IR_JMP:        {name: "JMP", ty: IR_TY_JMP},
	IR_KILL:       {name: "KILL", ty: IR_TY_REG},
	IR_LABEL:      {name: "", ty: IR_TY_LABEL},
	IR_LABEL_ADDR: {name: "LABEL_ADDR", ty: IR_TY_LABEL_ADDR},
	IR_EQ:         {name: "EQ", ty: IR_TY_REG_REG},
	IR_NE:         {name: "NE", ty: IR_TY_REG_REG},
	IR_LE:         {name: "LE", ty: IR_TY_REG_REG},
	IR_LT:         {name: "LT", ty: IR_TY_REG_REG},
	IR_AND:        {name: "AND", ty: IR_TY_REG_REG},
	IR_OR:         {name: "OR", ty: IR_TY_REG_REG},
	IR_XOR:        {name: "XOR", ty: IR_TY_BINARY},
	IR_SHL:        {name: "SHL", ty: IR_TY_REG_REG},
	IR_SHR:        {name: "SHR", ty: IR_TY_REG_REG},
	IR_LOAD:       {name: "LOAD", ty: IR_TY_MEM},
	IR_MOD:        {name: "MOD", ty: IR_TY_REG_REG},
	IR_NEG:        {name: "NEG", ty: IR_TY_REG},
	IR_MOV:        {name: "MOV", ty: IR_TY_REG_REG},
	IR_MUL:        {name: "MUL", ty: IR_TY_BINARY},
	IR_NOP:        {name: "NOP", ty: IR_TY_NOARG},
	IR_RETURN:     {name: "RET", ty: IR_TY_REG},
	IR_STORE:      {name: "STORE", ty: IR_TY_MEM},
	IR_STORE_ARG:  {name: "STORE_ARG", ty: IR_TY_STORE_ARG},
	IR_SUB:        {name: "SUB", ty: IR_TY_BINARY},
	IR_BPREL:      {name: "BPREL", ty: IR_TY_REG_IMM},
	IR_IF:         {name: "IF", ty: IR_TY_REG_LABEL},
	IR_UNLESS:     {name: "UNLESS", ty: IR_TY_REG_LABEL},
	0:             {name: "", ty: 0},
}

func tostr(ir *IR) string {
	info := irinfo[ir.op]
	switch info.ty {
	case IR_TY_BINARY:
		if ir.is_imm {
			return format("\t%s r%d, %d", info.name, ir.lhs, ir.rhs)
		}
		return format("\t%s r%d, r%d", info.name, ir.lhs, ir.rhs)
	case IR_TY_LABEL:
		return format(".L%d:", ir.lhs)
	case IR_TY_LABEL_ADDR:
		return format("\t%s r%d, %s", info.name, ir.lhs, ir.name)
	case IR_TY_IMM:
		return format("\t%s %d", info.name, ir.lhs)
	case IR_TY_REG:
		return format("\t%s r%d", info.name, ir.lhs)
	case IR_TY_JMP:
		return format("\t%s r%d", info.name, ir.lhs)
	case IR_TY_REG_REG:
		return format("\t%s r%d, r%d", info.name, ir.lhs, ir.rhs)
	case IR_TY_MEM:
		return format("\t%s%d r%d, r%d", info.name, ir.size, ir.lhs, ir.rhs)
	case IR_TY_REG_IMM:
		return format("\t%s r%d, %d", info.name, ir.lhs, ir.rhs)
	case IR_TY_STORE_ARG:
		return format("\t%s%d %d, %d", info.name, ir.size, ir.lhs, ir.rhs)
	case IR_TY_REG_LABEL:
		return format("\t%s r%d, .L%d", info.name, ir.lhs, ir.rhs)
	case IR_TY_CALL:
		{
			sb := new_sb()
			sb_append(sb, format("r%d = %s(", ir.lhs, ir.name))
			for i := 0; i < ir.nargs; i++ {
				if i != 0 {
					sb_append(sb, ", ")
				}
				sb_append(sb, format("r%d", ir.args[i]))
			}
			sb_append(sb, ")\n")
			return sb_get(sb)
		}
	default:
		//asset(info.ty == IR_TY_NOARG)
		return format("\t%s", info.name)
	}
	return ""
}

func dump_ir(irv *Vector) {
	for i := 0; i < irv.len; i++ {
		fn := irv.data[i].(*Function)
		fmt.Fprintf(os.Stderr, "%s():\n", fn.name)
		for j := 0; j < fn.ir.len; j++ {
			fmt.Fprintf(os.Stderr, "%s\n", tostr(fn.ir.data[j].(*IR)))
		}
	}
}
