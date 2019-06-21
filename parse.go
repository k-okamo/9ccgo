package main

var (
	pos     = 0
	int_ty  = Type{ty: INT, ptr_of: nil}
	char_ty = Type{ty: CHAR, ptr_of: nil}
)

const (
	ND_NUM       = iota + 256 // Number literal
	ND_STR                    // String literal
	ND_IDENT                  // Identigier
	ND_VARDEF                 // Variable definition
	ND_LVAR                   // Local variable reference
	ND_GVAR                   // Global variable reference
	ND_IF                     // "if"
	ND_FOR                    // "for"
	ND_ADDR                   // address-of operator ("&")
	ND_DEREF                  // pointer dereference ("*")
	ND_EQ                     // ==
	ND_NE                     // !=
	ND_LOGOR                  // ||
	ND_LOGAND                 // &&
	ND_RETURN                 // "return"
	ND_SIZEOF                 // "sizeof"
	ND_CALL                   // Function call
	ND_FUNC                   // Function definition
	ND_COMP_STMT              // Compound statement
	ND_EXPR_STMT              // Expressions statement
)

const (
	INT = iota
	CHAR
	PTR
	ARY
)

type Node struct {
	op    int     // Node type
	ty    *Type   // C type
	lhs   *Node   // left-hand side
	rhs   *Node   // right-hand side
	val   int     // Number literal
	expr  *Node   // "return" or expression stmt
	stmts *Vector // Compound statement

	name string // Identifier

	// Global variable
	data string
	len  int

	// "if" ( cond ) then "else" els
	// "for" ( init; cond; inc ) body
	cond *Node
	then *Node
	els  *Node
	init *Node
	body *Node
	inc  *Node

	// Function definition
	stacksize int
	globals   *Vector

	// Local variable
	offset int

	// Function call
	args *Vector
}

type Type struct {
	ty int

	// Pointer
	ptr_of *Type

	//Array
	ary_of *Type
	len    int
}

func expect(ty int) {
	t := tokens.data[pos].(*Token)
	if t.ty != ty {
		error("%c (%d) expected, but got %c (%d)", ty, ty, t.ty, t.ty)
	}
	pos++
}

func consume(ty int) bool {
	t := tokens.data[pos].(*Token)
	if t.ty != ty {
		return false
	}
	pos++
	return true
}

func get_type() *Type {
	t := tokens.data[pos].(*Token)
	if t.ty == TK_INT {
		return &int_ty
	}
	if t.ty == TK_CHAR {
		return &char_ty
	}
	return nil
}

func new_binop(op int, lhs, rhs *Node) *Node {
	node := new(Node)
	node.op = op
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func new_expr(op int, expr *Node) *Node {
	node := new(Node)
	node.op = op
	node.expr = expr
	return node
}

func primary() *Node {
	t := tokens.data[pos].(*Token)
	pos++

	if t.ty == '(' {
		node := assign()
		expect(')')
		return node
	}

	node := new(Node)
	if t.ty == TK_NUM {
		node.ty = &int_ty
		node.op = ND_NUM
		node.val = t.val
		return node
	}

	if t.ty == TK_STR {
		node.ty = ary_of(&char_ty, len(t.str))
		node.op = ND_STR
		node.data = t.str
		node.len = t.len
		return node
	}

	if t.ty == TK_IDENT {
		node.name = t.name

		if !consume('(') {
			node.op = ND_IDENT
			return node
		}

		node.op = ND_CALL
		node.args = new_vec()
		if consume(')') {
			return node
		}

		vec_push(node.args, assign())
		for consume(',') {
			vec_push(node.args, assign())
		}
		expect(')')
		return node
	}

	error("number expected, but got %s", t.input)
	return nil
}

func postfix() *Node {
	lhs := primary()
	for consume('[') {
		lhs = new_expr(ND_DEREF, new_binop('+', lhs, assign()))
		expect(']')
	}
	return lhs
}

func unary() *Node {
	if consume('*') {
		return new_expr(ND_DEREF, mul())
	}
	if consume('&') {
		return new_expr(ND_ADDR, mul())
	}
	if consume(TK_SIZEOF) {
		return new_expr(ND_SIZEOF, unary())
	}
	return postfix()
}

func mul() *Node {
	lhs := unary()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty != '*' && t.ty != '/' {
			return lhs
		}
		pos++
		lhs = new_binop(t.ty, lhs, unary())
	}
	return lhs
}

func read_array(ty *Type) *Type {
	v := new_vec()
	for consume('[') {
		l := primary()
		if l.op != ND_NUM {
			error("number expected")
		}
		vec_push(v, l)
		expect(']')
	}
	for i := v.len - 1; i >= 0; i-- {
		l := v.data[i].(*Node)
		ty = ary_of(ty, l.val)
	}
	return ty
}

func parse_add() *Node {

	lhs := mul()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty != '+' && t.ty != '-' {
			return lhs
		}
		pos++
		lhs = new_binop(t.ty, lhs, mul())
	}
	return lhs
}

