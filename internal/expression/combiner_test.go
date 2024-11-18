package expression

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ expression = &Or{}

func TestCombiner_NewCombinerOR(t *testing.T) {
	w1 := NewWhere("age", ">", "28")
	w2 := NewWhere("name", "==", "Tom")
	w3 := NewWhere("salary", ">", "5000")
	w4 := NewWhere("city", "==", "New York")

	or1 := NewCombinerOR(w1, w2)
	require.Equal(t, []*Where{w1, w2}, or1.items)

	or2 := NewCombinerOR(w1, w2, w3, w4)
	require.Equal(t, []*Where{w1, w2, w3, w4}, or2.items)
}

func TestCombiner_Or(t *testing.T) {
	w1 := NewWhere("age", ">", "28")
	w2 := NewWhere("name", "==", "Tom")
	w3 := NewWhere("salary", ">", "5000")
	w4 := NewWhere("city", "==", "New York")

	or := NewCombinerOR(w1, w2, w3, w4)
	actually := or.Or()
	expected := []*Where{w1, w2, w3, w4}
	require.Equal(t, expected, actually)
}

func TestCombiner_QueryMod(t *testing.T) {
	t.Skip("ignore at the moment")
}

func TestCombiner_ToString(t *testing.T) {
	w1 := NewWhere("age", ">", "28")
	w2 := NewWhere("name", "==", "Tom")
	w3 := NewWhere("salary", ">", "5000")
	w4 := NewWhere("city", "==", "New York")

	or := NewCombinerOR(w1, w2, w3, w4)

	expected := fmt.Sprintf("(%s) Or (%s) Or (%s) Or (%s)", w1.ToString(), w2.ToString(), w3.ToString(), w4.ToString())
	actually := or.ToString()
	require.Equal(t, expected, actually)
}
