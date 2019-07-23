package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode"
	"unsafe"
)

func new_sb() *StringBuilder {
	sb := new(StringBuilder)
	sb.data = ""
	sb.capacity = 8
	sb.len = 0
	return sb
}

func sb_grow(sb *StringBuilder, len int) {
	if sb.len+len <= sb.capacity {
		return
	}
	for sb.len+len > sb.capacity {
		sb.capacity *= 2
	}
}

func sb_add(sb *StringBuilder, s string) {
	sb_grow(sb, 1)
	sb.data += s
	sb.len += len(s)
}

func sb_append(sb *StringBuilder, s string) {
	sb_append_n(sb, s, len(s))
}

func sb_append_n(sb *StringBuilder, s string, len int) {
	sb_grow(sb, len)
	sb.data += s
	sb.len += len
}

func sb_get(sb *StringBuilder) string {
	return sb.data
}

func roundup(x, align int) int {
	return (x + align - 1) & ^(align - 1)
}

func ptr_to(base *Type) *Type {
	ty := new(Type)
	ty.ty = PTR
	ty.size = 8
	ty.align = 8
	ty.ptr_to = base
	return ty
}

func ary_of(base *Type, length int) *Type {
	ty := new(Type)
	ty.ty = ARY
	ty.size = base.size * length
	ty.align = base.align
	ty.ary_of = base
	ty.len = length
	return ty
}

func size_of(ty *Type) int {
	if ty.ty == CHAR {
		return 1
	}
	if ty.ty == INT {
		return 4
	}
	if ty.ty == PTR {
		return 8
	}
	// assert(ty.ty == ARY)
	return size_of(ty.ary_of) * ty.len
}

func align_of(ty *Type) int {
	if ty.ty == CHAR {
		return 1
	}
	if ty.ty == INT {
		return 4
	}
	if ty.ty == PTR {
		return 8
	}
	// assert(ty.ty == ARY)
	return align_of(ty.ary_of)
}

func copy_node(src, dst *Node) {
	if src == nil {
		return
	}

	// value
	dst.op = src.op
	dst.val = src.val
	dst.data = src.data
	dst.len = src.len
	dst.name = src.name
	dst.stacksize = src.stacksize
	dst.offset = src.offset
	dst.is_extern = src.is_extern

	// Node
	copy_node(src.lhs, dst.lhs)
	copy_node(src.rhs, dst.rhs)
	copy_node(src.expr, dst.expr)
	copy_node(src.cond, dst.cond)
	copy_node(src.then, dst.then)
	copy_node(src.els, dst.els)
	copy_node(src.init, dst.init)
	copy_node(src.body, dst.body)

	// Type
	copy_type(src.ty, dst.ty)

	// Vector
	copy_vector(src.stmts, dst.stmts)
	copy_vector(src.globals, dst.globals)
	copy_vector(src.args, dst.args)
}

func copy_type(src, dst *Type) {
	if src == nil {
		return
	}

	// value
	dst.ty = src.ty
	dst.len = src.len
	dst.size = src.size
	dst.align = src.align
	dst.offset = src.offset

	// Type
	copy_type(src.ptr_to, dst.ptr_to)
	copy_type(src.ary_of, dst.ary_of)

	// Vector
	copy_vector(src.members, dst.members)
}

func copy_vector(src, dst *Vector) {
	if src == nil {
		return
	}

	// value
	dst.len = src.len
	dst.capacity = src.capacity
	dst.data = make([]interface{}, dst.capacity, dst.len)
	for i := range src.data {
		dst.data[i] = src.data[i]
	}
}

func new_map() *Map {
	m := new(Map)
	m.keys = new_vec()
	m.vals = new_vec()
	return m
}

func map_put(m *Map, key string, val interface{}) {
	vec_push(m.keys, key)
	vec_push(m.vals, val)
}

func map_puti(m *Map, key string, val int) {
	map_put(m, key, val)
}

func map_get(m *Map, key string) interface{} {
	for i := m.keys.len - 1; i >= 0; i-- {
		if m.keys.data[i].(string) == key {
			return m.vals.data[i]
		}
	}
	return nil
}

func map_geti(m *Map, key string, default_ int) int {
	for i := m.keys.len - 1; i >= 0; i-- {
		if m.keys.data[i].(string) == key {
			return m.vals.data[i].(int)
		}
	}
	return default_
}

