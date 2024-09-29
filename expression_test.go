package sqlinjector

import (
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal/expression"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestODataExpression(t *testing.T) {
	q := url.Values(map[string][]string{
		defaultQueryNameLimit:  {"10"},
		defaultQueryNameOffset: {"5"},
		defaultQueryNameWhere:  {"field1 eq 'value1' and parent.field2 ne 'value2' and field3 gt 10"},
		defaultQueryNameOrder:  {"field1:asc, parent.field3, field2 desc, field4"},
	})

	e, err := ODataExpression(&q)
	require.NoError(t, err)
	require.NotNil(t, e)

	c, ok := e.(*combiner)
	require.True(t, ok)
	require.NotNil(t, c)

	require.Len(t, c.expressions, 9)

	t.Run("Limit", func(t *testing.T) {
		for _, expr := range c.expressions[0:1] {
			e, ok := expr.(*expression.Limit)
			require.True(t, ok)
			require.NotNil(t, e)
			require.Equal(t, "Limit 10", e.ToString())
		}
	})
	t.Run("Offset", func(t *testing.T) {
		for _, expr := range c.expressions[1:2] {
			e, ok := expr.(*expression.Offset)
			require.True(t, ok)
			require.NotNil(t, e)
			require.Equal(t, "Offset 5", e.ToString())
		}
	})
	t.Run("Where", func(t *testing.T) {
		for i, expr := range c.expressions[2:5] {
			e, ok := expr.(*expression.Where)
			require.True(t, ok)
			require.NotNil(t, e)
			expected := ""
			switch i {
			case 0:
				expected = "field1 eq value1"
			case 1:
				expected = "parents.field2 ne value2"
			case 2:
				expected = "field3 gt 10"
			}
			require.Equal(t, expected, e.ToString(), fmt.Sprintf("index: %d", i))
		}
	})
	t.Run("OrderBy", func(t *testing.T) {
		for i, expr := range c.expressions[5:9] {
			e, ok := expr.(*expression.OrderBy)
			require.True(t, ok)
			require.NotNil(t, e)
			expected := ""
			switch i {
			case 0:
				expected = "field1 ASC"
			case 1:
				expected = "parents.field3"
			case 2:
				expected = "field2 DESC"
			case 3:
				expected = "field4"
			}
			require.Equalf(t, expected, e.ToString(), fmt.Sprintf("index: %d", i))
		}
	})
}
