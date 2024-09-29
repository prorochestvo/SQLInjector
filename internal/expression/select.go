package expression

import (
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strings"
)

func NewSelect(column string, columns ...string) *Select {
	c := make([]string, 0, len(columns)+1)
	c = append(c, column)
	c = append(c, columns...)
	return &Select{columns: c}
}

type Select struct {
	columns []string
}

func (s *Select) Select() []string {
	return s.columns
}

func (s *Select) QueryMod() []qm.QueryMod {
	return []qm.QueryMod{qm.Select(s.columns...)}
}

func (s *Select) ToString() string {
	return fmt.Sprintf("Select %s", strings.Join(s.columns, ", "))
}
