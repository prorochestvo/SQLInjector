package expression

import (
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strings"
)

func NewRelation(table string, tables ...string) *Relation {
	t := make([]string, 0, len(tables)+1)
	t = append(t, table)
	t = append(t, tables...)
	return &Relation{tables: t}
}

type Relation struct {
	tables []string
}

func (r *Relation) Relation() []string {
	return r.tables
}

func (r *Relation) QueryMod() []qm.QueryMod {
	return []qm.QueryMod{qm.Load(qm.Rels(r.tables...))}
}

func (r *Relation) ToString() string {
	return fmt.Sprintf("JOIN %s", strings.Join(r.tables, ", "))
}
