package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

var (
	input_file string
	tokens     *Vector
	keywords   *Map
	symbols    = []Keyword{
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

type Keyword struct {
	name string
	ty   int
}

// Finds a line pointed by a given pointer from the input line
// to print it out.
func print_line(pos string) {
	curline, start := input_file, input_file
	line, col := 0, 0

	for i, c := range input_file {

		if c == '\n' {
			curline = input_file[i+1:]
			line++
			col = 0
			//start = start[i:]
			start = input_file[i+1:]
			continue
		}

		if start != pos {
			col++
			start = input_file[i+1:]
			continue
		}

		fmt.Fprintf(os.Stderr, "error at %s:%d:%d\n\n", filename, line+1, col+1)
		for i, c2 := range curline {
			if c2 == '\n' {
				curline = curline[:i]
				break
			}
		}
		fmt.Fprintf(os.Stderr, "%s\n", curline)

		for i := 0; i < col-1; i++ {
			fmt.Fprintf(os.Stderr, " ")
		}
		fmt.Fprintf(os.Stderr, "^\n\n")
		return
	}
}

func bad_token(t *Token, msg string) {
	print_line(t.start)
	error(msg)
}

// Atomic unit in the grammer is called "token".
// For example, `123`, `"abc"` and `while` are tokens.
// The tokenizer splits an inpuit string into tokens.
// Spaces and comments are removed by the tokenizer.

func add_t(ty int, start string) *Token {
	t := new(Token)
	t.ty = ty
	t.start = start
	vec_push(tokens, t)
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

func block_comment(pos string) string {
	for s := pos[2:]; len(s) != 0; s = s[1:] {
		if strncmp(s, "*/", 2) != 0 {
			return s[2:]
		}
	}
	print_line(pos)
	error("unclosed comment")
	return ""
}

func char_literal(p string) string {
	t := add_t(TK_NUM, p)
	p = p[1:]

	if len(p) == 0 {
		goto err
	}

	if rune(p[0]) != '\\' {
		t.val = int(p[0])
		p = p[1:]
	} else {
		if len(p) < 2 {
			goto err
		}
		esc := escaped[rune(p[1])]
		if esc != 0 {
			t.val = esc
		} else {
			t.val = int(p[1])
		}
		p = p[2:]
	}

	if p[0] == '\'' {
		return p[1:]
	}

err:
	bad_token(t, "unclosed character literal")
	return ""
}

func string_literal(p string) string {

	t := add_t(TK_STR, p)
	p = p[1:]
	sb := new_sb()

	for rune(p[0]) != '"' {
		if len(p) == 0 {
			goto err
		}

		if p[0] != '\\' {
			sb_add(sb, string(p[0]))
			p = p[1:]
			continue
		}

		p = p[1:]
		if len(p) == 0 {
			goto err
		}
		esc := escaped[rune(p[0])]
		if esc != 0 {
			sb_add(sb, string(esc))
		} else {
			sb_add(sb, string(p[0]))
		}
		p = p[1:]
	}

	t.str = sb_get(sb)
	t.len = sb.len
	return p[1:]

err:
	bad_token(t, "unclosed string literal")
	return ""
}

func ident_t(p string) string {
	len := 1
	for isalpha(rune(p[len])) || unicode.IsDigit(rune(p[len])) || p[len] == '_' {
		len++
	}

	name := strndup(p, len)
	ty := map_geti(keywords, name, TK_IDENT)
	t := add_t(ty, p)
	t.name = name
	return p[len:]
}

func hexadecimal(p string) string {
	t := add_t(TK_NUM, p)
	p = p[2:]

	if !isxdigit(string(p[0])) {
		bad_token(t, "bad hexadecimal number")
	}

	for {
		c := int(p[0])
		if '0' <= c && c <= '9' {
			t.val = t.val*16 + c - '0'
			p = p[1:]
		} else if 'a' <= c && c <= 'f' {
			t.val = t.val*16 + c - 'a' + 10
			p = p[1:]
		} else if 'A' <= c && c <= 'F' {
			t.val = t.val*16 + c - 'A' + 10
			p = p[1:]
		} else {
			return p
		}
	}
	return ""
}

func octal(p string) string {
	t := add_t(TK_NUM, p)
	p = p[1:]

	c := p[0]
	for '0' <= c && c <= '7' {
		t.val = t.val*8 + int(c) - '0'
		p = p[1:]
		c = p[0]
	}
	return p
}

func decimal(p string) string {
	t := add_t(TK_NUM, p)
	for unicode.IsDigit(rune(p[0])) {
		t.val = t.val*10 + int(p[0]) - '0'
		p = p[1:]
	}
	return p
}

func number(p string) string {
	if strncasecmp(p, "0x", 2) == 0 {
		return hexadecimal(p)
	}
	if p[0] == '0' {
		return octal(p)
	}
	return decimal(p)
}

// Tokenized input is stored to this array
func scan() {
	p := input_file

loop:
	for len(p) != 0 {
		c := rune(p[0])
		if unicode.IsSpace(c) {
			p = p[1:]
			continue
		}

		// Line comment
		if strncmp(p, "//", 2) == 0 {
			for len(p) != 0 && c != '\n' {
				p = p[1:]
				c = rune(p[0])
			}
			continue
		}

		// Block comment
		if strncmp(p, "/*", 2) == 0 {
			p = block_comment(p)
			continue
		}

		// Character literal
		if c == '\'' {
			p = char_literal(p)
			continue
		}

		// String literal
		if c == '"' {
			p = string_literal(p)
			continue
		}

		// Multi-letter symbol
		for _, sym := range symbols {
			length := len(sym.name)
			if length > len(p) {
				length = len(p)
			}
			if strncmp(p, sym.name, length) != 0 {
				continue
			}
			add_t(sym.ty, p)
			p = p[length:]
			continue loop
		}

		// Single-letter symbol
		if strchr("+-*/;=(),{}<>[]&.!?:|^%~", c) != "" {
			add_t(int(c), p)
			p = p[1:]
			continue
		}

		// Keyword or identifier
		if isalpha(c) || c == '_' {
			p = ident_t(p)
			continue
		}

		// Number
		if unicode.IsDigit(c) {
			p = number(p)
			continue
		}

		print_line(p)
		error("cannot tokenize")
	}

	add_t(TK_EOF, p)
}

func canonicalize_newline() {
	input_file = strings.Replace(input_file, "\r\n", "\n", -1)
}

func remove_backslash_newline() {
	input_file = strings.Replace(input_file, "\\\n", "", -1)
}

func append_t(x, y *Token) {
	sb := new_sb()
	sb_append_n(sb, x.str, x.len)
	sb_append_n(sb, y.str, y.len)
	x.str = sb_get(sb)
	x.len = sb.len
}

func join_string_literals() {
	v := new_vec()
	var last *Token

	for i := 0; i < tokens.len; i++ {
		t := tokens.data[i].(*Token)
		if last != nil && last.ty == TK_STR && t.ty == TK_STR {
			append_t(last, t)
			continue
		}

		last = t
		vec_push(v, t)
	}
	tokens = v
}

func tokenize(p string) *Vector {
	tokens = new_vec()
	keywords = keyword_map()
	input_file = p

	canonicalize_newline()
	remove_backslash_newline()
	scan()
	join_string_literals()
	return tokens
}
