package expression

import (
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strconv"
)

func NewLimitFrom(expr string) ([]*Limit, error) {
	l := make([]*Limit, 0, 1)
	if expr != "" {
		n, err := strconv.ParseInt(expr, 10, 64)
		if err != nil || n < 0 {
			if err == nil {
				err = fmt.Errorf("incorrect value: %d", n)
			}
			return nil, err
		}
		l = append(l, &Limit{limit: int(n)})
	}
	return l, nil
}

func NewLimit(value int) *Limit {
	return &Limit{limit: max(value, 0)}
}

type Limit struct {
	limit int
}

func (l *Limit) Limit() int {
	return l.limit
}

func (l *Limit) QueryMod() []qm.QueryMod {
	return []qm.QueryMod{qm.Limit(l.limit)}
}

func (l *Limit) ToString() string {
	return fmt.Sprintf("Limit %d", l.limit)
}
