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
		{name: "_Alignof", ty: TK_ALIGNOF},
		{name: "break", ty: TK_BREAK},
		{name: "char", ty: TK_CHAR},
		{name: "do", ty: TK_DO},
		{name: "else", ty: TK_ELSE},
		{name: "extern", ty: TK_EXTERN},
		{name: "for", ty: TK_FOR},
		{name: "if", ty: TK_IF},
		{name: "int", ty: TK_INT},
		{name: "return", ty: TK_RETURN},
		{name: "sizeof", ty: TK_SIZEOF},
		{name: "struct", ty: TK_STRUCT},
		{name: "typedef", ty: TK_TYPEDEF},
		{name: "void", ty: TK_VOID},
		{name: "while", ty: TK_WHILE},
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
	TK_NUM     = iota + 256 // Number literal
	TK_STR                  // String literal
	TK_IDENT                // Identifier
	TK_ARROW                // ->
	TK_EXTERN               // "extern"
	TK_TYPEDEF              // "typedef"
	TK_INT                  // "int"
	TK_CHAR                 // "char"
	TK_VOID                 // "void"
	TK_STRUCT               // "struct"
	TK_IF                   // "if"
	TK_ELSE                 // "else"
	TK_FOR                  // "for"
	TK_DO                   // "do"
	TK_WHILE                // "while"
	TK_BREAK                // "break"
	TK_EQ                   // ==
	TK_NE                   // !=
	TK_LE                   // <=
	TK_GE                   // >=
	TK_LOGOR                // ||
	TK_LOGAND               // &&
	TK_SHL                  // <<
	TK_SHR                  // >>
	TK_INC                  // ++
	TK_DEC                  // --
	TK_RETURN               // "return"
	TK_SIZEOF               // "sizeof"
	TK_ALIGNOF              // "_Alignof"
	TK_EOF                  // End marker
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

/*
func read_string(sb *StringBuilder, s string) int {

	i := 0
	c := []rune(s)[0]
	for i < len(s) && c != '"' {
		if i == len(s) {
			error("premature end of input")
		}
		if c != '\\' {
			sb_add(sb, string(c))
			i++
			c = []rune(s)[i]
			continue
		}

		i++
		c = []rune(s)[i]
		switch {
		case c == 'a':
			sb_add(sb, "\a")
		case c == 'b':
			sb_add(sb, "\b")
		case c == 'f':
			sb_add(sb, "\f")
		case c == 'n':
			sb_add(sb, "\n")
		case c == 'r':
			sb_add(sb, "\r")
		case c == 't':
			sb_add(sb, "\t")
		case c == 'v':
			sb_add(sb, "\v")
		case c == '0':
			error("premature end of input.")
		default:
			sb_add(sb, s)
		}
		i++
	}
	return i + 1
}
*/

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
		if strchr("+-*/;=(),{}<>[]&.!?:|^%", c) != "" {
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
