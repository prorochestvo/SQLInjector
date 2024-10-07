package expression

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ expression = &Limit{}

func TestNewLimitFrom(t *testing.T) {

	t.Run("EmptyString", func(t *testing.T) {
		result, err := NewLimitFrom("") // This string
		require.NoError(t, err)
		require.Equal(t, 0, len(result))
	})
	t.Run("DefaultBehavior", func(t *testing.T) {
		result, err := NewLimitFrom("1")

		require.NoError(t, err)
		require.Equal(t, 1, len(result))
		require.Equal(t, 1, result[0].limit)
	})
	t.Run("NegativeNumber", func(t *testing.T) {
		_, err := NewLimitFrom("-1")
		require.Error(t, err)
	})
	t.Run("ZeroValue", func(t *testing.T) {
		result, err := NewLimitFrom("0")
		require.NoError(t, err)
		require.Equal(t, 1, len(result))
		require.Equal(t, 0, result[0].limit)
	})
	t.Run("UnexpectedSymbols", func(t *testing.T) {
		_, err := NewLimitFrom("abc")
		require.Error(t, err)
	})
	t.Run("NonDecimalBase", func(t *testing.T) {
		_, err := NewLimitFrom("1.5")
		require.Error(t, err)
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
			require.Equal(t, testCase.expected, r.limit)
		})
	}
}
