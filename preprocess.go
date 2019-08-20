package main

// C preprocessor

var (
	macros *Map
	ctx_p  *Context_p
)

const (
	OBJLIKE = iota
	FUNCLIKE
)

type Context_p struct {
	input  *Vector
	output *Vector
	pos    int
	next   *Context_p
}

type Macro struct {
	ty     int
	tokens *Vector
	params *Vector
}

func new_ctx_p(next *Context_p, input *Vector) *Context_p {
	c := new(Context_p)
	c.input = input
	c.output = new_vec()
	c.next = next
	return c
}

func new_macro(ty int, name string) *Macro {
	m := new(Macro)
	m.ty = ty
	m.tokens = new_vec()
	m.params = new_vec()
	map_put(macros, name, m)
	return m
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

func ident_p(msg string) string {
	t := get(TK_IDENT, "parameter name expected")
	return t.name
}

func peek() *Token { return ctx_p.input.data[ctx_p.pos].(*Token) }

func consume_p(ty int) bool {
	if peek().ty != ty {
		return false
	}
	ctx_p.pos++
	return true
}

func read_until_eol() *Vector {
	v := new_vec()
	for !eof() {
		t := next()
		if t.ty == '\n' {
			break
		}
		vec_push(v, t)
	}
	return v
}

func new_int_p(val int) *Token {
	t := new(Token)
	t.ty = TK_NUM
	t.val = val
	return t
}

func new_param(val int) *Token {
	t := new(Token)
	t.ty = TK_PARAM
	t.val = val
	return t
}

func is_ident(t *Token, s string) bool {
	return t.ty == TK_IDENT && strcmp(t.name, s) == 0
}

func replace_params(m *Macro) {
	params := m.params
	tokens := m.tokens

	// Replaces macro parameter tokens with TK_PARAM tokens
	mm := new_map()
	for i := 0; i < params.len; i++ {
		name := params.data[i].(string)
		map_puti(mm, name, i)
	}

	for i := 0; i < tokens.len; i++ {
		t := tokens.data[i].(*Token)
		if t.ty != TK_IDENT {
			continue
		}
		n := map_geti(mm, t.name, -1)
		if n == -1 {
			continue
		}
		tokens.data[i] = new_param(n)
	}

	// Process '#' followed by a macro parameter.
	v := new_vec()
	for i := 0; i < tokens.len; i++ {
		t1 := tokens.data[i].(*Token)
		t2 := tokens.data[i+1]

		if i != tokens.len-1 && t1.ty == '#' && t2.(*Token).ty == TK_PARAM {
			t2.(*Token).stringize = true
			vec_push(v, t2)
			i++
		} else {
			vec_push(v, t1)
		}
	}
	m.tokens = v
}

func read_one_arg() *Vector {
	v := new_vec()
	start := peek()
	level := 0

	for !eof() {
		t := peek()
		if level == 0 {
			if t.ty == ')' || t.ty == ',' {
				return v
			}
		}

		next()
		if t.ty == '(' {
			level++
		} else if t.ty == ')' {
			level--
		}
		vec_push(v, t)
	}
	bad_token(start, "unclosed macro argument")
	return nil
}

func read_args() *Vector {
	v := new_vec()
	if consume_p(')') {
		return v
	}
	vec_push(v, read_one_arg())
	for !consume_p(')') {
		get(',', "comma expected")
		vec_push(v, read_one_arg())
	}
	return v
}

func stringize(tokens *Vector) *Token {
	sb := new_sb()

	for i := 0; i < tokens.len; i++ {
		t := tokens.data[i].(*Token)
		if i != 0 {
			sb_add(sb, " ")
		}
		sb_append(sb, tokstr(t))
	}

	t := new(Token)
	t.ty = TK_STR
	t.str = sb_get(sb)
	t.len = sb.len
	return t
}

func apply(m *Macro, start *Token) {
	if m.ty == OBJLIKE {
		append_p(m.tokens)
		return
	}

	// Function-like macro
	get('(', "comma expected")
	args := read_args()
	if m.params.len != args.len {
		bad_token(start, "number of parameter does not match")
	}

	for i := 0; i < m.tokens.len; i++ {
		t := m.tokens.data[i].(*Token)

		if is_ident(t, "__LINE__") {
			add_p(new_int_p(line(t)))
			continue
		}

		if t.ty == TK_PARAM {
			if t.stringize {
				add_p(stringize(args.data[t.val].(*Vector)))
			} else {
				append_p(args.data[t.val].(*Vector))
			}
			continue
		}
		add_p(t)
	}
}

func funclike_macro(name string) {
	m := new_macro(FUNCLIKE, name)
	vec_push(m.params, ident_p("parameter name expected"))
	for !consume_p(')') {
		get(',', "comma expected")
		vec_push(m.params, ident_p("parameter name expected"))
	}
	m.tokens = read_until_eol()
	replace_params(m)
}

func objlike_macro(name string) {
	m := new_macro(OBJLIKE, name)
	m.tokens = read_until_eol()
}

func define() {
	name := ident_p("macro name expected")
	if consume_p('(') {
		funclike_macro(name)
		return
	}
	objlike_macro(name)
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
			m := map_get(macros, t.name)
			if m != nil {
				apply(m.(*Macro), t)
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
