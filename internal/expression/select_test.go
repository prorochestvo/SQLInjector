package expression

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"reflect"
	"testing"
)

var _ expression = &Select{}

func TestNewSelect(t *testing.T) {
	s := NewSelect("id")
	expectedColumns := []string{"id"}
	if !reflect.DeepEqual(s.Select(), expectedColumns) {
		t.Errorf("Expected %v, got %v", expectedColumns, s.Select())
	}

	s = NewSelect("id", "name", "age")
	expectedColumns = []string{"id", "name", "age"}
	if !reflect.DeepEqual(s.Select(), expectedColumns) {
		t.Errorf("Expected %v, got %v", expectedColumns, s.Select())
	}
}

func TestSelect(t *testing.T) {
	s := NewSelect("id")
	expected := []string{"id"}
	if !reflect.DeepEqual(s.Select(), expected) {
		t.Errorf("Expected %v, got %v", expected, s.Select())
	}

	s = NewSelect("id", "name", "age")
	expected = []string{"id", "name", "age"}
	if !reflect.DeepEqual(s.Select(), expected) {
		t.Errorf("Expected %v, got %v", expected, s.Select())
	}
}

func TestSelect_QueryMod(t *testing.T) {
	s := NewSelect("id")
	expectedQueryMod := []qm.QueryMod{qm.Select("id")}
	if !reflect.DeepEqual(s.QueryMod(), expectedQueryMod) {
		t.Errorf("Expected %v, got %v", expectedQueryMod, s.QueryMod())
	}

	s = NewSelect("id", "name", "age")
	expectedQueryMod = []qm.QueryMod{qm.Select("id", "name", "age")}
	if !reflect.DeepEqual(s.QueryMod(), expectedQueryMod) {
		t.Errorf("Expected %v, got %v", expectedQueryMod, s.QueryMod())
	}
}

func TestSelect_ToString(t *testing.T) {
	s := NewSelect("id")
	expectedString := "Select id"
	if s.ToString() != expectedString {
		t.Errorf("Expected %v, got %v", expectedString, s.ToString())
	}

	s = NewSelect("id", "name", "age")
	expectedString = "Select id, name, age"
	if s.ToString() != expectedString {
		t.Errorf("Expected %v, got %v", expectedString, s.ToString())
	}
}
