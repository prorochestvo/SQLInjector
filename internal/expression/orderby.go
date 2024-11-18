package expression

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strings"
)

func NewOrderByFrom(expr string) ([]*OrderBy, error) {
	e := strings.Split(expr, ",")
	o := make([]*OrderBy, 0, len(e))
	for _, item := range e {
		item = strings.TrimSpace(item)
		item = strings.ReplaceAll(item, ":", " ")

		f := ""
		d := defaulting

		if strings.Contains(item, " ") {
			parts := strings.SplitN(item, " ", 2)
			if l := len(parts); l == 2 {
				f = parts[0]
				if w := parts[1]; w != "" {
					if w[0] == 'd' || w[0] == 'D' {
						d = Descending
					} else if w[0] == 'a' || w[0] == 'A' {
						d = Ascending
					}
				}
			} else if l == 1 {
				f = parts[0]
			} else if l != 1 && l != 2 {
				continue
			}
		} else {
			f = item
		}

		t := ""
		if strings.Contains(f, ".") {
			s := strings.SplitN(f, ".", 2)
			t = s[0]
			f = s[1]
		}

		o = append(o, &OrderBy{
			Table:     t,
			Column:    f,
			Direction: d,
		})
	}
	return o, nil
}

func NewOrderBy(column string, direction Direction) *OrderBy {
	t, c := extractTableNameAndColumn(column)
	return NewOrderByWithTable(t, c, direction)
}

func NewOrderByWithTable(table, column string, direction Direction) *OrderBy {
	return &OrderBy{Table: table, Column: column, Direction: direction}
}

type OrderBy struct {
	Table     string
	Column    string
	Direction Direction
}

func (o *OrderBy) OrderBy() string {
	return string(o.Direction)
}

func (o *OrderBy) QueryMod() []qm.QueryMod {
	mods := make([]qm.QueryMod, 0)

	c := joinTableNameAndColumn(o.Table, o.Column, &mods)

	var d string
	if o.Direction != defaulting {
		d = " " + string(o.Direction)
	}

	return []qm.QueryMod{qm.OrderBy(c + d)}
}

func (o *OrderBy) ToString() string {
	c := o.Column
	if o.Table != "" {
		t := toSnakeCase(o.Table)
		c = toPluralize(t) + "." + c
	}

	var d string
	if o.Direction != defaulting {
		d = " " + string(o.Direction)
	}

	return c + d
}

type Direction string

const (
	defaulting Direction = ""
	Ascending  Direction = "ASC"
	Descending Direction = "DESC"
)
