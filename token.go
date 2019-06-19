package main

import (
	"fmt"
	"unicode"
)

var (
	tokens   *Vector
	keywords *Map
	symbols  = []Keyword{{name: "&&", ty: TK_LOGAND},
		{name: "||", ty: TK_LOGOR},
		{name: "else", ty: TK_ELSE},
		{name: "for", ty: TK_FOR},
		{name: "if", ty: TK_IF},
		{name: "return", ty: TK_RETURN},
		{name: "int", ty: TK_INT},
		{name: "sizeof", ty: TK_SIZEOF},
	}
)

const (
	TK_NUM    = iota + 256 // Number literal
	TK_IDENT               // Identifier
	TK_INT                 // "int"
	TK_IF                  // "if"
	TK_ELSE                // "else"
	TK_FOR                 // "for"
	TK_LOGOR               // ||
	TK_LOGAND              // &&
	TK_RETURN              // "return"
	TK_SIZEOF              // "sizeof"
	TK_EOF                 // End marker
)

// Token type
type Token struct {
	ty    int    // Token type
	val   int    // Number literal
	name  string // Identifier
	input string // Token string (for error reporting)
}

type Keyword struct {
	name string
	ty   int
}

func add_token(v *Vector, ty int, input string) *Token {
	t := new(Token)
	t.ty = ty
	t.input = input
	vec_push(v, t)
	return t
}

func tokenize(s string) *Vector {

	v := new_vec()

loop:
	for len(s) != 0 {
		// Skip whitespace
		c := []rune(s)[0]
		if unicode.IsSpace(c) {
			s = s[1:]
			continue
		}

		// Multi-letter token or keywords
		for _, sym := range symbols {
			length := len(sym.name)
			if length > len(s) {
				length = len(s)
			}
			if strncmp(s, sym.name, length) != 0 {
				continue
			}
			add_token(v, sym.ty, s)
			s = s[length:]
			continue loop
		}

		// Single-letter token
		if strchr("+-*/;=(),{}<>[]&", c) != "" {
			add_token(v, int(c), s)
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
			t := add_token(v, TK_IDENT, s)
			t.name = strndup(s, length)
			s = s[length:]
			continue
		}

		// Number
		if unicode.IsDigit(c) {
			t := add_token(v, TK_NUM, s)
			i := 0
			cc := []rune(s)[i]
			for unicode.IsDigit(cc) {
				t.val = t.val*10 + (int(cc) - '0')
				i++
				cc = []rune(s)[i]
			}
			s = s[i:]
			continue
		}

		error("cannot tokenize: %s\n", string(c))
	}

	add_token(v, TK_EOF, s)
	return v
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
		case TK_INT:
			ty = "TK_INT   "
		case TK_IF:
			ty = "TK_IF    "
		case TK_ELSE:
			ty = "TK_ELSE  "
		case TK_FOR:
			ty = "TK_FOR   "
		case TK_RETURN:
			ty = "TK_RETURN"
		case TK_LOGOR:
			ty = "TK_LOGOR "
		case TK_LOGAND:
			ty = "TK_LOGAND"
		case TK_EOF:
			ty = "TK_EOF   "
		case ';':
			ty = ";        "
		case '+':
			ty = "+        "
		case '-':
			ty = "-        "
		case '*':
			ty = "*        "
		case '/':
			ty = "/        "
		case '(':
			ty = "(        "
		case ')':
			ty = ")        "
		case '{':
			ty = "{        "
		case '}':
			ty = "}        "
		case '<':
			ty = "<        "
		case '>':
			ty = ">        "
		default:
			ty = "         "
		}
		fmt.Printf("[%02d] ty: %s, val: %d, input: %s\n", i, ty, t.val, t.input)
	}
	fmt.Println("")
}
