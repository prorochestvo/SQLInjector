package expression

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var _ expression = &Or{}

func TestNewCombinerOR(t *testing.T) {

	w1 := NewWhere("age", ">", "28")
	w2 := NewWhere("name", "==", "Tom")
	w3 := NewWhere("salary", ">", "5000")
	w4 := NewWhere("city", "==", "New York")

	or1 := NewCombinerOR(w1, w2)
	require.Equal(t, 2, len(or1.items), "The length of or1.items should be 2")

	expectedItems1 := []string{
		w1.ToString(),
		w2.ToString(),
	}

	for i, item := range or1.items {
		require.Equal(t, expectedItems1[i], item.ToString(), "The item string should match the expected value")
	}

	or2 := NewCombinerOR(w1, w2, w3, w4)
	require.Equal(t, 4, len(or2.items), "The length of or2.items should be 4")

	expectedItems2 := []string{
		w1.ToString(),
		w2.ToString(),
		w3.ToString(),
		w4.ToString(),
	}

	for i, item := range or2.items {
		require.Equal(t, expectedItems2[i], item.ToString(), "The item string should match the expected value")
	}
}

func TestCombiner_Or(t *testing.T) {

	w1 := NewWhere("age", ">", "28")
	w2 := NewWhere("name", "==", "Tom")
	w3 := NewWhere("salary", ">", "5000")
	w4 := NewWhere("city", "==", "New York")

	or := NewCombinerOR(w1, w2, w3, w4)

	result := or.Or()

	expected := []*Where{w1, w2, w3, w4}

	require.Equal(t, len(expected), len(result), "The length of the result slice should match the expected length")

	for i := range result {
		require.Equal(t, expected[i].ToString(), result[i].ToString(), "The string representation of the Where clause should match the expected value")
	}
}

//func TestCombiner_QueryMod(t *testing.T) {
//	w1 := NewWhere("age", ">", "28")
//	w2 := NewWhere("name", "==", "Tom")
//	w3 := NewWhere("salary", ">", "5000")
//	w4 := NewWhere("city", "==", "New York")
//
//	combiner := NewCombinerOR(w1, w2, w3, w4)
//
//	expected := []qm.QueryMod{
//		qm.Expr(
//			qm.Where("age > ?", "28"),
//			qm.Or2(qm.Where("name == ?", "Tom")),
//			qm.Or2(qm.Where("salary > ?", "5000")),
//			qm.Or2(qm.Where("city == ?", "New York")),
//		),
//	}
//
//	result := combiner.QueryMod()
//
//	t.Logf("Generated QueryMod: %v", result)
//
//	require.Equal(t, expected, result, "The QueryMod result should match the expected QueryMod slice")
//}

func TestOr_ToString(t *testing.T) {
	w1 := NewWhere("age", ">", "28")
	w2 := NewWhere("name", "==", "Tom")
	w3 := NewWhere("salary", ">", "5000")
	w4 := NewWhere("city", "==", "New York")

	or := NewCombinerOR(w1, w2, w3, w4)

	expected := "(age > 28) Or (name == Tom) Or (salary > 5000) Or (city == New York)"

	result := or.ToString()

	require.Equal(t, expected, result, "The ToString result should match the expected string")
}