func format(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func new_vec() *Vector {
	v := new(Vector)
	v.data = make([]interface{}, 16)
	v.capacity = 16
	v.len = 0
	return v
}

func vec_push(v *Vector, elem interface{}) {
	if v.len == v.capacity {
		v.data = append(v.data, make([]interface{}, v.capacity)...)
		v.capacity *= 2
	}
	v.data[v.len] = elem
	v.len++
}

// An error reporting function
func error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func popcount(x uint) int {
	ret := 0
	for n := uint(0); n < uint(unsafe.Sizeof(x))*8; n++ {
		if x&(1<<n) != 0 {
			ret++
		}
	}
	return ret
}

func ctz(x uint) int {
	ret := 0
	a := uint(1)
	for n := uint(0); n < uint(unsafe.Sizeof(x))*8; n++ {
		if x&a != 0 {
			return ret
		}
		a = a * 2
		ret++
	}
	return ret
}

func strtol(s string, b int) (int, string) {
	if !unicode.IsDigit([]rune(s)[0]) {
		return 0, s
	}

	j := len(s)
	for i, c := range s {
		if !unicode.IsDigit(c) {
			j = i
			break
		}
	}
	n, _ := strconv.ParseInt(s[:j], b, 32)
	return int(n), s[j:]

}

func strchr(s string, c rune) string {
	for i, r := range s {
		if c == r {
			return s[i:]
		}
	}
	return ""
}

func strndup(s string, size int) string {
	if len(s) <= size {
		return s
	}
	return s[:size]
}

func strncmp(s1, s2 string, n int) int {
	if n == 0 || s1 == s2 {
		return 0
	}
	switch {
	case s1 == "":
		return -1
	case s2 == "":
		return 1
	case s1[:1] > s2[:1]:
		return 1
	case s1[:1] < s2[:1]:
		return -1
	}
	return strncmp(s1[1:], s2[1:], n-1)
}

func strncasecmp(s1, s2 string, n int) int {
	return strncmp(strings.ToUpper(s1), strings.ToUpper(s2), n)
}

func isgraph(c rune) bool {
	return 0x21 <= c && c <= 0x7e
}

func isprint(c rune) bool {
	return 0x20 <= c && c <= 0x7e
}

func isalpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isxdigit(s string) bool {
	_, err := strconv.ParseInt(s, 16, 64)
	return err == nil
}

// Testing
func expect_test(file string, line, expected, actual int) {
	if expected == actual {
		return
	}
	fmt.Fprintf(os.Stderr, "%s:%d: %d expected, but got %d\n", file, line, expected, actual)
	os.Exit(1)
}

func expect_test_bool(file string, line int, expected, actual bool) {
	if expected == actual {
		return
	}
	fmt.Fprintf(os.Stderr, "%s:%d: %v expected, but got %v\n", file, line, expected, actual)
	os.Exit(1)
}

func vec_test() {
	vec := new_vec()
	_, file, line, _ := runtime.Caller(0)
	expect_test(file, line+1, 0, vec.len)

	for i := 0; i < 100; i++ {
		vec_push(vec, i)
	}

	expect_test(file, line+7, 100, vec.len)
	expect_test(file, line+8, 0, vec.data[0].(int))
	expect_test(file, line+9, 50, vec.data[50].(int))
	expect_test(file, line+10, 99, vec.data[99].(int))
}

func map_test() {
	m := new_map()
	_, file, line, _ := runtime.Caller(0)
	//expect_test(file, line+1, 0, map_get(m, "foo").(int))

	map_put(m, "foo", 2)
	expect_test(file, line+4, 2, map_get(m, "foo").(int))

	map_put(m, "bar", 4)
	expect_test(file, line+7, 4, map_get(m, "bar").(int))

	map_put(m, "foo", 6)
	expect_test(file, line+10, 6, map_get(m, "foo").(int))
}

func sb_test() {
	sb1 := new_sb()
	_, file, line, _ := runtime.Caller(0)
	expect_test(file, line+1, 0, len(sb_get(sb1)))

	sb2 := new_sb()
	sb_append(sb2, "foo")
	expect_test_bool(file, line+5, true, sb_get(sb2) == "foo")

	sb3 := new_sb()
	sb_append(sb3, "foo")
	sb_append(sb3, "bar")
	expect_test_bool(file, line+10, true, sb_get(sb3) == "foobar")

	sb4 := new_sb()
	sb_append(sb4, "foo")
	sb_append(sb4, "bar")
	sb_append(sb4, "foo")
	sb_append(sb4, "bar")
	expect_test_bool(file, line+17, true, sb_get(sb4) == "foobarfoobar")

}

func util_test() {
	vec_test()
	map_test()
	sb_test()
}
