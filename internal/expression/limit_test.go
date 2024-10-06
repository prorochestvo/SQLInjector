package expression

import (
	"fmt"
	"github.com/stretchr/testify/require"
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
			require.Equal(t, err, tt.err)
			require.NoError(t, err)
			if !tt.err {
				require.Equal(t, len(result), len(tt.expected))
				if len(result) > 0 {
					require.Equal(t, result[0].limit, tt.expected[0].limit)
				}
			}
		})
	}

	t.Run("DefaultBehavior", func(t *testing.T) {
	})
	t.Run("NegativeNumber", func(t *testing.T) {
	})
	t.Run("UnexpectedSymbols", func(t *testing.T) {
	})
	t.Run("NonDecimalBase", func(t *testing.T) {
	})
}

func TestNewLimit(t *testing.T) {
	testCases := []struct {
		value    int
		expected int
	}{
		{-10, 0},
		{0, 0},
		{1, 1},
		{100, 100},
		{0xFFFFFF, 0xFFFFFF},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("TestCases%d", testCase.value), func(t *testing.T) {
			r := NewLimit(testCase.value)
			// TODO: REVIEW: could use this approach to test the value in all our tests
			require.Equal(t, testCase.expected, r.limit)
		})
	}
}
