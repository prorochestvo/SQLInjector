package expression

import (
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gorm.io/gorm"
	"testing"
)

var _ expression = &Relation{}

func TestNewRelationFrom(t *testing.T) {
	table := "users"
	tables := []string{"orders", "payments"}

	expected := &Relation{
		tables: []string{"users", "orders", "payments"},
	}

	result := NewRelation(table, tables...)

	require.Equal(t, expected.ToString(), result.ToString(), "The relation result should match the expected relation")

}

func TestRelation(t *testing.T) {
	relation := &Relation{
		tables: []string{"users", "orders", "payments"},
	}

	expected := []string{"users", "orders", "payments"}

	result := relation.Relation()

	require.ElementsMatch(t, expected, result, "The relation result should match the expected tables")

}

func TestQueryMod(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to the in-memory database")

	_ = db

	relation := &Relation{
		tables: []string{"users", "orders"},
	}

	expected := []qm.QueryMod{
		qm.Load(qm.Rels("users", "orders")),
	}

	result := relation.QueryMod()

	require.Equal(t, expected, result, "The QueryMod result should match the expected QueryMod slice")
}

func TestToString(t *testing.T) {
	relation := &Relation{
		tables: []string{"users", "orders"},
	}

	expected := "JOIN users, orders"

	result := relation.ToString()

	require.Equal(t, expected, result, "The string representation of the relation should match the expected format")
}
