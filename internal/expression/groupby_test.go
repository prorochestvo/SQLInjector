package expression

import (
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"testing"
)

var _ expression = &GroupBy{}

func TestNewGroupBy(t *testing.T) {
	tests := []struct {
		input          string
		expectedTable  string
		expectedColumn string
	}{
		{"users/name", "users", "name"},
		{"name", "", "name"},
		{"orders/id", "orders", "id"},
		{"", "", ""},
	}

	for _, tt := range tests {
		groupBy := NewGroupBy(tt.input)

		require.Equal(t, tt.expectedTable, groupBy.Table, "Expected table to be %v, got %v", tt.expectedTable, groupBy.Table)
		require.Equal(t, tt.expectedColumn, groupBy.Column, "Expected column to be %v, got %v", tt.expectedColumn, groupBy.Column)
	}
}

func TestNewGroupWithTable(t *testing.T) {
	tests := []struct {
		table          string
		column         string
		expectedTable  string
		expectedColumn string
	}{
		{"users", "name", "users", "name"},
		{"orders", "id", "orders", "id"},
		{"", "age", "", "age"},
		{"products", "", "products", ""},
	}

	for _, tt := range tests {
		groupBy := NewGroupWithTable(tt.table, tt.column)

		require.Equal(t, tt.expectedTable, groupBy.Table, "Expected table to be %v, got %v", tt.expectedTable, groupBy.Table)
		require.Equal(t, tt.expectedColumn, groupBy.Column, "Expected column to be %v, got %v", tt.expectedColumn, groupBy.Column)
	}
}
func TestGroupBy_GroupBy(t *testing.T) {
	tests := []struct {
		column         string
		expectedOutput string
	}{
		{"name", "name"},
		{"age", "age"},
		{"salary", "salary"},
		{"", ""},
	}

	for _, tt := range tests {
		groupBy := NewGroupWithTable("users", tt.column)

		output := groupBy.GroupBy()
		require.Equal(t, tt.expectedOutput, output, "Expected output to be %v, got %v", tt.expectedOutput, output)
	}
}

func TestGROUPBY_QueryMod(t *testing.T) {
	tests := []struct {
		column         string
		expectedOutput []qm.QueryMod
	}{
		{"name", []qm.QueryMod{qm.GroupBy("name")}},
		{"age", []qm.QueryMod{qm.GroupBy("age")}},
		{"salary", []qm.QueryMod{qm.GroupBy("salary")}},
		{"", []qm.QueryMod{qm.GroupBy("")}},
	}

	for _, tt := range tests {
		groupBy := NewGroupWithTable("users", tt.column)

		output := groupBy.QueryMod()
		require.Equal(t, tt.expectedOutput, output, "Expected output to be %v, got %v", tt.expectedOutput, output)
	}
}

func TestGROUPBY_ToString(t *testing.T) {
	tests := []struct {
		column         string
		expectedOutput string
	}{
		{"name", "name"},
		{"age", "age"},
		{"salary", "salary"},
		{"", ""},
	}

	for _, tt := range tests {
		groupBy := NewGroupWithTable("users", tt.column)

		output := groupBy.ToString()
		require.Equal(t, tt.expectedOutput, output, "Expected output to be %v, got %v", tt.expectedOutput, output)
	}
}
