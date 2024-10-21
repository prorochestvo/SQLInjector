package expression

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

var _ expression = &Or{}

func TestNewCombinerOR(t *testing.T) {

	w1 := NewWhere("age", ">", "28")
	w2 := NewWhere("name", "==", "Tom")
	w3 := NewWhere("salary", ">", "5000")
	w4 := NewWhere("city", "==", "New York")

	or1 := NewCombinerOR(w1, w2)
	require.Equal(t, 2, len(or1.items))

	expectedItems1 := []*Where{w1, w2}
	// TODO: REVIEW: reflect is very expensive, so could you create test cases without using it.
	// TODO: REVIEW: all expressions has ToString() method, so you can use it to compare expected and actual results.
	if !reflect.DeepEqual(or1.items, expectedItems1) {
		t.Errorf("Expected %v, got %v", expectedItems1, or1.items)
	}

	or2 := NewCombinerOR(w1, w2, w3, w4)
	require.Equal(t, 4, len(or2.items))

	expectedItems2 := []*Where{w1, w2, w3, w4}
	if !reflect.DeepEqual(or2.items, expectedItems2) {
		t.Errorf("Expected %v, got %v", expectedItems2, or2.items)
	}
}
func TestOR(t *testing.T) {

	w1 := NewWhere("age", ">", "28")
	w2 := NewWhere("name", "==", "Tom")
	w3 := NewWhere("salary", ">", "5000")
	w4 := NewWhere("city", "==", "New York")

	or := NewCombinerOR(w1, w2, w3, w4)

	result := or.Or()

	expected := []*Where{w1, w2, w3, w4}

	// TODO: REVIEW: reflect is very expensive, so could you create test cases without using it.
	// TODO: REVIEW: all expressions has ToString() method, so you can use it to compare expected and actual results.
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}
