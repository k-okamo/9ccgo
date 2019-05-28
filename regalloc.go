package main

var (
	regs    = []string{"rdi", "rsi", "r10", "r11", "r12", "r13", "r14", "r15"}
	used    [8]bool
	reg_map []int
)

func alloc(ir_reg int) int {
	if reg_map[ir_reg] != -1 {
		r := reg_map[ir_reg]
		//assert(used[r])
		return r
	}

	for i := 0; i < len(regs); i++ {
		if used[i] == true {
			continue
		}
		used[i] = true
		reg_map[ir_reg] = i
		return i
	}
	error("register exhausted")
	return -1
}

func kill(r int) {
	//assert(used[r])
	used[r] = false
}

func alloc_regs(irv *Vector) {

	reg_map = make([]int, irv.len)
	for i := range reg_map {
		reg_map[i] = -1
	}

	for i := 0; i < irv.len; i++ {
		ir := irv.data[i].(*IR)

		switch ir.op {
		case IR_IMM:
			ir.lhs = alloc(ir.lhs)
		case IR_MOV, '+', '-', '*':
			ir.lhs = alloc(ir.lhs)
			ir.rhs = alloc(ir.rhs)
		case IR_RETURN:
			kill(reg_map[ir.lhs])
		case IR_KILL:
			kill(reg_map[ir.lhs])
			ir.op = IR_NOP
		default:
			//assert(0&& "unknown operator")
		}
	}
}
