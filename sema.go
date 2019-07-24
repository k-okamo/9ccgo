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
// - Scales operands for pointer arithmetic. E.g. ptr+1 becomes ptr+4
//   for integer and becomes ptr+8 for pointer.
//
// - Reject bad assignments, such as `1=2+3`.

import (
	"fmt"
	"os"
)

var (
	globals   *Vector
	stacksize int
	str_label int
	env       *Env
)

type Env struct {
	vars *Map
	next *Env
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

func find_var(name string) *Var {
	for e := env; e != nil; e = e.next {
		v := map_get(e.vars, name)
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
	if op != ND_LVAR && op != ND_GVAR && op != ND_DEREF && op != ND_DOT {
		error("not an lvalue: %d (%s)", op, node.name)
	}
}

func new_int(val int) *Node {
	node := new(Node)
	node.op = ND_NUM
	node.ty = new(Type)
	node.ty.ty = INT
	node.val = val
	return node
}

func scale_ptr(node *Node, ty *Type) *Node {
	e := new(Node)
	e.op = '*'
	e.lhs = node
	e.rhs = new_int(ty.ptr_to.size)
	return e
}

func walk(node *Node, decay bool) *Node {
	switch node.op {
	case ND_NUM, ND_NULL, ND_BREAK:
		return node
	case ND_STR:
		{
			// A string literal is converted to a reference to an anonymous
			// global variable of type char array.
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
			v := find_var(node.name)
			if v == nil {
				error("undefined variable: %s", node.name)
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
				node.init = walk(node.init, true)
			}
			return node
		}
	case ND_IF:
		node.cond = walk(node.cond, true)
		node.then = walk(node.then, true)
		if node.els != nil {
			node.els = walk(node.els, true)
		}
		return node
	case ND_FOR:
		env = new_env(env)
		node.init = walk(node.init, true)
		if node.cond != nil {
			node.cond = walk(node.cond, true)
		}
		if node.inc != nil {
			node.inc = walk(node.inc, true)
		}
		node.body = walk(node.body, true)
		env = env.next
		return node
	case ND_DO_WHILE:
		node.cond = walk(node.cond, true)
		node.body = walk(node.body, true)
		return node
	case '+', '-':
		node.lhs = walk(node.lhs, true)
		node.rhs = walk(node.rhs, true)

		if node.rhs.ty.ty == PTR {
			swap(&node.lhs, &node.rhs)
		}
		if node.rhs.ty.ty == PTR {
			error("pointer %c pointer' is not defined", node.op)
		}

		if node.lhs.ty.ty == PTR {
			node.rhs = scale_ptr(node.rhs, node.lhs.ty)
		}

		node.ty = node.lhs.ty
		return node
	case ND_ADD_EQ, ND_SUB_EQ:
		node.lhs = walk(node.lhs, false)
		check_lval(node.lhs)
		node.rhs = walk(node.rhs, true)
		node.ty = node.lhs.ty

		if node.lhs.ty.ty == PTR {
			node.rhs = scale_ptr(node.rhs, node.lhs.ty)
		}
		return node
	case '=', ND_MUL_EQ, ND_DIV_EQ, ND_MOD_EQ, ND_SHL_EQ, ND_SHR_EQ, ND_BITAND_EQ, ND_XOR_EQ, ND_BITOR_EQ:
		node.lhs = walk(node.lhs, false)
		check_lval(node.lhs)
		node.rhs = walk(node.rhs, true)
		node.ty = node.lhs.ty
		return node

	case ND_DOT:
		node.expr = walk(node.expr, true)
		if node.expr.ty.ty != STRUCT {
			error("struct expected before '.'")
		}

		ty := node.expr.ty
		if ty.members == nil {
			error("incomplete type")
		}
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
		node.cond = walk(node.cond, true)
		node.then = walk(node.then, true)
		node.els = walk(node.els, true)
		node.ty = node.then.ty
		return node
	case '*', '/', '%', '<', '|', '^', '&', ND_EQ, ND_NE, ND_LE, ND_SHL, ND_SHR, ND_LOGAND, ND_LOGOR:
		node.lhs = walk(node.lhs, true)
		node.rhs = walk(node.rhs, true)
		node.ty = node.lhs.ty
		return node
	case ',':
		node.lhs = walk(node.lhs, true)
		node.rhs = walk(node.rhs, true)
		node.ty = node.rhs.ty
		return node
	case ND_POST_INC, ND_POST_DEC, ND_NEG, '!', '~':
		node.expr = walk(node.expr, true)
		node.ty = node.expr.ty
		return node
	case ND_ADDR:
		node.expr = walk(node.expr, true)
		check_lval(node.expr)
		node.ty = ptr_to(node.expr.ty)
		return node
	case ND_DEREF:
		node.expr = walk(node.expr, true)

		if node.expr.ty.ty != PTR {
			error("operand must be a pointer")
		}

		if node.expr.ty.ptr_to.ty == VOID {
			error("cannot dereference void pointer")
		}

		node.ty = node.expr.ty.ptr_to
		return maybe_decay(node, decay)
	case ND_RETURN, ND_EXPR_STMT:
		node.expr = walk(node.expr, true)
		return node
	case ND_SIZEOF:
		{
			expr := walk(node.expr, false)
			return new_int(expr.ty.size)
		}
	case ND_ALIGNOF:
		{
			expr := walk(node.expr, false)
			return new_int(expr.ty.align)
		}
	case ND_CALL:
		{
			v := find_var(node.name)
			if v != nil && v.ty.ty == FUNC {
				node.ty = v.ty.returning
			} else {
				fmt.Fprintf(os.Stderr, "bad function: %s\n", node.name)
				node.ty = &int_ty
			}

			for i := 0; i < node.args.len; i++ {
				node.args.data[i] = walk(node.args.data[i].(*Node), true)
			}
			return node
		}
	case ND_COMP_STMT:
		{
			env = new_env(env)
			for i := 0; i < node.stmts.len; i++ {
				node.stmts.data[i] = walk(node.stmts.data[i].(*Node), true)
			}
			env = env.next
			return node
		}
	case ND_STMT_EXPR:
		node.body = walk(node.body, true)
		node.ty = &int_ty
		return node
	default:
		//assert(0 && "unknouwn node type")
	}
	return nil
}

func sema(nodes *Vector) *Vector {
	env = new_env(nil)
	globals = new_vec()

	for i := 0; i < nodes.len; i++ {
		node := nodes.data[i].(*Node)

		if node.op == ND_VARDEF {
			v := new_global(node.ty, node.name, node.data, node.len)
			v.is_extern = node.is_extern
			vec_push(globals, v)
			map_put(env.vars, node.name, v)
			continue
		}

		//assert(node.op == ND_FUNC || node.op == ND_FUNC)

		v := new_global(node.ty, node.name, "", 0)
		map_put(env.vars, node.name, v)

		if node.op == ND_DECL {
			continue
		}

		stacksize = 0

		for i := 0; i < node.args.len; i++ {
			node.args.data[i] = walk(node.args.data[i].(*Node), true)
		}
		node.body = walk(node.body, true)

		node.stacksize = stacksize
	}

	return globals
}
