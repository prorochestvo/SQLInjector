package expression

import (
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"testing"
)

func TestTable_ExtractTableNameAndColumn(t *testing.T) {
	tests := []struct {
		name           string
		column         string
		expectedTable  string
		expectedColumn string
	}{
		{"With table and column", "users/name", "users", "name"},
		{"Without table", "name", "", "name"},
		{"With another table and column", "orders/id", "orders", "id"},
		{"Empty input", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, column := extractTableNameAndColumn(tt.column)

			require.Equal(t, tt.expectedTable, table)
			require.Equal(t, tt.expectedColumn, column)
		})
	}
}

func TestTable_JoinTableNameAndColumn(t *testing.T) {
	var mods []qm.QueryMod

	column := joinTableNameAndColumn("User", "name", &mods)
	expectedColumn := "\"users\".\"name\""
	expectedJoin := qm.InnerJoin("\"users\" ON \"users\".\"id\" = \"user_id\"")

	require.Equal(t, expectedColumn, column)
	require.Len(t, mods, 1)
	require.Equal(t, expectedJoin, mods[0])

	column = joinTableNameAndColumn("", "name", nil)
	expectedColumn = "\"name\""

	require.Equal(t, expectedColumn, column)
}

func TestTable_ToPluralize(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		expected string
	}{
		{"Pluralize regular word", "box", "boxes"},
		{"Pluralize word ending with 'y'", "city", "cities"},
		{"Pluralize irregular noun (leaf)", "leaf", "leaves"},
		{"Pluralize irregular noun (man)", "man", "men"},
		{"Pluralize irregular noun (woman)", "woman", "women"},
		{"Pluralize regular noun", "dog", "dogs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plural := toPluralize(tt.word)
			require.Equal(t, tt.expected, plural)
		})
	}
}

func TestTable_ToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Convert UserName to snake_case", "UserName", "user_name"},
		{"Convert FirstName to snake_case", "FirstName", "first_name"},
		{"Convert userID to snake_case", "userID", "user_i_d"},
		{"Leave simple unchanged", "simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snake := toSnakeCase(tt.input)
			require.Equal(t, tt.expected, snake)
		})
	}
}

type expression interface {
	QueryMod() []qm.QueryMod
	ToString() string
}
