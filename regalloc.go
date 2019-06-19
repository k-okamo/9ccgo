package main

var (
	regs    = []string{"rbp", "r10", "r11", "rbx", "r12", "r13", "r14", "r15"}
	regs8   = []string{"rp1", "r10b", "r11b", "b1", "r12b", "r13b", "r14b", "r15b"}
	regs32  = []string{"ebp", "r10d", "r11d", "ebx", "r12d", "r13d", "r14d", "r15d"}
	used    [8]bool
	reg_map []int
)

func alloc(ir_reg int) int {
	if reg_map[ir_reg] != -1 {
		r := reg_map[ir_reg]
		//assert("used[r])
		return r
	}

	for i := 0; i < len(regs); i++ {
		if used[i] == true {
			continue
		}
		reg_map[ir_reg] = i
		used[i] = true
		return i
	}
	error("register exhausted")
	return -1
}

func visit(irv *Vector) {
	// r0 is a reserved register that is always mapped to rbp.
	reg_map[0] = 0
	used[0] = true

	for i := 0; i < irv.len; i++ {
		ir := irv.data[i].(*IR)

		switch irinfo[ir.op].ty {
		case IR_TY_REG, IR_TY_REG_IMM, IR_TY_REG_LABEL, IR_TY_LABEL_ADDR:
			ir.lhs = alloc(ir.lhs)
		case IR_TY_REG_REG:
			ir.lhs = alloc(ir.lhs)
			ir.rhs = alloc(ir.rhs)
		case IR_TY_CALL:
			ir.lhs = alloc(ir.lhs)
			for i := 0; i < ir.nargs; i++ {
				ir.args[i] = alloc(ir.args[i])
			}
		}

		if ir.op == IR_KILL {
			//assert(used[ir.lhs])
			used[ir.lhs] = false
			ir.op = IR_NOP
		}
	}
}

func alloc_regs(fns *Vector) {
	for i := 0; i < fns.len; i++ {
		fn := fns.data[i].(*Function)

		reg_map = make([]int, fn.ir.len)
		for j := 0; j < fn.ir.len; j++ {
			reg_map[j] = -1
		}
		visit(fn.ir)
	}
}
