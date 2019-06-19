package main

var (
	vars      *Map
	stacksize int
)

type Var struct {
	ty     *Type
	offset int
}

func swap(p, q **Node) {
	r := *p
	*p = *q
	*q = r
}

func addr_of(base *Node, ty *Type) *Node {
	node := new(Node)
	node.op = ND_ADDR
	node.ty = ptr_of(ty)

	cp := new(Node)
	// memcpy(cp, base, sizeof(Node))
	copy_node(base, cp)
	node.expr = cp
	return node
}

func walk(node *Node, decay bool) {
	switch node.op {
	case ND_NUM:
		return
	case ND_IDENT:
		{
			v := map_get(vars, node.name).(*Var)
			if v == nil {
				error("undetined variable: %s", node.name)
			}
			node.op = ND_LVAR
			node.offset = v.offset

			if decay && v.ty.ty == ARY {
				*node = *addr_of(node, v.ty.ary_of)
			} else {
				node.ty = v.ty
			}
			return
		}
	case ND_VARDEF:
		{
			stacksize += size_of(node.ty)
			node.offset = stacksize
			v := new(Var)
			v.ty = node.ty
			v.offset = stacksize
			map_put(vars, node.name, v)

			if node.init != nil {
				walk(node.init, true)
			}
			return
		}
	case ND_IF:
		walk(node.cond, true)
		walk(node.then, true)
		if node.els != nil {
			walk(node.els, true)
		}
		return
	case ND_FOR:
		walk(node.init, true)
		walk(node.cond, true)
		walk(node.inc, true)
		walk(node.body, true)
		return
	case '+', '-':
		walk(node.lhs, true)
		walk(node.rhs, true)

		if node.rhs.ty.ty == PTR {
			swap(&node.lhs, &node.rhs)
		}
		if node.rhs.ty.ty == PTR {
			error("pointer %c pointer' is not defined", node.op)
		}

		node.ty = node.lhs.ty
		return
	case '=':
		walk(node.lhs, false)
		walk(node.rhs, true)
		node.ty = node.lhs.ty
		return
	case '*', '/', '<', ND_LOGAND, ND_LOGOR:
		walk(node.lhs, true)
		walk(node.rhs, true)
		node.ty = node.lhs.ty
		return
	case ND_DEREF:
		walk(node.expr, true)
		if node.expr.ty.ty != PTR {
			error("operand must be a pointer")
		}
		node.ty = node.expr.ty.ptr_of
		return
	case ND_RETURN:
		walk(node.expr, true)
		return
	case ND_CALL:
		for i := 0; i < node.args.len; i++ {
			walk(node.args.data[i].(*Node), true)
		}
		node.ty = &int_ty
		return
	case ND_FUNC:
		for i := 0; i < node.args.len; i++ {
			walk(node.args.data[i].(*Node), true)
		}
		walk(node.body, true)
		return
	case ND_COMP_STMT:
		for i := 0; i < node.stmts.len; i++ {
			walk(node.stmts.data[i].(*Node), true)
		}
		return
	case ND_EXPR_STMT:
		walk(node.expr, true)
		return
	default:
		//assert(0 && "unknouwn node type")
	}
}

func sema(nodes *Vector) {
	for i := 0; i < nodes.len; i++ {
		node := nodes.data[i].(*Node)
		//assert(node.op == ND_FUNC)

		vars = new_map()
		stacksize = 0
		walk(node, true)
		node.stacksize = stacksize
	}
}
