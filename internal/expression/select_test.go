package expression

import (
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gorm.io/gorm"
	"testing"
)

var _ expression = &Select{}

func TestNewSelect(t *testing.T) {
	s := NewSelect("id")
	expectedColumns := []string{"id"}

	require.Equal(t, expectedColumns, s.Select(), "The selected columns should match the expected columns")

	s = NewSelect("id", "name", "age")
	expectedColumns = []string{"id", "name", "age"}

	require.Equal(t, expectedColumns, s.Select(), "The selected columns should match the expected columns")
}

func TestSelect(t *testing.T) {
	s := NewSelect("id")
	expected := []string{"id"}

	require.Equal(t, expected, s.Select(), "The selected columns should match the expected columns")

	s = NewSelect("id", "name", "age")
	expected = []string{"id", "name", "age"}

	require.Equal(t, expected, s.Select(), "The selected columns should match the expected columns")
}

func TestSelect_QueryMod(t *testing.T) {

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to the in-memory database")

	_ = db

	s := NewSelect("id")
	expectedQueryMod := []qm.QueryMod{qm.Select("id")}

	require.Equal(t, expectedQueryMod, s.QueryMod(), "Expected and actual QueryMod should be equal for single column select")

	s = NewSelect("id", "name", "age")
	expectedQueryMod = []qm.QueryMod{qm.Select("id", "name", "age")}

	require.Equal(t, expectedQueryMod, s.QueryMod(), "Expected and actual QueryMod should be equal for multiple columns select")
}

func TestSelect_ToString(t *testing.T) {
	s := NewSelect("id")
	expectedString := "Select id"

	require.Equal(t, expectedString, s.ToString(), "Expected and actual Select string should be equal for single column select")

	s = NewSelect("id", "name", "age")
	expectedString = "Select id, name, age"

	require.Equal(t, expectedString, s.ToString(), "Expected and actual Select string should be equal for multiple columns select")
}
