package main

// Semantics analyzer. This pass plays a few important roles as shown
// below:
//
// - Add types to nodes. For exapmle, a tree that represents "1+2" is
//   typed as INT becaluse the result type of an addition of two
//   integers is integer.
//
// - Resolve variable names based on the C scope rules.
//   Local variables are resolved to offsets from the base pointer.
//   Global variables are resolved to their names.
//
// - Insert nodes to make array-to-pointer conversion explicit.
//   Recall that, in C, "array of T" is automatically converted to
//   "pointer to T" in most contexts.
//
// - Reject bad assignments, such as `1=2+3`.

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
	name      string
	is_extern bool
	data      string
	len       int
}

func new_env(next *Env) *Env {
	env := new(Env)
	env.vars = new_map()
	env.next = next
	return env
}

func new_global(ty *Type, name, data string, len int) *Var {
	v := new(Var)
	v.ty = ty
	v.is_local = false
	v.name = name
	v.data = data
	v.len = len
	return v
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
	node.ty = ptr_to(base.ty.ary_of)
	node.expr = base
	return node
}

func check_lval(node *Node) {
	op := node.op
	if op == ND_LVAR || op == ND_GVAR || op == ND_DEREF || op == ND_DOT {
		return
	}
	error("not an lvalue: %d (%s)", op, node.name)
}

func new_int(val int) *Node {
	node := new(Node)
	node.op = ND_NUM
	node.ty = new(Type)
	node.ty.ty = INT
	node.val = val
	return node
}

func walk(node *Node, env *Env, decay bool) *Node {
	switch node.op {
	case ND_NUM:
		return node
	case ND_STR:
		{
			v := new_global(node.ty, format(".L.str%d", str_label), node.data, node.len)
			str_label++
			vec_push(globals, v)

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

			if v.is_local {
				ret := new(Node)
				ret.op = ND_LVAR
				ret.offset = v.offset
				ret.ty = v.ty
				return maybe_decay(ret, decay)
			}

			ret := new(Node)
			ret.op = ND_GVAR
			ret.ty = v.ty
			ret.name = v.name
			return maybe_decay(ret, decay)
		}
	case ND_VARDEF:
		{
			stacksize = roundup(stacksize, node.ty.align)
			stacksize += node.ty.size
			node.offset = stacksize
			v := new(Var)
			v.ty = node.ty
			v.is_local = true
			v.offset = stacksize
			map_put(env.vars, node.name, v)

			if node.init != nil {
				node.init = walk(node.init, env, true)
			}
			return node
		}
	case ND_IF:
		node.cond = walk(node.cond, env, true)
		node.then = walk(node.then, env, true)
		if node.els != nil {
			node.els = walk(node.els, env, true)
		}
		return node
	case ND_FOR:
		node.init = walk(node.init, env, true)
		node.cond = walk(node.cond, env, true)
		node.inc = walk(node.inc, env, true)
		node.body = walk(node.body, env, true)
		return node
	case ND_DO_WHILE:
		node.cond = walk(node.cond, env, true)
		node.body = walk(node.body, env, true)
		return node
	case '+', '-':
		node.lhs = walk(node.lhs, env, true)
		node.rhs = walk(node.rhs, env, true)

		if node.rhs.ty.ty == PTR {
			swap(&node.lhs, &node.rhs)
		}
		if node.rhs.ty.ty == PTR {
			error("pointer %c pointer' is not defined", node.op)
		}

		node.ty = node.lhs.ty
		return node
	case '=':
		node.lhs = walk(node.lhs, env, false)
		check_lval(node.lhs)
		node.rhs = walk(node.rhs, env, true)
		node.ty = node.lhs.ty
		return node

	case ND_DOT:
		node.expr = walk(node.expr, env, true)
		if node.expr.ty.ty != STRUCT {
			error("struct expected before '.'")
		}

		ty := node.expr.ty
		for i := 0; i < ty.members.len; i++ {
			m := ty.members.data[i].(*Node)
			if m.name != node.name {
				continue
			}
			node.ty = m.ty
			node.offset = m.ty.offset
			return maybe_decay(node, decay)
		}
		error("member missing: %s", node.name)
	case '?':
		node.cond = walk(node.cond, env, true)
		node.then = walk(node.then, env, true)
		node.els = walk(node.els, env, true)
		node.ty = node.then.ty
		return node
	case '*', '/', '<', '|', '^', ND_EQ, ND_NE, ND_LOGAND, ND_LOGOR:
		node.lhs = walk(node.lhs, env, true)
		node.rhs = walk(node.rhs, env, true)
		node.ty = node.lhs.ty
		return node
	case ',':
		node.lhs = walk(node.lhs, env, true)
		node.rhs = walk(node.rhs, env, true)
		node.ty = node.rhs.ty
		return node
	case '!':
		node.expr = walk(node.expr, env, true)
		node.ty = node.expr.ty
		return node
	case ND_ADDR:
		node.expr = walk(node.expr, env, true)
		check_lval(node.expr)
		node.ty = ptr_to(node.expr.ty)
		return node
	case ND_DEREF:
		node.expr = walk(node.expr, env, true)

		if node.expr.ty.ty != PTR {
			error("operand must be a pointer")
		}

		if node.expr.ty.ptr_to.ty == VOID {
			error("cannot dereference void pointer")
		}

		node.ty = node.expr.ty.ptr_to
		return node
	case ND_RETURN:
		node.expr = walk(node.expr, env, true)
		return node
	case ND_SIZEOF:
		{
			expr := walk(node.expr, env, false)
			return new_int(expr.ty.size)
		}
	case ND_ALIGNOF:
		{
			expr := walk(node.expr, env, false)
			return new_int(expr.ty.align)
		}
	case ND_CALL:
		for i := 0; i < node.args.len; i++ {
			node.args.data[i] = walk(node.args.data[i].(*Node), env, true)
		}
		node.ty = &int_ty
		return node
	case ND_FUNC:
		for i := 0; i < node.args.len; i++ {
			node.args.data[i] = walk(node.args.data[i].(*Node), env, true)
		}
		node.body = walk(node.body, env, true)
		return node
	case ND_COMP_STMT:
		{
			newenv := new_env(env)
			for i := 0; i < node.stmts.len; i++ {
				node.stmts.data[i] = walk(node.stmts.data[i].(*Node), newenv, true)
			}
			return node
		}
	case ND_EXPR_STMT:
		node.expr = walk(node.expr, env, true)
		return node
	case ND_STMT_EXPR:
		node.body = walk(node.body, env, true)
		node.ty = &int_ty
		return node
	case ND_NULL:
		return node
	default:
		//assert(0 && "unknouwn node type")
	}
	return nil
}

func sema(nodes *Vector) *Vector {
	globals = new_vec()
	topenv := new_env(nil)

	for i := 0; i < nodes.len; i++ {
		node := nodes.data[i].(*Node)

		if node.op == ND_VARDEF {
			v := new_global(node.ty, node.name, node.data, node.len)
			v.is_extern = node.is_extern
			vec_push(globals, v)
			map_put(topenv.vars, node.name, v)
			continue
		}

		//assert(node.op == ND_FUNC)

		stacksize = 0
		walk(node, topenv, true)
		node.stacksize = stacksize
	}

	return globals
}
