package expression

import (
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gorm.io/gorm"
	"testing"
)

var _ expression = &Select{}

func TestSelect_NewSelect(t *testing.T) {
	s := NewSelect("id")
	expectedColumns := []string{"id"}

	require.Equal(t, expectedColumns, s.Select())

	s = NewSelect("id", "name", "age")
	expectedColumns = []string{"id", "name", "age"}

	require.Equal(t, expectedColumns, s.Select())
}

func TestSelect_Select(t *testing.T) {
	s := NewSelect("id")
	expected := []string{"id"}

	require.Equal(t, expected, s.Select())

	s = NewSelect("id", "name", "age")
	expected = []string{"id", "name", "age"}

	require.Equal(t, expected, s.Select())
}

func TestSelect_QueryMod(t *testing.T) {

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	_ = db

	s := NewSelect("id")
	expectedQueryMod := []qm.QueryMod{qm.Select("id")}

	require.Equal(t, expectedQueryMod, s.QueryMod())

	s = NewSelect("id", "name", "age")
	expectedQueryMod = []qm.QueryMod{qm.Select("id", "name", "age")}

	require.Equal(t, expectedQueryMod, s.QueryMod())
}

func TestSelect_ToString(t *testing.T) {
	s := NewSelect("id")
	expectedString := "Select id"

	require.Equal(t, expectedString, s.ToString())

	s = NewSelect("id", "name", "age")
	expectedString = "Select id, name, age"

	require.Equal(t, expectedString, s.ToString())
}
