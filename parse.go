package main

var (
	pos = 0
)

const (
	ND_NUM       = iota + 256 // Number literal
	ND_IDENT                  // Identigier
	ND_VARDEF                 // Variable definition
	ND_IF                     // "if"
	ND_FOR                    // "for"
	ND_LOGOR                  // ||
	ND_LOGAND                 // &&
	ND_RETURN                 // "return"
	ND_CALL                   // Function call
	ND_FUNC                   // Function definition
	ND_COMP_STMT              // Compound statement
	ND_EXPR_STMT              // Expressions statement
)

type Node struct {
	ty    int     // Node type
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

	// Function call
	args *Vector
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

func new_node(op int, lhs, rhs *Node) *Node {
	node := new(Node)
	node.ty = op
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func term() *Node {
	t := tokens.data[pos].(*Token)
	pos++

	if t.ty == '(' {
		node := assign()
		expect(')')
		return node
	}

	node := new(Node)
	if t.ty == TK_NUM {
		node.ty = ND_NUM
		node.val = t.val
		return node
	}
	if t.ty == TK_IDENT {
		node.name = t.name

		if !consume('(') {
			node.ty = ND_IDENT
			return node
		}

		node.ty = ND_CALL
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

func mul() *Node {
	lhs := term()
	for {
		t := tokens.data[pos].(*Token)
		if t.ty != '*' && t.ty != '/' {
			return lhs
		}
		pos++
		lhs = new_node(t.ty, lhs, term())
	}
	return lhs
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

func stmt() *Node {
	node := new(Node)
	t := tokens.data[pos].(*Token)

	switch t.ty {
	case TK_INT:
		pos++
		node.ty = ND_VARDEF
		t = tokens.data[pos].(*Token)
		if t.ty != TK_IDENT {
			error("variable name expected, but got %s", t.input)
		}
		node.name = t.name
		pos++

		if consume('=') {
			node.init = assign()
		}
		expect(';')
		return node
	case TK_IF:
		pos++
		node.ty = ND_IF
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
		node.ty = ND_FOR
		expect('(')
		node.init = assign()
		expect(';')
		node.cond = assign()
		expect(';')
		node.inc = assign()
		expect(')')
		node.body = stmt()
		return node
	case TK_RETURN:
		pos++
		node.ty = ND_RETURN
		node.expr = assign()
		expect(';')
		return node
	case '{':
		pos++
		node.ty = ND_COMP_STMT
		node.stmts = new_vec()
		for !consume('}') {
			vec_push(node.stmts, stmt())
		}
		return node
	default:
		node.ty = ND_EXPR_STMT
		node.expr = assign()
		expect(';')
		return node
	}
	return nil
}

func compound_stmt() *Node {

	node := new(Node)
	node.ty = ND_COMP_STMT
	node.stmts = new_vec()

	for !consume('}') {
		vec_push(node.stmts, stmt())
	}
	return node
}

func function() *Node {
	node := new(Node)
	node.ty = ND_FUNC
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
		vec_push(node.args, term())
		for consume(',') {
			vec_push(node.args, term())
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
