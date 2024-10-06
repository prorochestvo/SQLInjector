package expression

import (
	"fmt"
	"testing"
)

var _ expression = &Offset{}

func TestNewOffsetFrom(t *testing.T) {
	tests := []struct {
		expr     string
		expected []*Offset
		err      bool
	}{

		{"", []*Offset{}, false},
		{"0", []*Offset{{offset: 0}}, false},
		{"1", []*Offset{{offset: 1}}, false},
		{"100", []*Offset{{offset: 100}}, false},
		{"-1", nil, true},
		{"abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("expr=%s", tt.expr), func(t *testing.T) {
			result, err := NewOffsetFrom(tt.expr)
			if (err != nil) != tt.err {
				t.Errorf("unexpected error status: got %v, expected %v", err != nil, tt.err)
			}
			if !tt.err && len(result) != len(tt.expected) {
				t.Errorf("unexpected result length: got %d, expected %d", len(result), len(tt.expected))
			}
			if !tt.err && len(result) > 0 && result[0].offset != tt.expected[0].offset {
				t.Errorf("unexpected offset: got %d, expected %d", result[0].offset, tt.expected[0].offset)
			}
		})
	}
}

func TestNewOffset(t *testing.T) {
	tests := []struct {
		value    int
		expected int
	}{

		{-10, 0},
		{0, 0},
		{1, 1},
		{100, 100},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("value=%d", tt.value), func(t *testing.T) {
			result := NewOffset(tt.value)
			if result.offset != tt.expected {
				t.Errorf("unexpected offset: got %d, expected %d", result.offset, tt.expected)
			}
		})
	}
}
