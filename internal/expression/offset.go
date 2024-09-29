package expression

import (
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strconv"
)

func NewOffsetFrom(expr string) ([]*Offset, error) {
	o := make([]*Offset, 0, 1)
	if expr != "" {
		n, err := strconv.ParseInt(expr, 10, 64)
		if err != nil || n < 0 {
			if err == nil {
				err = fmt.Errorf("incorrect value: %d", n)
			}
			return nil, err
		}
		o = append(o, &Offset{offset: int(n)})
	}
	return o, nil
}

func NewOffset(value int) *Offset {
	return &Offset{offset: max(value, 0)}
}

type Offset struct {
	offset int
}

func (o *Offset) Offset() int {
	return o.offset
}

func (o *Offset) QueryMod() []qm.QueryMod {
	return []qm.QueryMod{qm.Offset(o.offset)}
}

func (o *Offset) ToString() string {
	return fmt.Sprintf("Offset %d", o.offset)
}
