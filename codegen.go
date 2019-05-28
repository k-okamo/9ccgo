package main

import (
	"fmt"
)

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
