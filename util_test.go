package main

import (
	"testing"
)

func Test_strtol(t *testing.T) {
	cases := []struct {
		str  string
		ret  int
		str2 string
	}{
		{"123", 123, ""},
		{"123a", 123, "a"},
		{"a123", 0, "a123"},
	}

	for _, c := range cases {
		n, s := strtol(c.str, 10)
		if n != c.ret || s != c.str2 {
			t.Errorf("expected (%d, %s), got (%d, %s)\n", c.ret, c.str2, n, s)
		}
	}
}

func Test_strndup(t *testing.T) {
	cases := []struct {
		str  string
		size int
		ret  string
	}{
		{"abcde", 4, "abcd"},
		{"abcde", 5, "abcde"},
		{"abcde", 6, "abcde"},
		{"", 1, ""},
	}

	for _, c := range cases {
		ret := strndup(c.str, c.size)
		if ret != c.ret {
			t.Errorf("expected: %s, got: %s\n", c.ret, ret)
		}
	}
}
