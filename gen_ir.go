package main

// 9ccgo's code generation is two-pass. In the first pass, abstract
// syntax trees are compiled to IT (intermediate representation).
//
// IR resembles the real x86-64 instruction set, but it has infinite
// number of registers. We don't try too hard to reuse registers in
// this pass. Instead, we "kill" registers to mark them as dead when
// we are done with them and use new registers.
//
// Such infinite number of registers are mapped to a finite registers
// in a later pass.

var (
	code         *Vector
	nreg         = 1
	nlabel       = 1
	return_label int
	return_reg   int
	break_label  int
)

func add(op, lhs, rhs int) *IR {
	ir := new(IR)
	ir.op = op
	ir.lhs = lhs
	ir.rhs = rhs
	vec_push(code, ir)
	return ir
}

func add_imm(op, lhs, rhs int) *IR {
	ir := add(op, lhs, rhs)
	ir.is_imm = true
	return ir
}

func kill(r int) {
	add(IR_KILL, r, -1)
}

func label(x int) {
	add(IR_LABEL, x, -1)
}

func jmp(x int) {
	add(IR_JMP, x, -1)
}

func load(node *Node, dst, src int) {
	ir := add(IR_LOAD, dst, src)
	ir.size = node.ty.size
}

func store(node *Node, dst, src int) {
	ir := add(IR_STORE, dst, src)
	ir.size = node.ty.size
}

func store_arg(node *Node, bpoff, argreg int) {
	ir := add(IR_STORE_ARG, bpoff, argreg)
	ir.size = node.ty.size
}

// In C, all expressions that can be written on the left-hand side of
// the '=' operator must habe an address in memory. IN other words, if
// you can apply the '&' operator to take an address of some
// expression E, you can assign E to a new value.
//
// Other expressions, such as `1+2`, cannot be written on the lhs of
// '=', since they are just temporary values that don't have an address.
//
// The stuff that can be written on the lhs of '=' os called lvalue.
// Other values are called rvalue. An lvalue is essentially an address.
//
// When lvalues appear on the rvalue context, they are converted to
// rvalues by loading their values from their addresses. You can think
// '&' as an operator that suppresses such auutomatic lvalue-to-rvalue
// conversion.
//
// This function evaluates a given node as an lvalue.
func gen_lval(node *Node) int {
	if node.op == ND_DEREF {
		return gen_expr(node.expr)
	}

	if node.op == ND_DOT {
		r := gen_lval(node.expr)
		add_imm(IR_ADD, r, node.offset)
		return r
	}

	if node.op == ND_LVAR {
		r := nreg
		nreg++
		add(IR_BPREL, r, node.offset)
		return r
	}
	// assert(node.op == ND_GVAR)
	r := nreg
	nreg++
	ir := add(IR_LABEL_ADDR, r, -1)
	ir.name = node.name
	return r
}

func gen_binop(ty int, node *Node) int {
	lhs, rhs := gen_expr(node.lhs), gen_expr(node.rhs)
	add(ty, lhs, rhs)
	kill(rhs)
	return lhs
}

func get_inc_scale(node *Node) int {
	if node.ty.ty == PTR {
		return node.ty.ptr_to.size
	}
	return 1
}

func gen_pre_inc(node *Node, num int) int {
	addr := gen_lval(node.expr)
	val := nreg
	nreg++
	load(node, val, addr)
	add_imm(IR_ADD, val, num*get_inc_scale(node))
	store(node, addr, val)
	kill(addr)
	return val
}

func gen_post_inc(node *Node, num int) int {
	val := gen_pre_inc(node, num)
	add_imm(IR_SUB, val, num*get_inc_scale(node))
	return val
}

func to_assign_op(op int) int {
	switch op {
	case ND_MUL_EQ:
		return IR_MUL
	case ND_DIV_EQ:
		return IR_DIV
	case ND_MOD_EQ:
		return IR_MOD
	case ND_ADD_EQ:
		return IR_ADD
	case ND_SUB_EQ:
		return IR_SUB
	case ND_SHL_EQ:
		return IR_SHL
	case ND_SHR_EQ:
		return IR_SHR
	case ND_BITAND_EQ:
		return IR_AND
	case ND_XOR_EQ:
		return IR_XOR
	default:
		//assert(op == ND_BITOR_EQ)
		return IR_OR
	}
}

