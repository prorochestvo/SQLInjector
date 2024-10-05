package expression

import (
	"fmt"
	"testing"
)

var _ expression = &Limit{}

func TestNewLimitFrom(t *testing.T) {

	tests := []struct {
		expr     string
		expected []*Limit
		err      bool
	}{

		{"", []*Limit{}, false},
		{"0", []*Limit{{limit: 0}}, false},
		{"1", []*Limit{{limit: 1}}, false},
		{"10", []*Limit{{limit: 10}}, false},
		{"-1", nil, true},
		{"abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("expr=%s", tt.expr), func(t *testing.T) {
			result, err := NewLimitFrom(tt.expr)
			if (err != nil) != tt.err {
				t.Errorf("unexpected error status: got %v, expected %v", err != nil, tt.err)
			}
			if !tt.err && len(result) != len(tt.expected) {
				t.Errorf("unexpected result length: got %d, expected %d", len(result), len(tt.expected))
			}
			if !tt.err && len(result) > 0 && result[0].limit != tt.expected[0].limit {
				t.Errorf("unexpected limit: got %d, expected %d", result[0].limit, tt.expected[0].limit)
			}
		})
	}
}

func TestNewLimit(t *testing.T) {
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
			result := NewLimit(tt.value)
			if result.limit != tt.expected {
				t.Errorf("unexpected limit: got %d, expected %d", result.limit, tt.expected)
			}
		})
	}
}
