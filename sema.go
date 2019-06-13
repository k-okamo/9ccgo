package main

var (
	vars      *Map
	stacksize int
)

type Var struct {
	ty     *Type
	offset int
}

func walk(node *Node) {
	switch node.op {
	case ND_NUM:
		return
	case ND_IDENT:
		{
			v := map_get(vars, node.name).(*Var)
			if v == nil {
				error("undetined variable: %s", node.name)
			}
			node.ty = v.ty
			node.op = ND_LVAR
			node.offset = v.offset
			return
		}
	case ND_VARDEF:
		{
			stacksize += 8
			node.offset = stacksize
			v := new(Var)
			v.ty = node.ty
			v.offset = stacksize
			map_put(vars, node.name, v)

			if node.init != nil {
				walk(node.init)
			}
			return
		}
	case ND_IF:
		walk(node.cond)
		walk(node.then)
		if node.els != nil {
			walk(node.els)
		}
		return
	case ND_FOR:
		walk(node.init)
		walk(node.cond)
		walk(node.inc)
		walk(node.body)
		return
	case '+':
		walk(node.lhs)
		walk(node.rhs)
		node.ty = node.lhs.ty
		return
	case '-', '*', '/', '=', '<', ND_LOGAND:
		walk(node.lhs)
		walk(node.rhs)
		return
	case ND_LOGOR:
		walk(node.lhs)
		walk(node.rhs)
		node.ty = node.lhs.ty
		return
	case ND_DEREF, ND_RETURN:
		walk(node.expr)
		return
	case ND_CALL:
		for i := 0; i < node.args.len; i++ {
			walk(node.args.data[i].(*Node))
		}
		node.ty = &int_ty
		return
	case ND_FUNC:
		for i := 0; i < node.args.len; i++ {
			walk(node.args.data[i].(*Node))
		}
		walk(node.body)
		return
	case ND_COMP_STMT:
		for i := 0; i < node.stmts.len; i++ {
			walk(node.stmts.data[i].(*Node))
		}
		return
	case ND_EXPR_STMT:
		walk(node.expr)
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
		walk(node)
		node.stacksize = stacksize
	}
}
