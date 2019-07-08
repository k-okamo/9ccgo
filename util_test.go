package main

// Unit tests for out data structures
//
// This kind of file is usually built as an independent executable in
// a common build config, but in 9ccgo I took a different approach.
// This file is just a part of the main executable. This scheme greatly
// simplifies build config.
//
// In return for the simplicity, the main execurable becomes slightly
// larger, but that's not a problem for top programs like 9ccgo.
// What is most important is to write tests while keeping everything simple.

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

func Test_strncmp(t *testing.T) {
	cases := []struct {
		s1  string
		s2  string
		n   int
		ret int
	}{
		{"ABC", "ABD", 2, 0},
		{"ABC", "ABC", 2, 0},
		{"ABC", "AAA", 2, 1},
		{"ABC", "ABCD", 2, 0},
		{"ABC", "AB", 2, 0},
		{"ABC", "B", 2, -1},
		{"ABC", "A", 2, 1},
	}

	for _, c := range cases {
		ret := strncmp(c.s1, c.s2, c.n)
		if ret != c.ret {
			t.Errorf("s1: %s, s2: %s, n: %d, expecred %d, got: %d\n", c.s1, c.s2, c.n, c.ret, ret)
		}
	}
}

func Test_isgraph(t *testing.T) {
	cases := []struct {
		c   rune
		ret bool
	}{
		{'a', true},
		{'A', true},
		{'1', true},
		{' ', false},
		{'\t', false},
	}

	for _, c := range cases {
		ret := isgraph(c.c)
		if ret != c.ret {
			t.Errorf("c: %s, expected: %v, got: %v\n", string(c.c), ret, c.ret)
		}
	}
}

func Test_popcount(t *testing.T) {
	cases := []struct {
		x   uint
		ret int
	}{
		{1, 1},
		{2, 1},
		{8, 1},
		{32, 1},
		{128, 1},
		{3, 2},
	}

	for _, c := range cases {
		ret := popcount(c.x)
		if ret != c.ret {
			t.Errorf("expected: %d, got: %d\n", c.ret, ret)
		}
	}
}

func Test_ctz(t *testing.T) {
	cases := []struct {
		x   uint
		ret int
	}{
		{1, 0},
		{2, 1},
		{3, 0},
		{8, 3},
	}

	for _, c := range cases {
		ret := ctz(c.x)
		if ret != c.ret {
			t.Errorf("expected: %d, got: %d\n", c.ret, ret)
		}
	}
}
