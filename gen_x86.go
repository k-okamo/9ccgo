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

func emit(format string, a ...interface{}) {
	fmt.Printf("\t"+format+"\n", a...)
}

func emit_cmp(ir *IR, insn string) {
	emit("cmp %s, %s", regs[ir.lhs], regs[ir.rhs])
	emit("%s %s", insn, regs8[ir.lhs])
	emit("movzb %s, %s", regs[ir.lhs], regs8[ir.lhs])
}

func reg(r, size int) string {
	if size == 1 {
		return regs8[r]
	}
	if size == 4 {
		return regs32[r]
	}
	// assert(size == 8)
	return regs[r]
}

func gen(fn *Function) {

	ret := format(".Lend%d", glabel)
	glabel++

	fmt.Printf(".global %s\n", fn.name)
	fmt.Printf("%s:\n", fn.name)
	emit("push rbp")
	emit("mov rbp, rsp")
	emit("sub rsp, %d", roundup(fn.stacksize, 16))
	emit("push r12")
	emit("push r13")
	emit("push r14")
	emit("push r15")

	for i := 0; i < fn.ir.len; i++ {
		ir := fn.ir.data[i].(*IR)
		lhs := ir.lhs
		rhs := ir.rhs

		switch ir.op {
		case IR_IMM:
			emit("mov %s, %d", regs[lhs], rhs)
		case IR_BPREL:
			emit("lea %s, [rbp-%d]", regs[lhs], rhs)
		case IR_MOV:
			emit("mov %s, %s", regs[lhs], regs[rhs])
		case IR_RETURN:
			emit("mov rax, %s", regs[lhs])
			emit("jmp %s", ret)
		case IR_CALL:
			{
				for i := 0; i < ir.nargs; i++ {
					emit("mov %s, %s", argreg64[i], regs[ir.args[i]])
				}
				emit("push r10")
				emit("push r11")
				emit("mov rax, 0")
				emit("call %s", ir.name)
				emit("pop r11")
				emit("pop r10")
				emit("mov %s, rax", regs[lhs])
			}
		case IR_LABEL:
			fmt.Printf(".L%d:\n", lhs)
		case IR_LABEL_ADDR:
			emit("lea %s, %s", regs[lhs], ir.name)
		case IR_NEG:
			emit("neg %s", regs[lhs])
		case IR_EQ:
			emit_cmp(ir, "sete")
		case IR_NE:
			emit_cmp(ir, "setne")
		case IR_LT:
			emit_cmp(ir, "setl")
		case IR_LE:
			emit_cmp(ir, "setle")
		case IR_AND:
			emit("and %s, %s", regs[lhs], regs[rhs])
		case IR_OR:
			emit("or %s, %s", regs[lhs], regs[rhs])
		case IR_XOR:
			emit("xor %s, %s", regs[lhs], regs[rhs])
		case IR_SHL:
			emit("mov cl, %s", regs8[rhs])
			emit("shl %s, cl", regs[lhs])
		case IR_SHR:
			emit("mov cl, %s", regs8[rhs])
			emit("shr %s, cl", regs[lhs])
		case IR_JMP:
			emit("jmp .L%d", lhs)
		case IR_IF:
			emit("cmp %s, 0", regs[lhs])
			emit("jne .L%d", rhs)
		case IR_UNLESS:
			emit("cmp %s, 0", regs[lhs])
			emit("je .L%d", rhs)
		case IR_LOAD:
			emit("mov %s, [%s]", reg(lhs, ir.size), regs[rhs])
			if ir.size == 1 {
				emit("movzb %s, %s", regs[lhs], regs8[lhs])
			}
		case IR_STORE:
			emit("mov [%s], %s", regs[lhs], reg(rhs, ir.size))
		case IR_STORE8_ARG:
			emit("mov [rbp-%d], %s", lhs, argreg8[rhs])
		case IR_STORE32_ARG:
			emit("mov [rbp-%d], %s", lhs, argreg32[rhs])
		case IR_STORE64_ARG:
			emit("mov [rbp-%d], %s", lhs, argreg64[rhs])
		case IR_ADD:
			emit("add %s, %s", regs[lhs], regs[rhs])
		case IR_ADD_IMM:
			emit("add %s, %d", regs[lhs], rhs)
		case IR_SUB:
			emit("sub %s, %s", regs[lhs], regs[rhs])
		case IR_SUB_IMM:
			emit("sub %s, %d", regs[lhs], rhs)
		case IR_MUL:
			emit("mov rax, %s", regs[rhs])
			emit("mul %s", regs[lhs])
			emit("mov %s, rax", regs[lhs])
		case IR_MUL_IMM:
			if rhs < 256 && popcount(uint(rhs)) == 1 {
				emit("shl %s, %d", regs[lhs], ctz(uint(rhs)))
				break
			}
			emit("mov rax, %d", rhs)
			emit("mul %s", regs[lhs])
			emit("mov %s, rax", regs[lhs])
		case IR_DIV:
			emit("mov rax, %s", regs[lhs])
			emit("cqo")
			emit("div %s", regs[rhs])
			emit("mov %s, rax", regs[lhs])
		case IR_MOD:
			emit("mov rax, %s", regs[lhs])
			emit("cqo")
			emit("div %s", regs[rhs])
			emit("mov %s, rdx", regs[lhs])
		case IR_NOP:
			break
		default:
			//assert(0 && "unknown operator")
		}
	}

	fmt.Printf("%s:\n", ret)
	emit("pop r15")
	emit("pop r14")
	emit("pop r13")
	emit("pop r12")
	emit("mov rsp, rbp")
	emit("pop rbp")
	emit("ret")
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
		emit(".ascii \"%s\"", escape(v.data, v.len))
	}

	fmt.Printf(".text\n")
	for i := 0; i < fns.len; i++ {
		gen(fns.data[i].(*Function))
	}
}
