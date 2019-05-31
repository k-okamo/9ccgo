package main

import (
	"fmt"
	"unicode"
)

var (
	tokens   *Vector
	keywords *Map
)

const (
	TK_NUM    = iota + 256 // Number literal
	TK_IDENT               // Identifier
	TK_RETURN              // "return"
	TK_EOF                 // End marker
)

// Token type
type Token struct {
	ty    int    // Token type
	val   int    // Number literal
	name  string // Identifier
	input string // Token string (for error reporting)
}

func add_token(v *Vector, ty int, input string) *Token {
	t := new(Token)
	t.ty = ty
	t.input = input
	vec_push(v, t)
	return t
}

func scan(s string) *Vector {

	v := new_vec()
	i := 0
	for len(s) != 0 {
		// Skip whitespace
		c := []rune(s)[0]
		if unicode.IsSpace(c) {
			s = s[1:]
			continue
		}

		// Single-letter token
		if strchr("+-*/;=()", c) != "" {
			add_token(v, int(c), s)
			i++
			s = s[1:]
			continue
		}

		// Identifier
		if IsAlpha(c) || c == '_' {
			length := 1
		LABEL:
			for {
				if len(s[length:]) == 0 {
					break LABEL
				}
				c2 := []rune(s)[length]
				if !IsAlpha(c2) && !unicode.IsDigit(c2) && c2 != '_' {
					break LABEL
				}
				length++
			}
			name := strndup(s, length)
			ty := map_get(keywords, name).(int)
			if ty == 0 {
				ty = TK_IDENT
			}

			t := add_token(v, ty, s)
			t.name = name
			i++
			s = s[length:]
			continue
		}

		// Number
		if unicode.IsDigit(c) {
			t := add_token(v, TK_NUM, s)
			val := 0
			val, s = strtol(s, 10)
			t.val = val
			i++
			continue
		}

		error("cannot tokenize: %s\n", string(c))
	}

	add_token(v, TK_EOF, s)
	return v
}

func tokenize(s string) *Vector {
	keywords = new_map()
	map_put(keywords, "return", TK_RETURN)

	return scan(s)
}

// [Debug] tokens print
func print_tokens(tokens *Vector) {
	if !debug {
		return
	}
	fmt.Println("-- tokens info --")
	for i := 0; i < tokens.len; i++ {
		t := tokens.data[i].(*Token)
		ty := ""
		switch t.ty {
		case TK_NUM:
			ty = "TK_NUM   "
		case TK_IDENT:
			ty = "TK_IDENT "
		case TK_RETURN:
			ty = "TK_RETURN"
		case TK_EOF:
			ty = "TK_EOF   "
		case ';':
			ty = ";        "
		default:
			ty = "         "
		}
		fmt.Printf("[%02d] ty: %s, val: %d, input: %s\n", i, ty, t.val, t.input)
	}
	fmt.Println("")
}