func rel() *Node {
	lhs := parse_add()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty == '<' {
			pos++
			lhs = new_binop('<', lhs, parse_add())
			continue
		}
		if t.ty == '>' {
			pos++
			lhs = new_binop('<', parse_add(), lhs)
			continue
		}
		return lhs
	}
}

func equality() *Node {
	lhs := rel()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty == TK_EQ {
			pos++
			lhs = new_binop(ND_EQ, lhs, rel())
			continue
		}
		if t.ty == TK_NE {
			pos++
			lhs = new_binop(ND_NE, lhs, rel())
			continue
		}
		return lhs
	}
}

func logand() *Node {
	lhs := equality()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty != TK_LOGAND {
			return lhs
		}
		pos++
		lhs = new_binop(ND_LOGAND, lhs, equality())
	}
	return lhs
}

func logor() *Node {
	lhs := logand()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty != TK_LOGOR {
			return lhs
		}
		pos++
		lhs = new_binop(ND_LOGOR, lhs, logand())
	}
	return lhs
}

func assign() *Node {
	lhs := logor()
	if consume('=') {
		return new_binop('=', lhs, logor())
	}
	return lhs
}

func ttype() *Type {
	t := tokens.data[pos].(*Token)
	ty := get_type()
	if ty == nil {
		error("typename expected, but got %s", t.input)
	}
	pos++
	for consume('*') {
		ty = ptr_of(ty)
	}
	return ty
}

func decl() *Node {
	node := new(Node)
	node.op = ND_VARDEF

	// Read the first half of type name (e.g. `int *`).
	node.ty = ttype()

	// Read an identifier.
	t := tokens.data[pos].(*Token)
	if t.ty != TK_IDENT {
		error("variable name expected, but got %s", t.input)
	}
	node.name = t.name
	pos++

	// Read the second half of type name (e.g. `[3][5]`).
	node.ty = read_array(node.ty)

	// Read an initializer.
	if consume('=') {
		node.init = assign()
	}
	expect(';')
	return node
}

func param() *Node {
	node := new(Node)
	node.op = ND_VARDEF
	node.ty = ttype()

	t := tokens.data[pos].(*Token)
	if t.ty != TK_IDENT {
		error("parameter name expected, but got %s", t.input)
	}
	node.name = t.name
	pos++
	return node
}

func expr_stmt() *Node {
	node := new_expr(ND_EXPR_STMT, assign())
	expect(';')
	return node
}

func stmt() *Node {
	node := new(Node)
	t := tokens.data[pos].(*Token)

	switch t.ty {
	case TK_INT, TK_CHAR:
		return decl()
	case TK_IF:
		pos++
		node.op = ND_IF
		expect('(')
		node.cond = assign()
		expect(')')

		node.then = stmt()

		if consume(TK_ELSE) {
			node.els = stmt()
		}
		return node
	case TK_FOR:
		pos++
		node.op = ND_FOR
		expect('(')
		if get_type() != nil {
			node.init = decl()
		} else {
			node.init = expr_stmt()
		}
		node.cond = assign()
		expect(';')
		node.inc = assign()
		expect(')')
		node.body = stmt()
		return node
	case TK_RETURN:
		pos++
		node.op = ND_RETURN
		node.expr = assign()
		expect(';')
		return node
	case '{':
		pos++
		node.op = ND_COMP_STMT
		node.stmts = new_vec()
		for !consume('}') {
			vec_push(node.stmts, stmt())
		}
		return node
	default:
		return expr_stmt()
	}
	return nil
}

func compound_stmt() *Node {

	node := new(Node)
	node.op = ND_COMP_STMT
	node.stmts = new_vec()

	for !consume('}') {
		vec_push(node.stmts, stmt())
	}
	return node
}

func toplevel() *Node {

	ty := ttype()
	if ty == nil {
		t := tokens.data[pos].(*Token)
		error("typename expected, but got %s", t.input)
	}

	t := tokens.data[pos].(*Token)
	if t.ty != TK_IDENT {
		error("function or variable name expected, but got %s", t.input)
	}

	name := t.name
	pos++

	// Function
	if consume('(') {
		node := new(Node)
		node.op = ND_FUNC
		node.ty = ty
		node.name = name
		node.args = new_vec()

		if !consume(')') {
			vec_push(node.args, param())
			for consume(',') {
				vec_push(node.args, param())
			}
			expect(')')
		}

		expect('{')
		node.body = compound_stmt()
		return node
	}

	// Global variable
	node := new(Node)
	node.op = ND_VARDEF
	node.ty = read_array(ty)
	node.name = name
	node.data = ""
	node.len = size_of(node.ty)
	expect(';')
	return node
}

func parse(tokens_ *Vector) *Vector {
	tokens = tokens_
	pos = 0
	v := new_vec()
	for (tokens.data[pos].(*Token)).ty != TK_EOF {
		vec_push(v, toplevel())
	}
	return v
}
