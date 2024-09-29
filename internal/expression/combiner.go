package expression

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strings"
)

func NewCombinerOR(l *Where, r *Where, e ...*Where) *Or {
	items := make([]*Where, 0, len(e)+2)
	items = append(items, l, r)
	items = append(items, e...)
	return &Or{items: items}
}

type Or struct {
	items []*Where
}

func (o *Or) Or() []*Where {
	return o.items
}

func (o *Or) QueryMod() []qm.QueryMod {
	mods := make([]qm.QueryMod, 0, len(o.items))
	orMods := make([]qm.QueryMod, 0, len(o.items))
	for i, e := range o.items {
		mod := e.QueryMod()
		l := len(mod)
		if l == 0 {
			continue
		}
		l--
		for m := 0; m < l; m++ {
			mods = append(mods, mod[m])
		}
		if i == 0 {
			orMods = append(orMods, mod[l])
		} else {
			orMods = append(orMods, qm.Or2(mod[l]))
		}
	}
	mods = append(mods, qm.Expr(orMods...))
	return mods
}

func (o *Or) ToString() string {
	w := make([]string, len(o.items))
	for i, item := range o.items {
		w[i] = item.ToString()
	}
	s := strings.Join(w, ") Or (")
	if len(w) > 1 {
		s = "(" + s + ")"
	}
	return s
}
