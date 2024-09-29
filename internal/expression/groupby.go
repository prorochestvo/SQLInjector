package expression

import "github.com/volatiletech/sqlboiler/v4/queries/qm"

func NewGroupBy(column string) *GroupBy {
	t, c := extractTableNameAndColumn(column)
	return NewGroupWithTable(t, c)
}

func NewGroupWithTable(table, column string) *GroupBy {
	return &GroupBy{Table: table, Column: column}
}

type GroupBy struct {
	Table  string
	Column string
}

func (g *GroupBy) GroupBy() string {
	return g.Column
}

func (g *GroupBy) QueryMod() []qm.QueryMod {
	return []qm.QueryMod{qm.GroupBy(g.Column)}
}

func (g *GroupBy) ToString() string {
	return g.Column
}
