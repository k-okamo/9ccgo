package main

// Atomic unit in the grammer is called "token".
// For example, `123`, `"abc"` and `while` are tokens.
// The tokenizer splits an inpuit string into tokens.
// Spaces and comments are removed by the tokenizer.

import (
	"fmt"
	"unicode"
)

var (
	tokens   *Vector
	keywords *Map
	symbols  = []Keyword{
		{name: "<<=", ty: TK_SHL_EQ},
		{name: ">>=", ty: TK_SHR_EQ},
		{name: "!=", ty: TK_NE},
		{name: "&&", ty: TK_LOGAND},
		{name: "++", ty: TK_INC},
		{name: "--", ty: TK_DEC},
		{name: "->", ty: TK_ARROW},
		{name: "<<", ty: TK_SHL},
		{name: "<=", ty: TK_LE},
		{name: "==", ty: TK_EQ},
		{name: ">=", ty: TK_GE},
		{name: ">>", ty: TK_SHR},
		{name: "||", ty: TK_LOGOR},
		{name: "*=", ty: TK_MUL_EQ},
		{name: "/=", ty: TK_DIV_EQ},
		{name: "%=", ty: TK_MOD_EQ},
		{name: "+=", ty: TK_ADD_EQ},
		{name: "-=", ty: TK_SUB_EQ},
		{name: "&=", ty: TK_BITAND_EQ},
		{name: "^=", ty: TK_XOR_EQ},
		{name: "|=", ty: TK_BITOR_EQ},
	}
	escaped = map[rune]int{
		'a': '\a',
		'b': '\b',
		'f': '\f',
		'n': '\n',
		'r': '\r',
		't': '\t',
		'v': '\v',
		'e': '\033',
		'E': '\033',
	}
)

const (
	TK_NUM       = iota + 256 // Number literal
	TK_STR                    // String literal
	TK_IDENT                  // Identifier
	TK_ARROW                  // ->
	TK_EXTERN                 // "extern"
	TK_TYPEDEF                // "typedef"
	TK_INT                    // "int"
	TK_CHAR                   // "char"
	TK_VOID                   // "void"
	TK_STRUCT                 // "struct"
	TK_IF                     // "if"
	TK_ELSE                   // "else"
	TK_FOR                    // "for"
	TK_DO                     // "do"
	TK_WHILE                  // "while"
	TK_BREAK                  // "break"
	TK_EQ                     // ==
	TK_NE                     // !=
	TK_LE                     // <=
	TK_GE                     // >=
	TK_LOGOR                  // ||
	TK_LOGAND                 // &&
	TK_SHL                    // <<
	TK_SHR                    // >>
	TK_INC                    // ++
	TK_DEC                    // --
	TK_MUL_EQ                 // *=
	TK_DIV_EQ                 // /=
	TK_MOD_EQ                 // %=
	TK_ADD_EQ                 // +=
	TK_SUB_EQ                 // -=
	TK_SHL_EQ                 // <<=
	TK_SHR_EQ                 // >>=
	TK_BITAND_EQ              // &=
	TK_XOR_EQ                 // ^=
	TK_BITOR_EQ               // |=
	TK_RETURN                 // "return"
	TK_SIZEOF                 // "sizeof"
	TK_ALIGNOF                // "_Alignof"
	TK_EOF                    // End marker
)

// Token type
type Token struct {
	ty    int    // Token type
	val   int    // Number literal
	name  string // Identifier
	input string // Token string (for error reporting)

	// String literal
	str string
	len int
}

type Keyword struct {
	name string
	ty   int
}

func read_char(result *int, s string) string {

	i := 0
	c := []rune(s)[0]
	if c != '\\' {
		*result = int(c)
		i++
		c = []rune(s)[i]
	} else {
		i++
		c = []rune(s)[i]
		if i == len(s) {
			error("premature end of input")
		}
		esc, ok := escaped[c]
		if ok {
			*result = esc
		} else {
			*result = int(c)
		}
		i++
		c = []rune(s)[i]
	}

	if c != '\'' {
		error("unclosed character literal")
	}
	i++

	return s[i:]
}

