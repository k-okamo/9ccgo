package main

var (
	pos    = 0
	int_ty = Type{ty: INT, ptr_of: nil}
)

const (
	ND_NUM       = iota + 256 // Number literal
	ND_IDENT                  // Identigier
	ND_VARDEF                 // Variable definition
	ND_LVAR                   // Variable reference
	ND_IF                     // "if"
	ND_FOR                    // "for"
	ND_ADDR                   // address-of operator ("&")
	ND_DEREF                  // pointer dereference ("*")
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

func is_typename() bool {
	t := tokens.data[pos].(*Token)
	return t.ty == TK_INT
}

func new_node(op int, lhs, rhs *Node) *Node {
	node := new(Node)
	node.op = op
	node.lhs = lhs
	node.rhs = rhs
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
		node := new(Node)
		node.op = ND_DEREF
		node.expr = new_node('+', lhs, primary())
		lhs = node
		expect(']')
	}
	return lhs
}

func unary() *Node {
	if consume('*') {
		node := new(Node)
		node.op = ND_DEREF
		node.expr = mul()
		return node
	}
	if consume('&') {
		node := new(Node)
		node.op = ND_ADDR
		node.expr = mul()
		return node
	}
	if consume(TK_SIZEOF) {
		node := new(Node)
		node.op = ND_SIZEOF
		node.expr = unary()
		return node
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
		lhs = new_node(t.ty, lhs, unary())
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
		lhs = new_node(t.ty, lhs, mul())
	}
	return lhs
}

func rel() *Node {
	lhs := parse_add()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty == '<' {
			pos++
			lhs = new_node('<', lhs, parse_add())
			continue
		}
		if t.ty == '>' {
			pos++
			lhs = new_node('<', parse_add(), lhs)
			continue
		}
		return lhs
	}
}

func logand() *Node {
	lhs := rel()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty != TK_LOGAND {
			return lhs
		}
		pos++
		lhs = new_node(ND_LOGAND, lhs, rel())
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
		lhs = new_node(ND_LOGOR, lhs, logand())
	}
	return lhs
}

func assign() *Node {
	lhs := logor()
	if consume('=') {
		return new_node('=', lhs, logor())
	}
	return lhs
}

func ttype() *Type {
	t := tokens.data[pos].(*Token)
	if t.ty != TK_INT {
		error("typename expected, but got %s", t.input)
	}
	pos++
	ty := &int_ty
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
	node := new(Node)
	node.op = ND_EXPR_STMT
	node.expr = assign()
	expect(';')
	return node
}

func stmt() *Node {
	node := new(Node)
	t := tokens.data[pos].(*Token)

	switch t.ty {
	case TK_INT:
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
		if is_typename() {
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

func function() *Node {
	node := new(Node)
	node.op = ND_FUNC
	node.args = new_vec()

	t := tokens.data[pos].(*Token)
	if t.ty != TK_INT {
		error("function return type expected, but got %s", t.input)
	}
	pos++
	t = tokens.data[pos].(*Token)
	if t.ty != TK_IDENT {
		error("function name expected, but got %s", t.input)
	}
	node.name = t.name
	pos++

	expect('(')
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

func parse(tokens_ *Vector) *Vector {
	tokens = tokens_
	pos = 0
	v := new_vec()
	for (tokens.data[pos].(*Token)).ty != TK_EOF {
		vec_push(v, function())
	}
	return v
}
