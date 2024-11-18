package expression

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"testing"
)

var _ expression = &GroupBy{}

func TestGroupBy_NewGroupBy(t *testing.T) {
	testCases := []struct {
		inputColumn    string
		expectedTable  string
		expectedColumn string
	}{
		{"users/name", "users", "name"},
		{"name", "", "name"},
		{"orders/id", "orders", "id"},
		{"", "", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.inputColumn, func(t *testing.T) {
			gb := NewGroupBy(testCase.inputColumn)
			require.Equal(t, testCase.expectedTable, gb.Table)
			require.Equal(t, testCase.expectedColumn, gb.Column)
		})
	}
}

func TestGroupBy_NewGroupWithTable(t *testing.T) {
	testCases := []struct {
		inputTable     string
		inputColumn    string
		expectedTable  string
		expectedColumn string
	}{
		{"users", "name", "users", "name"},
		{"orders", "id", "orders", "id"},
		{"", "age", "", "age"},
		{"products", "", "products", ""},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s/%s", testCase.inputTable, testCase.inputColumn), func(t *testing.T) {
			gb := NewGroupWithTable(testCase.inputTable, testCase.inputColumn)
			require.Equal(t, testCase.expectedTable, gb.Table)
			require.Equal(t, testCase.expectedColumn, gb.Column)
		})
	}
}
func TestGroupBy_GroupBy(t *testing.T) {
	testCases := []struct {
		inputColumn   string
		expectedValue string
	}{
		{"user/name", "name"},
		{"age", "age"},
		{"salary", "salary"},
		{"", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.inputColumn, func(t *testing.T) {
			gb := NewGroupBy(testCase.inputColumn)
			actually := gb.GroupBy()
			require.Equal(t, testCase.expectedValue, actually)
		})
	}
}

func TestGroupBy_QueryMod(t *testing.T) {
	testCases := []struct {
		inputColumn   string
		expectedValue []qm.QueryMod
	}{
		{"user/name", []qm.QueryMod{qm.GroupBy("name")}},
		{"age", []qm.QueryMod{qm.GroupBy("age")}},
		{"salary", []qm.QueryMod{qm.GroupBy("salary")}},
		{"", []qm.QueryMod{qm.GroupBy("")}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.inputColumn, func(t *testing.T) {
			gb := NewGroupBy(testCase.inputColumn)
			actually := gb.QueryMod()
			require.Equal(t, testCase.expectedValue, actually)
		})
	}
}

func TestGroupBy_ToString(t *testing.T) {
	testCases := []struct {
		inputColumn   string
		expectedValue string
	}{
		{"user/name", "name"},
		{"age", "age"},
		{"salary", "salary"},
		{"", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.inputColumn, func(t *testing.T) {
			gb := NewGroupBy(testCase.inputColumn)
			actually := gb.ToString()
			require.Equal(t, testCase.expectedValue, actually)
		})
	}
}
