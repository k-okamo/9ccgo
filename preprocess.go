package main

// C preprocessor

var (
	defined *Map
)

func append_p(v1, v2 *Vector) {
	for i := 0; i < v2.len; i++ {
		vec_push(v1, v2.data[i])
	}
}

func preprocess(tokens *Vector) *Vector {
	if defined == nil {
		defined = new_map()
	}

	v := new_vec()

	for i := 0; i < tokens.len; {
		t := tokens.data[i].(*Token)
		i++

		if t.ty == TK_IDENT {
			macro := map_get(defined, t.name)
			if macro != nil {
				append_p(v, macro.(*Vector))
			} else {
				vec_push(v, t)
			}
			continue
		}

		if t.ty != '#' {
			vec_push(v, t)
			continue
		}

		t = tokens.data[i].(*Token)
		i++
		if t.ty != TK_IDENT {
			bad_token(t, "identifier expected")
		}

		if strcmp(t.name, "define") == 0 {
			t = tokens.data[i].(*Token)
			i++
			if t.ty != TK_IDENT {
				bad_token(t, "macro name expected")
			}
			name := t.name

			v2 := new_vec()
			for i < tokens.len {
				t = tokens.data[i].(*Token)
				i++
				if t.ty == '\n' {
					break
				}
				vec_push(v2, t)
			}

			map_put(defined, name, v2)
			continue
		}

		if strcmp(t.name, "include") == 0 {
			t = tokens.data[i].(*Token)
			i++
			if t.ty != TK_STR {
				bad_token(t, "string expected")
			}

			path := t.str

			t = tokens.data[i].(*Token)
			i++
			if t.ty != '\n' {
				bad_token(t, "newline expected")
			}
			append_p(v, tokenize(path, false))
			continue
		}
	}
	return v
}
