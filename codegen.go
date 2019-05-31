package main

import (
	"fmt"
)

var (
	n int
)

func gen_label() string {
	buf := fmt.Sprintf(".L%d", n)
	n++
	return buf
}

func gen_X86(irv *Vector) {

	ret := gen_label()
	fmt.Printf("\tpush rbp\n")
	fmt.Printf("\tmov rbp, rsp\n")

	for i := 0; i < irv.len; i++ {
		ir := irv.data[i].(*IR)

		switch ir.op {
		case IR_IMM:
			fmt.Printf("\tmov %s, %d\n", regs[ir.lhs], ir.rhs)
		case IR_MOV:
			fmt.Printf("\tmov %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_RETURN:
			fmt.Printf("\tmov rax, %s\n", regs[ir.lhs])
			//fmt.Printf("\tret\n")
			fmt.Printf("\tjmp %s\n", ret)
		case IR_ALLOCA:
			if ir.rhs != 0 {
				fmt.Printf("\tsub rsp, %d\n", ir.rhs)
			}
			fmt.Printf("\tmov %s, rsp\n", regs[ir.lhs])
		case IR_LOAD:
			fmt.Printf("\tmov %s, [%s]\n", regs[ir.lhs], regs[ir.rhs])
		case IR_STORE:
			fmt.Printf("\tmov [%s], %s\n", regs[ir.lhs], regs[ir.rhs])
		case '+':
			fmt.Printf("\tadd %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case '-':
			fmt.Printf("\tsub %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case '*':
			fmt.Printf("\tmov rax, %s\n", regs[ir.rhs])
			fmt.Printf("\tmul %s\n", regs[ir.lhs])
			fmt.Printf("\tmov %s, rax\n", regs[ir.lhs])
		case '/':
			fmt.Printf("\tmov rax, %s\n", regs[ir.lhs])
			fmt.Printf("\tcqo\n")
			fmt.Printf("\tdiv %s\n", regs[ir.rhs])
			fmt.Printf("\tmov %s, rax\n", regs[ir.lhs])
		case IR_NOP:
			break
		default:
			//assert(0 && "unknown operator")
		}
	}

	fmt.Printf("%s:\n", ret)
	fmt.Printf("\tmov rsp, rbp\n")
	fmt.Printf("\tmov rsp, rbp\n")
	fmt.Printf("\tpop rbp\n")
	fmt.Printf("\tret\n")
}
