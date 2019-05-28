package main

var (
	pos = 0
)

const (
	ND_NUM = iota + 256 // Number literal
)

type Node struct {
	ty  int   // Node type
	lhs *Node // left-hand side
	rhs *Node // right-hand side
	val int   // Number literal
}

func new_node(op int, lhs, rhs *Node) *Node {
	node := new(Node)
	node.ty = op
	node.lhs = lhs
	node.rhs = rhs
	return node
}

func new_node_num(val int) *Node {
	node := new(Node)
	node.ty = ND_NUM
	node.val = val
	return node
}

func number() *Node {
	t := (tokens.data[pos]).(*Token)
	if t.ty != TK_NUM {
		error("number expected, but got %s", t.input)
		return nil
	}
	pos++
	return new_node_num(t.val)
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

func parse(v *Vector) *Node {
	tokens = v
	pos = 0

	node := expr()
	t := tokens.data[pos].(*Token)
	if t.ty != TK_EOF {
		error("stray token: %s", t.input)
	}
	return node
}
