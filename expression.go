package main

import (
	"github.com/prorochestvo/sqlinjector/internal/expression"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"net/url"
	"strings"
)

type Expression interface {
	QueryMod() []qm.QueryMod
}

// ODataExpression parses Expression options from the request and returns a slice of query mods for SQLBoiler.
// The following options are supported:
// - $select - a comma-separated list of Columns to select
// - $limit - the maximum number of records to return
// - $offset - the number of records to skip
// - $filter - a filter expression in the form of "Column operator value" Where operator is one of eq, ne, gt, ge, lt, le, in, contains, startswith, endswith and value is a string or a number (without quotes and not supported Or operator).
// Example: $filter=field1 eq 'value1' and field2 ne 'value2' or field3 gt 10
// The function returns an error if the filter expression is invalid or an unsupported operator is used.
func ODataExpression(q *url.Values, defaultPagination ...int) (Expression, error) {
	var expr []Expression

	qLimit := strings.TrimSpace(q.Get(defaultQueryNameLimit))
	qOffset := strings.TrimSpace(q.Get(defaultQueryNameOffset))
	qFilter := strings.TrimSpace(q.Get(defaultQueryNameWhere))
	qSort := strings.TrimSpace(q.Get(defaultQueryNameOrder))

	lDefaultPagination := len(defaultPagination)

	if qLimit != "" {
		e, err := expression.NewLimitFrom(qLimit)
		if err != nil {
			return nil, err
		}
		for _, limit := range e {
			expr = append(expr, limit)
		}
	} else if lDefaultPagination > 0 {
		expr = append(expr, expression.NewLimit(defaultPagination[0]))
	}

	if qOffset != "" {
		e, err := expression.NewOffsetFrom(qOffset)
		if err != nil {
			return nil, err
		}
		for _, offset := range e {
			expr = append(expr, offset)
		}
	} else if lDefaultPagination > 1 {
		expr = append(expr, expression.NewLimit(defaultPagination[1]))
	}

	if qFilter != "" {
		e, err := expression.NewWhereFrom(qFilter)
		if err != nil {
			return nil, err
		}
		for _, where := range e {
			expr = append(expr, where)
		}
	}

	if qSort != "" {
		e, err := expression.NewOrderByFrom(qSort)
		if err != nil {
			return nil, err
		}
		for _, orderBy := range e {
			expr = append(expr, orderBy)
		}
	}

	return &combiner{expressions: expr}, nil
}

func Relation(t string, extra ...string) Expression {
	return expression.NewRelation(t, extra...)
}

func Where(c string, o operator, v ...interface{}) Expression {
	return expression.NewWhere(c, o, v...)
}

func Or(l Expression, r Expression, extra ...Expression) Expression {
	expr := make([]Expression, 0, len(extra)+2)

	lW, ok := l.(*expression.Where)
	if !ok {
		expr = append(expr, l)
		expr = append(expr, r)
		expr = append(expr, extra...)
		return &combiner{expressions: expr}
	}

	rW, ok := r.(*expression.Where)
	if !ok {
		expr = append(expr, l)
		expr = append(expr, r)
		expr = append(expr, extra...)
		return &combiner{expressions: expr}
	}

	extraW := make([]*expression.Where, 0, len(extra))
	for _, e := range extra {
		var eW *expression.Where
		if eW, ok = e.(*expression.Where); ok {
			extraW = append(extraW, eW)
		} else {
			expr = append(expr, e)
		}
	}

	e := expression.NewCombinerOR(lW, rW, extraW...)
	if len(expr) == 0 {
		return e
	}

	expr = append(expr, e)

	return &combiner{expressions: expr}
}

func GroupBy(c string, extra ...string) Expression {
	groupBy := expression.NewGroupBy(c)
	if len(extra) == 0 {
		return groupBy
	}

	expr := make([]Expression, 0, len(extra)+1)

	expr = append(expr, groupBy)

	for _, e := range extra {
		expr = append(expr, expression.NewGroupBy(e))
	}

	return &combiner{expressions: expr}
}

func OrderBy(c string, d direction, extra ...string) Expression {
	orderBy := expression.NewOrderBy(c, d)
	if len(extra) == 0 {
		return orderBy
	}

	expr := make([]Expression, 0, len(extra)+1)

	expr = append(expr, orderBy)

	for _, e := range extra {
		expr = append(expr, expression.NewOrderBy(e, d))
	}

	return &combiner{expressions: expr}
}

func Limit(v int) Expression {
	return expression.NewLimit(v)
}

func Offset(v int) Expression {
	return expression.NewOffset(v)
}

const (
	defaultQueryNameWhere  = "$filter"
	defaultQueryNameOrder  = "$sort"
	defaultQueryNameLimit  = "$Limit"
	defaultQueryNameOffset = "$Offset"
)

type operator = expression.Operator

const (
	Equal              operator = expression.Equal
	NotEqual           operator = expression.NotEqual
	GreaterThan        operator = expression.GreaterThan
	GreaterThanOrEqual operator = expression.GreaterThanOrEqual
	LessThan           operator = expression.LessThan
	LessThanOrEqual    operator = expression.LessThanOrEqual
	In                 operator = expression.In
	NotIn              operator = expression.NotIn
	IsNull             operator = expression.IsNull
	IsNotNull          operator = expression.IsNotNull
	Contains           operator = expression.Contains
	StartsWith         operator = expression.StartsWith
	EndsWith           operator = expression.EndsWith
)

type direction = expression.Direction

const (
	Ascending  direction = expression.Ascending
	Descending direction = expression.Descending
)

type combiner struct {
	expressions []Expression
}

func (c *combiner) QueryMod() []qm.QueryMod {
	mods := make([]qm.QueryMod, 0, len(c.expressions))
	for _, e := range c.expressions {
		mods = append(mods, e.QueryMod()...)
	}
	return mods
}
