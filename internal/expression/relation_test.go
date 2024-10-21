package expression

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"reflect"
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

	// TODO: REVIEW: all expressions has ToString() method, so you can use it to compare expected and actual results.
	// TODO: REVIEW: or you can use require.Equal() method to compare expected and actual results. (require.Equal(t, expected.tables, result.tables))
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
func TestRelation(t *testing.T) {
	relation := &Relation{
		tables: []string{"users", "orders", "payments"},
	}

	expected := []string{"users", "orders", "payments"}

	result := relation.Relation()

	// TODO: REVIEW: all expressions has ToString() method, so you can use it to compare expected and actual results.
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestQueryMod(t *testing.T) {
	relation := &Relation{
		tables: []string{"users", "orders"},
	}

	expected := []qm.QueryMod{
		qm.Load(qm.Rels("users", "orders")),
	}

	result := relation.QueryMod()

	// TODO: REVIEW: you can use sqlite in-memory database for testing.
	// TODO: however we can make this task later, because it is tough to implement at this moment
	//db, err := receptacle.NewSQLite3()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestToString(t *testing.T) {
	relation := &Relation{
		tables: []string{"users", "orders"},
	}

	expected := "JOIN users, orders"

	result := relation.ToString()

	// TODO: REVIEW: could you use require.Equal() method to compare expected and actual results.
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
