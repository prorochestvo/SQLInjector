package expression

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"reflect"
	"testing"
)

func TestExtractTableNameAndColumn(t *testing.T) {
	tests := []struct {
		column         string
		expectedTable  string
		expectedColumn string
	}{
		{"users/name", "users", "name"},
		{"name", "", "name"},
		{"orders/id", "orders", "id"},
		{"", "", ""},
	}

	for _, tt := range tests {
		table, column := extractTableNameAndColumn(tt.column)
		if table != tt.expectedTable || column != tt.expectedColumn {
			t.Errorf("extractTableNameAndColumn(%v) = %v, %v; expected %v, %v", tt.column, table, column, tt.expectedTable, tt.expectedColumn)
		}
	}
}

func TestJoinTableNameAndColumn(t *testing.T) {
	var mods []qm.QueryMod

	column := joinTableNameAndColumn("User", "name", &mods)
	expectedColumn := "\"users\".\"name\""
	expectedJoin := qm.InnerJoin("\"users\" ON \"users\".\"id\" = \"user_id\"")

	if column != expectedColumn {
		t.Errorf("joinTableNameAndColumn(\"User\", \"name\") = %v; expected %v", column, expectedColumn)
	}
	if len(mods) != 1 || !reflect.DeepEqual(mods[0], expectedJoin) {
		t.Errorf("Expected join mod %v, got %v", expectedJoin, mods)
	}

	column = joinTableNameAndColumn("", "name", nil)
	expectedColumn = "\"name\""

	if column != expectedColumn {
		t.Errorf("joinTableNameAndColumn(\"\", \"name\") = %v; expected %v", column, expectedColumn)
	}
}

func TestToPluralize(t *testing.T) {
	tests := []struct {
		word     string
		expected string
	}{
		{"box", "boxes"},
		{"city", "cities"},
		{"leaf", "leaves"},
		{"man", "men"},
		{"woman", "women"},
		{"dog", "dogs"},
	}

	for _, tt := range tests {
		plural := toPluralize(tt.word)
		if plural != tt.expected {
			t.Errorf("toPluralize(%v) = %v; expected %v", tt.word, plural, tt.expected)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserName", "user_name"},
		{"FirstName", "first_name"},
		{"userID", "user_i_d"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		snake := toSnakeCase(tt.input)
		if snake != tt.expected {
			t.Errorf("toSnakeCase(%v) = %v; expected %v", tt.input, snake, tt.expected)
		}
	}
}

type expression interface {
	QueryMod() []qm.QueryMod
	ToString() string
}
