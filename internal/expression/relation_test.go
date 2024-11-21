package expression

import (
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gorm.io/gorm"
	"testing"
)

var _ expression = &Relation{}

func TestRelation_NewRelation(t *testing.T) {
	table := "users"
	tables := []string{"orders", "payments"}

	expected := &Relation{
		tables: []string{"users", "orders", "payments"},
	}

	result := NewRelation(table, tables...)

	require.Equal(t, expected.ToString(), result.ToString())

}

func TestRelation_Relation(t *testing.T) {
	relation := &Relation{
		tables: []string{"users", "orders", "payments"},
	}

	expected := []string{"users", "orders", "payments"}

	result := relation.Relation()

	require.ElementsMatch(t, expected, result)

}

func TestRelation_QueryMod(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	_ = db

	relation := &Relation{
		tables: []string{"users", "orders"},
	}

	expected := []qm.QueryMod{
		qm.Load(qm.Rels("users", "orders")),
	}

	result := relation.QueryMod()

	require.Equal(t, expected, result)
}

func TestRelation_ToStringToString(t *testing.T) {
	relation := &Relation{
		tables: []string{"users", "orders"},
	}

	expected := "JOIN users, orders"

	result := relation.ToString()

	require.Equal(t, expected, result)
}
