package main

// This pass generats x86-64 assembly from IR.

import (
	"fmt"
)

var (
	n        int
	glabel   int
	argreg8  = []string{"dil", "sil", "dl", "cl", "r8b", "r9b"}
	argreg32 = []string{"edi", "esi", "edx", "ecx", "r8d", "r9d"}
	argreg64 = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
)

func escape(s string, length int) string {

	if len(s) == 0 {
		return string([]rune{'\\', '0', '0', '0', '\\', '0', '0', '0', '\\', '0', '0', '0', '\\', '0', '0', '0'})
	}

	escaped := map[rune]rune{
		'\b': 'b',
		'\f': 'f',
		'\n': 'n',
		'\r': 'r',
		'\t': 't',
		'\\': '\\',
		'\'': '\'',
		'"':  '"',
	}

	sb := new_sb()
	for _, c := range s {
		esc, ok := escaped[c]
		if ok {
			sb_add(sb, "\\")
			sb_add(sb, string(esc))
		} else if isgraph(c) || c == ' ' {
			sb_add(sb, string(c))
		} else {
			sb_append(sb, format("\\%03o", c))
		}
	}

	buf := string([]rune{'\\', '0', '0', '0'})
	sb_append(sb, buf)
	return sb_get(sb)
}

func gen_label() string {
	buf := fmt.Sprintf(".L%d", n)
	n++
	return buf
}

func emit_cmp(ir *IR, insn string) {
	fmt.Printf("\tcmp %s, %s\n", regs[ir.lhs], regs[ir.rhs])
	fmt.Printf("\t%s %s\n", insn, regs8[ir.lhs])
	fmt.Printf("\tmovzb %s, %s\n", regs[ir.lhs], regs8[ir.lhs])
}

func gen(fn *Function) {

	ret := format(".Lend%d", glabel)
	glabel++

	fmt.Printf(".global %s\n", fn.name)
	fmt.Printf("%s:\n", fn.name)
	fmt.Printf("\tpush rbp\n")
	fmt.Printf("\tmov rbp, rsp\n")
	fmt.Printf("\tsub rsp, %d\n", roundup(fn.stacksize, 16))
	fmt.Printf("\tpush r12\n")
	fmt.Printf("\tpush r13\n")
	fmt.Printf("\tpush r14\n")
	fmt.Printf("\tpush r15\n")

	for i := 0; i < fn.ir.len; i++ {
		ir := fn.ir.data[i].(*IR)

		switch ir.op {
		case IR_IMM:
			fmt.Printf("\tmov %s, %d\n", regs[ir.lhs], ir.rhs)
		case IR_BPREL:
			fmt.Printf("\tlea %s, [rbp-%d]\n", regs[ir.lhs], ir.rhs)
		case IR_MOV:
			fmt.Printf("\tmov %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_RETURN:
			fmt.Printf("\tmov rax, %s\n", regs[ir.lhs])
			fmt.Printf("\tjmp %s\n", ret)
		case IR_CALL:
			{
				for i := 0; i < ir.nargs; i++ {
					fmt.Printf("\tmov %s, %s\n", argreg64[i], regs[ir.args[i]])
				}
				fmt.Printf("\tpush r10\n")
				fmt.Printf("\tpush r11\n")
				fmt.Printf("\tmov rax, 0\n")
				fmt.Printf("\tcall %s\n", ir.name)
				fmt.Printf("\tpop r11\n")
				fmt.Printf("\tpop r10\n")
				fmt.Printf("\tmov %s, rax\n", regs[ir.lhs])
			}
		case IR_LABEL:
			fmt.Printf(".L%d:\n", ir.lhs)
		case IR_LABEL_ADDR:
			fmt.Printf("\tlea %s, %s\n", regs[ir.lhs], ir.name)
		case IR_EQ:
			emit_cmp(ir, "sete")
		case IR_NE:
			emit_cmp(ir, "setne")
		case IR_LT:
			emit_cmp(ir, "setl")
		case IR_OR:
			fmt.Printf("\tor %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_JMP:
			fmt.Printf("\tjmp .L%d\n", ir.lhs)
		case IR_IF:
			fmt.Printf("\tcmp %s, 0\n", regs[ir.lhs])
			fmt.Printf("\tjne .L%d\n", ir.rhs)
		case IR_UNLESS:
			fmt.Printf("\tcmp %s, 0\n", regs[ir.lhs])
			fmt.Printf("\tje .L%d\n", ir.rhs)
		case IR_LOAD8:
			fmt.Printf("\tmov %s, [%s]\n", regs8[ir.lhs], regs[ir.rhs])
			fmt.Printf("\tmovzb %s, %s\n", regs[ir.lhs], regs8[ir.lhs])
		case IR_LOAD32:
			fmt.Printf("\tmov %s, [%s]\n", regs32[ir.lhs], regs[ir.rhs])
		case IR_LOAD64:
			fmt.Printf("\tmov %s, [%s]\n", regs[ir.lhs], regs[ir.rhs])
		case IR_STORE8:
			fmt.Printf("\tmov [%s], %s\n", regs[ir.lhs], regs8[ir.rhs])
		case IR_STORE32:
			fmt.Printf("\tmov [%s], %s\n", regs[ir.lhs], regs32[ir.rhs])
		case IR_STORE64:
			fmt.Printf("\tmov [%s], %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_STORE8_ARG:
			fmt.Printf("\tmov [rbp-%d], %s\n", ir.lhs, argreg8[ir.rhs])
		case IR_STORE32_ARG:
			fmt.Printf("\tmov [rbp-%d], %s\n", ir.lhs, argreg32[ir.rhs])
		case IR_STORE64_ARG:
			fmt.Printf("\tmov [rbp-%d], %s\n", ir.lhs, argreg64[ir.rhs])
		case IR_ADD:
			fmt.Printf("\tadd %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_SUB:
			fmt.Printf("\tsub %s, %s\n", regs[ir.lhs], regs[ir.rhs])
		case IR_MUL:
			fmt.Printf("\tmov rax, %s\n", regs[ir.rhs])
			fmt.Printf("\tmul %s\n", regs[ir.lhs])
			fmt.Printf("\tmov %s, rax\n", regs[ir.lhs])
		case IR_DIV:
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
	fmt.Printf("\tpop r15\n")
	fmt.Printf("\tpop r14\n")
	fmt.Printf("\tpop r13\n")
	fmt.Printf("\tpop r12\n")
	fmt.Printf("\tmov rsp, rbp\n")
	fmt.Printf("\tpop rbp\n")
	fmt.Printf("\tret\n")
}

func gen_x86(globals, fns *Vector) {

	fmt.Printf(".intel_syntax noprefix\n")

	fmt.Printf(".data\n")
	for i := 0; i < globals.len; i++ {
		v := globals.data[i].(*Var)
		if v.is_extern {
			continue
		}
		fmt.Printf("%s:\n", v.name)
		fmt.Printf("\t.ascii \"%s\"\n", escape(v.data, v.len))
	}

	fmt.Printf(".text\n")
	for i := 0; i < fns.len; i++ {
		gen(fns.data[i].(*Function))
	}
}