func gen_assign_op(node *Node) int {
	src := gen_expr(node.rhs)
	dst := gen_lval(node.lhs)
	val := nreg
	nreg++

	load(node, val, dst)
	add(to_assign_op(node.op), val, src)
	kill(src)
	store(node, dst, val)
	kill(dst)
	return val
}

func gen_expr(node *Node) int {

	switch node.op {
	case ND_NUM:
		{
			r := nreg
			nreg++
			add(IR_IMM, r, node.val)
			return r
		}
	case ND_EQ:
		return gen_binop(IR_EQ, node)
	case ND_NE:
		return gen_binop(IR_NE, node)
	case ND_LOGAND:
		{
			x := nlabel
			nlabel++
			r1 := gen_expr(node.lhs)
			add(IR_UNLESS, r1, x)
			r2 := gen_expr(node.rhs)
			add(IR_MOV, r1, r2)
			kill(r2)
			add(IR_UNLESS, r1, x)
			add(IR_IMM, r1, 1)
			label(x)
			return r1
		}
	case ND_LOGOR:
		{
			x := nlabel
			nlabel++
			y := nlabel
			nlabel++
			r1 := gen_expr(node.lhs)
			add(IR_UNLESS, r1, x)
			add(IR_IMM, r1, 1)
			jmp(y)
			label(x)
			r2 := gen_expr(node.rhs)
			add(IR_MOV, r1, r2)
			kill(r2)
			add(IR_UNLESS, r1, y)
			add(IR_IMM, r1, 1)
			label(y)
			return r1
		}
	case ND_GVAR, ND_LVAR, ND_DOT:
		{
			r := gen_lval(node)
			load(node, r, r)
			return r
		}

	case ND_CALL:
		{
			var args [6]int
			for i := 0; i < node.args.len; i++ {
				args[i] = gen_expr(node.args.data[i].(*Node))
			}
			r := nreg
			nreg++

			ir := add(IR_CALL, r, -1)
			ir.name = node.name
			ir.nargs = node.args.len
			for i := 0; i < 6; i++ {
				ir.args[i] = args[i]
			}
			for i := 0; i < ir.nargs; i++ {
				kill(ir.args[i])
			}
			return r
		}
	case ND_ADDR:
		{
			return gen_lval(node.expr)
		}
	case ND_DEREF:
		{
			r := gen_expr(node.expr)
			load(node, r, r)
			return r
		}
	case ND_STMT_EXPR:
		{
			orig_label := return_label
			orig_reg := return_reg
			return_label = nlabel
			nlabel++
			r := nreg
			nreg++
			return_reg = r

			gen_stmt(node.body)
			label(return_label)

			return_label = orig_label
			return_reg = orig_reg
			return r
		}
	case ND_MUL_EQ, ND_DIV_EQ, ND_MOD_EQ, ND_ADD_EQ, ND_SUB_EQ, ND_SHL_EQ, ND_SHR_EQ, ND_BITAND_EQ, ND_XOR_EQ, ND_BITOR_EQ:
		return gen_assign_op(node)
	case '=':
		{
			rhs, lhs := gen_expr(node.rhs), gen_lval(node.lhs)
			store(node, lhs, rhs)
			kill(lhs)
			return rhs
		}
	case '+':
		return gen_binop(IR_ADD, node)
	case '-':
		return gen_binop(IR_SUB, node)
	case '*':
		return gen_binop(IR_MUL, node)
	case '/':
		return gen_binop(IR_DIV, node)
	case '%':
		return gen_binop(IR_MOD, node)
	case '<':
		return gen_binop(IR_LT, node)
	case ND_LE:
		return gen_binop(IR_LE, node)
	case '&':
		return gen_binop(IR_AND, node)
	case '|':
		return gen_binop(IR_OR, node)
	case '^':
		return gen_binop(IR_XOR, node)
	case ND_SHL:
		return gen_binop(IR_SHL, node)
	case ND_SHR:
		return gen_binop(IR_SHR, node)
	case '~':
		{
			r := gen_expr(node.expr)
			add_imm(IR_XOR, r, -1)
			return r
		}
	case ND_NEG:
		{
			r := gen_expr(node.expr)
			add(IR_NEG, r, -1)
			return r
		}
	case ND_POST_INC:
		return gen_post_inc(node, 1)
	case ND_POST_DEC:
		return gen_post_inc(node, -1)
	case ',':
		kill(gen_expr(node.lhs))
		return gen_expr(node.rhs)
	case '?':
		{
			x := nlabel
			nlabel++
			y := nlabel
			nlabel++
			r := gen_expr(node.cond)

			add(IR_UNLESS, r, x)
			r2 := gen_expr(node.then)
			add(IR_MOV, r, r2)
			kill(r2)
			jmp(y)

			label(x)
			r3 := gen_expr(node.els)
			add(IR_MOV, r, r3)
			kill(r2)
			label(y)
			return r
		}
	case '!':
		{
			lhs := gen_expr(node.expr)
			rhs := nreg
			nreg++
			add(IR_IMM, rhs, 0)
			add(IR_EQ, lhs, rhs)
			kill(rhs)
			return lhs
		}
	default:
		//assert(0 && "unknown AST type")
	}

	return 0
}

