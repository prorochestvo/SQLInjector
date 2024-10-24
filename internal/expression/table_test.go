package expression

import (
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

		require.Equal(t, tt.expectedTable, table, "Expected and actual table names should be equal for column: %v", tt.column)
		require.Equal(t, tt.expectedColumn, column, "Expected and actual column names should be equal for column: %v", tt.column)
	}
}

func TestJoinTableNameAndColumn(t *testing.T) {
	var mods []qm.QueryMod

	column := joinTableNameAndColumn("User", "name", &mods)
	expectedColumn := "\"users\".\"name\""
	expectedJoin := qm.InnerJoin("\"users\" ON \"users\".\"id\" = \"user_id\"")

	require.Equal(t, expectedColumn, column, "joinTableNameAndColumn(\"User\", \"name\") должен вернуть ожидаемое значение")
	require.Len(t, mods, 1, "Массив mods должен содержать один элемент")
	require.Equal(t, expectedJoin, mods[0], "Ожидается правильный join мод")

	column = joinTableNameAndColumn("", "name", nil)
	expectedColumn = "\"name\""

	require.Equal(t, expectedColumn, column, "joinTableNameAndColumn(\"\", \"name\") должен вернуть ожидаемое значение")
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

		require.Equal(t, tt.expected, plural, "toPluralize(%v) должен вернуть %v", tt.word, tt.expected)
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
		require.Equal(t, tt.expected, snake, "toSnakeCase(%v) должен вернуть %v", tt.input, tt.expected)
	}
}

type expression interface {
	QueryMod() []qm.QueryMod
	ToString() string
}
