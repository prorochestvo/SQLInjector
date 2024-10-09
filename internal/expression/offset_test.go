package expression

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ expression = &Offset{}

func TestNewOffsetFrom(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		result, err := NewOffsetFrom("")
		require.NoError(t, err)
		require.Len(t, result, 0)
	})
	t.Run("PositiveNumber", func(t *testing.T) {
		result, err := NewOffsetFrom("1")
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, 1, result[0].offset)
	})
	t.Run("NegativeNumber", func(t *testing.T) {
		_, err := NewOffsetFrom("-1")
		require.Error(t, err)
	})
	t.Run("ZeroValue", func(t *testing.T) {
		result, err := NewOffsetFrom("0")
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, 0, result[0].offset)
	})
	t.Run("UnexpectedSymbols", func(t *testing.T) {
		_, err := NewOffsetFrom("abc")
		require.Error(t, err)
	})
	t.Run("NonDecimalBase", func(t *testing.T) {
		_, err := NewOffsetFrom("1.5")
		require.Error(t, err)
	})
}

func TestNewOffset(t *testing.T) {
	testCases := []struct {
		value    int
		expected int
	}{
		{-10, 0},
		{0, 0},
		{1, 1},
		{100, 100},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("TestCase:%d", testCase.value), func(t *testing.T) {
			r := NewOffset(testCase.value)
			require.Equal(t, testCase.expected, r.offset)
		})
	}
}
