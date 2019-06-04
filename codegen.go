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

func gen(fn *Function) {

	ret := format(".Lend%d", label)
	label++

	fmt.Printf(".global %s\n", fn.name)
	fmt.Printf("%s:\n", fn.name)
	fmt.Printf("\tpush r12\n")
	fmt.Printf("\tpush r13\n")
	fmt.Printf("\tpush r14\n")
	fmt.Printf("\tpush r15\n")
	fmt.Printf("\tpush rbp\n")
	fmt.Printf("\tmov rbp, rsp\n")

	for i := 0; i < fn.ir.len; i++ {
		ir := fn.ir.data[i].(*IR)

		switch ir.op {
		case IR_IMM:
			fmt.Printf("\tmov %s, %d\n", regs[ir.lhs], ir.rhs)
		case IR_ADD_IMM:
			fmt.Printf("\tadd %s, %d\n", regs[ir.lhs], ir.rhs)
		case IR_MOV:
			fmt.Printf("\tmov %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_RETURN:
			fmt.Printf("\tmov rax, %s\n", regs[ir.lhs])
			fmt.Printf("\tjmp %s\n", ret)
		case IR_CALL:
			{
				arg := []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
				for i := 0; i < ir.nargs; i++ {
					fmt.Printf("\tmov %s, %s\n", arg[i], regs[ir.args[i]])
				}
				fmt.Printf("\tpush r10\n")
				fmt.Printf("\tpush r11\n")
				fmt.Printf("\tmov rax, 0\n")
				fmt.Printf("\tcall %s\n", ir.name)
				fmt.Printf("\tmov %s, rax\n", regs[ir.lhs])
				fmt.Printf("\tpop r11\n")
				fmt.Printf("\tpop r10\n")
				fmt.Printf("\tmov %s, rax\n", regs[ir.lhs])
			}
		case IR_LABEL:
			fmt.Printf("\t.L%d:\n", ir.lhs)
		case IR_JMP:
			fmt.Printf("\tjmp .L%d\n", ir.lhs)
		case IR_UNLESS:
			fmt.Printf("\tcmp %s, 0\n", regs[ir.lhs])
			fmt.Printf("\tje .L%d\n", ir.rhs)
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
	fmt.Printf("\tpop rbp\n")
	fmt.Printf("\tpop r15\n")
	fmt.Printf("\tpop r14\n")
	fmt.Printf("\tpop r13\n")
	fmt.Printf("\tpop r12\n")
	fmt.Printf("\tret\n")
}

func gen_x86(fns *Vector) {
	fmt.Printf(".intel_syntax noprefix\n")

	for i := 0; i < fns.len; i++ {
		gen(fns.data[i].(*Function))
	}
}
