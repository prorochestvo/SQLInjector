package expression

import (
	"errors"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"regexp"
	"strings"
)

// NewWhereFrom parses an Expression filter string into a slice of expression structs
func NewWhereFrom(filter string) ([]*Where, error) {
	pattern := regexp.MustCompile(`(\S+)\s*(eq|ne|gt|ge|lt|le|in|contains|startswith|endswith)\s*('[^']*'|[0-9]+)`)
	matches := pattern.FindAllStringSubmatch(filter, -1)

	if matches == nil {
		return nil, errors.New("invalid filter format")
	}

	var expr []*Where
	for _, match := range matches {
		f := match[1]
		o := Operator(match[2])
		v := match[3]

		switch o {
		case Equal:
		case NotEqual:
		case GreaterThan:
		case GreaterThanOrEqual:
		case LessThan:
		case LessThanOrEqual:
		//case In:
		//case NotIn:
		case Contains:
		case StartsWith:
		case EndsWith:
		//case IsNull:
		//case IsNotNull:
		default:
			return nil, fmt.Errorf("unsupported Operator: %s", match[2])
		}

		t := ""
		if strings.Contains(f, ".") {
			s := strings.SplitN(f, ".", 2)
			t = s[0]
			f = s[1]
		}

		var value interface{}
		if o == IsNull || o == IsNotNull {
			value = nil
		} else if o == In || o == NotIn {
			v = strings.TrimPrefix(v, "(")
			v = strings.TrimSuffix(v, ")")
			v = strings.TrimSpace(v)
			if strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'") {
				splitter := regexp.MustCompile(`^'\s*,\s*'$`)
				parts := splitter.Split(v[1:len(v)-1], -1)
				slice := make([]interface{}, len(parts))
				for i, p := range parts {
					slice[i] = p
				}
				value = slice
				parts = nil
				splitter = nil
			} else {
				parts := strings.Split(v, ",")
				slice := make([]interface{}, len(parts))
				for i, p := range parts {
					slice[i] = strings.TrimSpace(p)
				}
				value = slice
				parts = nil
			}
		} else {
			value = strings.Trim(v, "'")
		}

		expr = append(expr, &Where{
			Table:    t,
			Column:   f,
			Operator: o,
			Value:    value,
		})
	}

	return expr, nil
}

func NewWhere(column string, operator Operator, values ...interface{}) *Where {
	t, c := extractTableNameAndColumn(column)
	return NewWhereWithTable(
		t,
		c,
		operator,
		values...,
	)
}

func NewWhereWithTable(table string, column string, operator Operator, values ...interface{}) *Where {
	var v interface{} = nil
	if operator == In || operator == NotIn {
		v = values
	} else if l := len(values); l > 1 {
		v = values
	} else if l == 1 {
		v = values[0]
	}
	return &Where{
		Table:    table,
		Column:   column,
		Operator: operator,
		Value:    v,
	}
}

type Where struct {
	Table    string
	Column   string
	Operator Operator
	Value    interface{}
}

func (w *Where) Where() string {
	return joinTableNameAndColumn(w.Table, w.Column, nil)
}

func (w *Where) QueryMod() []qm.QueryMod {
	mods := make([]qm.QueryMod, 0)

	c := joinTableNameAndColumn(w.Table, w.Column, &mods)

	var m qm.QueryMod

	switch w.Operator {
	case Equal:
		m = qm.Where(c+" = ?", w.Value)
	case NotEqual:
		m = qm.Where(c+" <> ?", w.Value)
	case GreaterThan:
		m = qm.Where(c+" > ?", w.Value)
	case GreaterThanOrEqual:
		m = qm.Where(c+" >= ?", w.Value)
	case LessThan:
		m = qm.Where(c+" < ?", w.Value)
	case LessThanOrEqual:
		m = qm.Where(c+" <= ?", w.Value)
	case In:
		if vInterface, ok := w.Value.([]interface{}); ok {
			m = qm.WhereIn(c+" IN ?", vInterface...)
		} else {
			m = qm.WhereIn(c+" IN ?", w.Value)
		}
	case NotIn:
		if vInterface, ok := w.Value.([]interface{}); ok {
			m = qm.WhereNotIn(c+" IN ?", vInterface...)
		} else {
			m = qm.WhereNotIn(c+" IN ?", w.Value)
		}
	case Contains:
		m = qm.Where(c+" LIKE ?", fmt.Sprintf("%%%v%%", w.Value))
	case StartsWith:
		m = qm.Where(c+" LIKE ?", fmt.Sprintf("%%%v", w.Value))
	case EndsWith:
		m = qm.Where(c+" LIKE ?", fmt.Sprintf("%v%%", w.Value))
	case IsNull:
		m = qm.Where(c + " IS NULL")
	case IsNotNull:
		m = qm.Where(c + " IS NOT NULL")
	default:
		return nil
	}

	mods = append(mods, m)

	return mods
}

func (w *Where) ToString() string {
	f := w.Column
	if w.Table != "" {
		t := toSnakeCase(w.Table)
		f = toPluralize(t) + "." + f
	}
	return fmt.Sprintf("%s %s %v", f, w.Operator, w.Value)
}

type Operator string

const (
	Equal              Operator = "eq"
	NotEqual           Operator = "ne"
	GreaterThan        Operator = "gt"
	GreaterThanOrEqual Operator = "ge"
	LessThan           Operator = "lt"
	LessThanOrEqual    Operator = "le"
	In                 Operator = "in"
	NotIn              Operator = "notIn"
	IsNull             Operator = "isNull"
	IsNotNull          Operator = "isNotNull"
	Contains           Operator = "contains"
	StartsWith         Operator = "startswith"
	EndsWith           Operator = "endswith"
)
