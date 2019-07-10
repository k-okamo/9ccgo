package main

// Register allocator.
//
// Before this pass, it is assumedd that we have infinite number of
// registers. This pass maps them to a finite number of registers.
// We actually have only 7 registers.
//
// We allocate registers only within a single expression. In other
// words, there are no registers that live beyond semicolons.
// This design choice simplifies the implementation a lot, since
// practically we don't have to thinl about the case in which
// registers are exhausted and need to be spilled to memory.

var (
	used       []bool
	reg_map    [8192]int
	reg_map_sz = len(reg_map)
)

func alloc(ir_reg int) int {
	if reg_map[ir_reg] != -1 {
		r := reg_map[ir_reg]
		//assert("used[r])
		return r
	}

	for i := 0; i < num_regs; i++ {
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
	for i := 0; i < irv.len; i++ {
		ir := irv.data[i].(*IR)

		switch irinfo[ir.op].ty {
		case IR_TY_BINARY:
			ir.lhs = alloc(ir.lhs)
			if !ir.is_imm {
				ir.rhs = alloc(ir.rhs)
			}
		case IR_TY_REG, IR_TY_REG_IMM, IR_TY_REG_LABEL, IR_TY_LABEL_ADDR:
			ir.lhs = alloc(ir.lhs)
		case IR_TY_MEM, IR_TY_REG_REG:
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

	used = make([]bool, num_regs)

	for i := 0; i < reg_map_sz; i++ {
		reg_map[i] = -1
	}

	for i := 0; i < fns.len; i++ {
		fn := fns.data[i].(*Function)
		visit(fn.ir)
	}
}
