package main

import (
	"bytes"
	"testing"
)

type offsetTest struct {
	data       []byte
	offset     int
	byteOffset int
}

var offsetTests = []offsetTest{
	offsetTest{[]byte("abcdef"), 0, 0},
	offsetTest{[]byte("abcdef"), 1, 1},
	offsetTest{[]byte("abcdef"), 5, 5},
	offsetTest{[]byte("日本語def"), 0, 0},
	offsetTest{[]byte("日本語def"), 1, 3},
	offsetTest{[]byte("日本語def"), 5, 11},
}

func TestByteOffset(t *testing.T) {
	for _, test := range offsetTests {
		off, err := byteOffset(bytes.NewReader(test.data), test.offset)
		if err != nil {
			t.Errorf("got error %v", err)
		}
		if off != test.byteOffset {
			t.Errorf("expected byte offset %d, got %d", test.byteOffset, off)
		}
	}
}
