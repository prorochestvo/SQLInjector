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

	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
