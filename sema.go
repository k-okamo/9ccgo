package main

var (
	globals   *Vector
	stacksize int
	str_label int
)

type Env struct {
	vars *Map
	next *Env
}

type Var struct {
	ty       *Type
	is_local bool

	// local
	offset int

	// global
	name string
	data string
	len  int
}

func new_env(next *Env) *Env {
	env := new(Env)
	env.vars = new_map()
	env.next = next
	return env
}

func find(env *Env, name string) *Var {
	for ; env != nil; env = env.next {
		v := map_get(env.vars, name)
		if v != nil {
			return v.(*Var)
		}
	}
	return (*Var)(nil)
}

func swap(p, q **Node) {
	r := *p
	*p = *q
	*q = r
}

func maybe_decay(base *Node, decay bool) *Node {
	if !decay || base.ty.ty != ARY {
		return base
	}

	node := new(Node)
	node.op = ND_ADDR
	node.ty = ptr_of(base.ty.ary_of)
	node.expr = base
	return node
}

func walk(env *Env, node *Node, decay bool) *Node {
	switch node.op {
	case ND_NUM:
		return node
	case ND_STR:
		{
			v := new(Var)
			vec_push(globals, v)
			v.ty = node.ty
			v.is_local = false
			v.name = format(".L.str%d", str_label)
			str_label++
			v.data = node.str
			v.len = len(node.str) + 1

			ret := new(Node)
			ret.op = ND_GVAR
			ret.ty = node.ty
			ret.name = v.name
			return maybe_decay(ret, decay)
		}
	case ND_IDENT:
		{
			v := find(env, node.name)
			if v == nil {
				error("undetined variable: %s", node.name)
			}

			ret := new(Node)
			ret.op = ND_LVAR
			ret.offset = v.offset
			ret.ty = v.ty
			return maybe_decay(ret, decay)
		}
	case ND_VARDEF:
		{
			stacksize += size_of(node.ty)
			node.offset = stacksize
			v := new(Var)
			v.ty = node.ty
			v.is_local = true
			v.offset = stacksize
			map_put(env.vars, node.name, v)

			if node.init != nil {
				node.init = walk(env, node.init, true)
			}
			return node
		}
	case ND_IF:
		node.cond = walk(env, node.cond, true)
		node.then = walk(env, node.then, true)
		if node.els != nil {
			node.els = walk(env, node.els, true)
		}
		return node
	case ND_FOR:
		node.init = walk(env, node.init, true)
		node.cond = walk(env, node.cond, true)
		node.inc = walk(env, node.inc, true)
		node.body = walk(env, node.body, true)
		return node
	case '+', '-':
		node.lhs = walk(env, node.lhs, true)
		node.rhs = walk(env, node.rhs, true)

		if node.rhs.ty.ty == PTR {
			swap(&node.lhs, &node.rhs)
		}
		if node.rhs.ty.ty == PTR {
			error("pointer %c pointer' is not defined", node.op)
		}

		node.ty = node.lhs.ty
		return node
	case '=':
		node.lhs = walk(env, node.lhs, false)
		if node.lhs.op != ND_LVAR && node.lhs.op != ND_DEREF {
			error("not an lvalue: %d (%s)", node.op, node.name)
		}
		node.rhs = walk(env, node.rhs, true)
		node.ty = node.lhs.ty
		return node
	case '*', '/', '<', ND_LOGAND, ND_LOGOR:
		node.lhs = walk(env, node.lhs, true)
		node.rhs = walk(env, node.rhs, true)
		node.ty = node.lhs.ty
		return node
	case ND_ADDR:
		node.expr = walk(env, node.expr, true)
		node.ty = ptr_of(node.expr.ty)
		return node
	case ND_DEREF:
		node.expr = walk(env, node.expr, true)
		if node.expr.ty.ty != PTR {
			error("operand must be a pointer")
		}
		node.ty = node.expr.ty.ptr_of
		return node
	case ND_RETURN:
		node.expr = walk(env, node.expr, true)
		return node
	case ND_SIZEOF:
		{
			expr := walk(env, node.expr, false)

			ret := new(Node)
			ret.op = ND_NUM
			ret.ty = new(Type)
			ret.ty.ty = INT
			ret.val = size_of(expr.ty)
			return ret
		}
	case ND_CALL:
		for i := 0; i < node.args.len; i++ {
			node.args.data[i] = walk(env, node.args.data[i].(*Node), true)
		}
		node.ty = &int_ty
		return node
	case ND_FUNC:
		for i := 0; i < node.args.len; i++ {
			node.args.data[i] = walk(env, node.args.data[i].(*Node), true)
		}
		node.body = walk(env, node.body, true)
		return node
	case ND_COMP_STMT:
		{
			newenv := new_env(env)
			for i := 0; i < node.stmts.len; i++ {
				node.stmts.data[i] = walk(newenv, node.stmts.data[i].(*Node), true)
			}
			return node
		}
	case ND_EXPR_STMT:
		node.expr = walk(env, node.expr, true)
		return node
	default:
		//assert(0 && "unknouwn node type")
	}
	return nil
}

func sema(nodes *Vector) {
	for i := 0; i < nodes.len; i++ {
		node := nodes.data[i].(*Node)
		//assert(node.op == ND_FUNC)

		globals = new_vec()
		stacksize = 0
		walk(new_env(nil), node, true)
		node.stacksize = stacksize
		node.globals = globals
	}
}
