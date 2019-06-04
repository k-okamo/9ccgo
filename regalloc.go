package main

var (
	regs    = []string{"r10", "r11", "rbx", "r12", "r13", "r14", "r15"}
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

func visit(irv *Vector) {
	for i := 0; i < irv.len; i++ {
		ir := irv.data[i].(*IR)
		info := get_irinfo(ir)

		switch info.ty {
		case IR_TY_REG, IR_TY_REG_IMM, IR_TY_REG_LABEL:
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
			kill(reg_map[ir.lhs])
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
