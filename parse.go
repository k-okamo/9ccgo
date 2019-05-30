package main

var (
	pos = 0
)

const (
	ND_NUM       = iota + 256 // Number literal
	ND_RETURN                 // Return statement
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
}

func expect(ty int) {
	t := tokens.data[pos].(*Token)
	if t.ty != ty {
		error("%c (%d) expected, but got %c (%d)", ty, ty, t.ty, t.ty)
	}
	pos++
}

func new_node(op int, lhs, rhs *Node) *Node {
	node := new(Node)
	node.ty = op
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func number() *Node {
	t := tokens.data[pos].(*Token)
	if t.ty != TK_NUM {
		error("number expected, but got %s", t.input)
		return nil
	}
	pos++
	node := new(Node)
	node.ty = ND_NUM
	node.val = t.val
	return node
}

func mul() *Node {
	lhs := number()
	for {
		t := tokens.data[pos].(*Token)
		op := t.ty
		if op != '*' && op != '/' {
			return lhs
		}
		pos++
		lhs = new_node(op, lhs, number())
	}
	return lhs
}

func expr() *Node {

	lhs := mul()
	for {
		t := tokens.data[pos].(*Token)
		op := t.ty
		if op != '+' && op != '-' {
			return lhs
		}
		pos++
		lhs = new_node(op, lhs, mul())
	}
	return lhs
}

func stmt() *Node {

	node := new(Node)
	node.ty = ND_COMP_STMT
	node.stmts = new_vec()

	for {
		t := tokens.data[pos].(*Token)
		if t.ty == TK_EOF {
			return node
		}

		e := new(Node)
		if t.ty == TK_RETURN {
			pos++
			e.ty = ND_RETURN
			e.expr = expr()
		} else {
			e.ty = ND_EXPR_STMT
			e.expr = expr()
		}

		vec_push(node.stmts, e)
		expect(';')
	}
}

func parse(v *Vector) *Node {
	tokens = v
	pos = 0
	return stmt()
}
