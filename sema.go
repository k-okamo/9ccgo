package main

var (
	vars      *Map
	stacksize int
)

func walk(node *Node) {
	switch node.ty {
	case ND_NUM:
		return
	case ND_IDENT:
		if !map_exists(vars, node.name) {
			error("undetined variable: %s", node.name)
		}
		node.ty = ND_LVAR
		node.offset = map_get(vars, node.name).(int)
		return
	case ND_VARDEF:
		stacksize += 8
		map_put(vars, node.name, stacksize)
		node.offset = stacksize
		if node.init != nil {
			walk(node.init)
		}
		return
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
	case '+', '-', '*', '/', '=', '<', ND_LOGAND, ND_LOGOR:
		walk(node.lhs)
		walk(node.rhs)
		return
	case ND_RETURN:
		walk(node.expr)
		return
	case ND_CALL:
		for i := 0; i < node.args.len; i++ {
			walk(node.args.data[i].(*Node))
		}
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
		//assert(node.ty == ND_FUNC)

		vars = new_map()
		stacksize = 0
		walk(node)
		node.stacksize = stacksize
	}
}
