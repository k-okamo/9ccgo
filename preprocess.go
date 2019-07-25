package main

// C preprocessor

var (
	macros *Map
	ctx_p  *Context_p
)

type Context_p struct {
	input  *Vector
	output *Vector
	pos    int
	next   *Context_p
}

func new_ctx_p(next *Context_p, input *Vector) *Context_p {
	c := new(Context_p)
	c.input = input
	c.output = new_vec()
	c.next = next
	return c
}

func append_p(v *Vector) {
	for i := 0; i < v.len; i++ {
		vec_push(ctx_p.output, v.data[i])
	}
}

func add_p(t *Token) { vec_push(ctx_p.output, t) }

func next() *Token {
	// assert(ctx_p,pos < ctx_p.input.len)
	t := ctx_p.input.data[ctx_p.pos].(*Token)
	ctx_p.pos++
	return t
}

func eof() bool { return ctx_p.pos == ctx_p.input.len }

func get(ty int, msg string) *Token {
	t := next()
	if t.ty != ty {
		bad_token(t, msg)
	}
	return t
}

func define() {
	t := get(TK_IDENT, "macro name expected")
	name := t.name

	v := new_vec()
	for !eof() {
		t = next()
		if t.ty == '\n' {
			break
		}
		vec_push(v, t)
	}
	map_put(macros, name, v)
}

func include() {
	t := get(TK_STR, "string expected")
	path := t.str
	get('\n', "newline expected")
	append_p(tokenize(path, false))
}

func preprocess(tokens *Vector) *Vector {
	if macros == nil {
		macros = new_map()
	}
	ctx_p = new_ctx_p(ctx_p, tokens)

	for !eof() {
		t := next()

		if t.ty == TK_IDENT {
			macro := map_get(macros, t.name)
			if macro != nil {
				append_p(macro.(*Vector))
			} else {
				add_p(t)
			}
			continue
		}

		if t.ty != '#' {
			add_p(t)
			continue
		}

		t = get(TK_IDENT, "identifier expected")

		if strcmp(t.name, "define") == 0 {
			define()
		} else if strcmp(t.name, "include") == 0 {
			include()
		} else {
			bad_token(t, "unknown directive")
		}
	}

	v := ctx_p.output
	ctx_p = ctx_p.next
	return v
}
