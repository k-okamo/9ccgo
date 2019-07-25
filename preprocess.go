package main

// C preprocessor

func preprocess(tokens *Vector) *Vector {
	v := new_vec()

	for i := 0; i < tokens.len; {
		t := tokens.data[i].(*Token)
		if t.ty != '#' {
			i++
			vec_push(v, t)
			continue
		}

		i++
		t = tokens.data[i].(*Token)
		if t.ty != TK_IDENT || strcmp(t.name, "include") != 0 {
			bad_token(t, "'include' expected")
		}

		i++
		t = tokens.data[i].(*Token)
		if t.ty != TK_STR {
			bad_token(t, "string expected")
		}

		path := t.str

		i++
		t = tokens.data[i].(*Token)
		if t.ty != '\n' {
			bad_token(t, "newline expected")
		}

		nv := tokenize(path, false)
		for i := 0; i < nv.len; i++ {
			vec_push(v, nv.data[i].(*Token))
		}
	}
	return v
}
