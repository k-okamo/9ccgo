package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"unicode"
)

// Map
type Map struct {
	keys *Vector
	vals *Vector
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

func map_get(m *Map, key string) interface{} {
	for i := m.keys.len - 1; i >= 0; i-- {
		if m.keys.data[i].(string) == key {
			return m.vals.data[i]
		}
	}
	return 0
}

func map_exists(m *Map, key string) bool {
	for i := 0; i < m.keys.len; i++ {
		if m.keys.data[i] == key {
			return true
		}
	}
	return false
}

// Vector
type Vector struct {
	data     []interface{}
	capacity int
	len      int
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

func IsAlpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
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
	expect_test(file, line+1, 0, map_get(m, "foo").(int))

	map_put(m, "foo", 2)
	expect_test(file, line+4, 2, map_get(m, "foo").(int))

	map_put(m, "bar", 4)
	expect_test(file, line+7, 4, map_get(m, "bar").(int))

	map_put(m, "foo", 6)
	expect_test(file, line+10, 6, map_get(m, "foo").(int))

	expect_test_bool(file, line+12, true, map_exists(m, "foo"))
	expect_test_bool(file, line+13, false, map_exists(m, "baz"))
}

func util_test() {
	vec_test()
	map_test()
}
