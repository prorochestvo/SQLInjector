package expression

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"testing"
)

func TestExtractTableNameAndColumn(t *testing.T) {
	t.Skip("not implemented")
}

func TestJoinTableNameAndColumn(t *testing.T) {
	t.Skip("not implemented")
}

func TestToPluralize(t *testing.T) {
	t.Skip("not implemented")
}

func TestToSnakeCase(t *testing.T) {
	t.Skip("not implemented")
}

type expression interface {
	QueryMod() []qm.QueryMod
	ToString() string
}