func gen_stmt(node *Node) {
	switch node.op {
	case ND_NULL:
		return

	case ND_VARDEF:
		{
			if node.init == nil {
				return
			}
			rhs := gen_expr(node.init)
			lhs := nreg
			nreg++
			add(IR_BPREL, lhs, node.offset)
			store(node, lhs, rhs)
			kill(lhs)
			kill(rhs)
			return
		}
	case ND_IF:
		{
			if node.els != nil {
				x := nlabel
				nlabel++
				y := nlabel
				nlabel++
				r := gen_expr(node.cond)
				add(IR_UNLESS, r, x)
				kill(r)
				gen_stmt(node.then)
				jmp(y)
				label(x)
				gen_stmt(node.els)
				label(y)
				return
			}
			x := nlabel
			nlabel++
			r := gen_expr(node.cond)
			add(IR_UNLESS, r, x)
			kill(r)
			gen_stmt(node.then)
			label(x)
			return
		}
	case ND_FOR:
		{
			x := nlabel
			nlabel++
			y := nlabel
			nlabel++
			orig := break_label
			break_label = nlabel
			nlabel++

			gen_stmt(node.init)
			label(x)
			if node.cond != nil {
				r := gen_expr(node.cond)
				add(IR_UNLESS, r, y)
				kill(r)
			}
			gen_stmt(node.body)
			if node.inc != nil {
				gen_stmt(node.inc)
			}
			jmp(x)
			label(y)
			label(break_label)
			break_label = orig
			return
		}
	case ND_DO_WHILE:
		{
			x := nlabel
			nlabel++
			orig := break_label
			break_label = nlabel
			nlabel++
			label(x)
			gen_stmt(node.body)
			r := gen_expr(node.cond)
			add(IR_IF, r, x)
			kill(r)
			label(break_label)
			break_label = orig
			return
		}
	case ND_BREAK:
		if break_label == 0 {
			error("stray 'break' statement")
		}
		jmp(break_label)
	case ND_RETURN:
		{
			r := gen_expr(node.expr)

			// Statement expression (GNU extension)
			if return_label != 0 {
				add(IR_MOV, return_reg, r)
				kill(r)
				jmp(return_label)
				return
			}

			add(IR_RETURN, r, -1)
			kill(r)
			return
		}
	case ND_EXPR_STMT:
		{
			kill(gen_expr(node.expr))
			return
		}
	case ND_COMP_STMT:
		{
			for i := 0; i < node.stmts.len; i++ {
				gen_stmt((node.stmts.data[i]).(*Node))
			}
			return
		}
	default:
		error("unknown node: %d", node.op)
	}
}

func gen_ir(nodes *Vector) *Vector {
	v := new_vec()
	nlabel = 1

	for i := 0; i < nodes.len; i++ {
		node := nodes.data[i].(*Node)

		if node.op == ND_VARDEF || node.op == ND_DECL {
			continue
		}

		//assert(node.op == ND_FUNC)
		code = new_vec()

		for i := 0; i < node.args.len; i++ {
			arg := node.args.data[i].(*Node)
			store_arg(arg, arg.offset, i)
		}

		gen_stmt(node.body)

		fn := new(Function)
		fn.name = node.name
		fn.stacksize = node.stacksize
		fn.ir = code
		fn.globals = node.globals
		vec_push(v, fn)
	}
	return v
}
