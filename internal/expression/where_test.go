package expression

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var _ expression = &Where{}

func TestNewWhereFrom(t *testing.T) {
	w, err := NewWhereFrom("p.f ne 1 and f gt '22 (N1)' aNd pZaS.f ge '3' and f eq 4 AND f contains '55 (N2)'")
	require.NoError(t, err)
	require.NotNil(t, w)
	require.Len(t, w, 5)
	require.Equal(t, "ps.f ne 1", w[0].ToString())
	require.Equal(t, "f gt 22 (N1)", w[1].ToString())
	require.Equal(t, "p_za_ses.f ge 3", w[2].ToString())
	require.Equal(t, "f eq 4", w[3].ToString())
	require.Equal(t, "f contains 55 (N2)", w[4].ToString())
}

func TestNewWhere(t *testing.T) {
	t.Skip("not implemented")
}