func read_string(sb *StringBuilder, s string) string {
	i := 0
	c := []rune(s)[0]
	for c != '"' {
		if i == len(s) {
			error("premature end of input")
		}

		if c != '\\' {
			sb_add(sb, string(c))
			i++
			c = []rune(s)[i]
			continue
		}

		// c == '\\'
		i++
		c = []rune(s)[i]
		esc, ok := escaped[c]
		if ok {
			sb_add(sb, string(esc))
		} else {
			sb_add(sb, string(c))
		}
		i++
		c = []rune(s)[i]
	}
	return s[(i + 1):]
}

func add_token(v *Vector, ty int, input string) *Token {
	t := new(Token)
	t.ty = ty
	t.input = input
	vec_push(v, t)
	return t
}

func keyword_map() *Map {
	kmap := new_map()
	map_puti(kmap, "_Alignof", TK_ALIGNOF)
	map_puti(kmap, "break", TK_BREAK)
	map_puti(kmap, "char", TK_CHAR)
	map_puti(kmap, "do", TK_DO)
	map_puti(kmap, "else", TK_ELSE)
	map_puti(kmap, "extern", TK_EXTERN)
	map_puti(kmap, "for", TK_FOR)
	map_puti(kmap, "if", TK_IF)
	map_puti(kmap, "int", TK_INT)
	map_puti(kmap, "return", TK_RETURN)
	map_puti(kmap, "sizeof", TK_SIZEOF)
	map_puti(kmap, "struct", TK_STRUCT)
	map_puti(kmap, "typedef", TK_TYPEDEF)
	map_puti(kmap, "void", TK_VOID)
	map_puti(kmap, "while", TK_WHILE)
	return kmap
}

func tokenize(s string) *Vector {
	v := new_vec()
	keywords := keyword_map()

loop:
	for len(s) != 0 {
		// Skip whitespace
		c := []rune(s)[0]
		if unicode.IsSpace(c) {
			s = s[1:]
			continue
		}

		// Line comment
		if strncmp(s, "//", 2) == 0 {
			i := 0
			for i != len(s) && c != '\n' {
				i++
				c = []rune(s)[i]
			}
			s = s[i:]
			continue
		}

		// Block comment
		if strncmp(s, "/*", 2) == 0 {
			for s = s[2:]; len(s) != 0; s = s[1:] {
				if strncmp(s, "*/", 2) != 0 {
					continue
				}
				s = s[2:]
				continue loop
			}
			error("unclosed comment")
		}

		// Character literal
		if c == '\'' {
			t := add_token(v, TK_NUM, s)
			s = s[1:]
			s = read_char(&t.val, s)
			continue
		}

		// String literal
		if c == '"' {
			t := add_token(v, TK_STR, s)
			s = s[1:]

			sb := new_sb()
			s = read_string(sb, s)
			t.str = sb_get(sb)
			t.len = sb.len
			continue
		}

		// Multi-letter symbol
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

		// Single-letter symbol
		if strchr("+-*/;=(),{}<>[]&.!?:|^%", c) != "" {
			add_token(v, int(c), s)
			s = s[1:]
			continue
		}

		// Keyword or identifier
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
			ty := map_geti(keywords, name, -1)

			var t *Token
			if ty == -1 {
				t = add_token(v, TK_IDENT, s)
			} else {
				t = add_token(v, ty, s)
			}
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
		case TK_STR:
			ty = "TK_STR   "
		case TK_IDENT:
			ty = "TK_IDENT "
		case TK_INT:
			ty = "TK_INT   "
		case TK_CHAR:
			ty = "TK_CHAR  "
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
		case TK_SIZEOF:
			ty = "TK_SIZEOF"
		case TK_STRUCT:
			ty = "TK_STRUCT"
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
		case '=':
			ty = "=        "
		case ',':
			ty = ",        "
		case '(':
			ty = "(        "
		case ')':
			ty = ")        "
		case '{':
			ty = "{        "
		case '}':
			ty = "}        "
		case '[':
			ty = "[        "
		case ']':
			ty = "]        "
		case '&':
			ty = "&        "
		case '<':
			ty = "<        "
		case '>':
			ty = ">        "
		case '.':
			ty = ".        "
		default:
			ty = "         "
		}
		fmt.Printf("[%02d] ty: %s, val: %d, input: %s", i, ty, t.val, t.input)
	}
	fmt.Println("")
}
