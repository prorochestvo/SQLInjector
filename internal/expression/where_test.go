package expression

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

var _ expression = &Where{}

func TestNewWhereFrom(t *testing.T) {
	tests := []struct {
		filter   string
		expected []*Where
		hasError bool
	}{
		{
			filter: "User.name eq 'John'",
			expected: []*Where{
				{Table: "User", Column: "name", Operator: Equal, Value: "John"},
			},
			hasError: false,
		},
		{
			filter: "age gt 30",
			expected: []*Where{
				{Table: "", Column: "age", Operator: GreaterThan, Value: "30"},
			},
			hasError: false,
		},
		{
			filter: "User.age lt 25, User.name eq 'Alice'",
			expected: []*Where{
				{Table: "User", Column: "age", Operator: LessThan, Value: "25"},
				{Table: "User", Column: "name", Operator: Equal, Value: "Alice"},
			},
			hasError: false,
		},
		{
			filter:   "invalidFilter",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		result, err := NewWhereFrom(tt.filter)
		if tt.hasError {
			require.Error(t, err, "Expected an error for filter: %v", tt.filter)
		} else {
			require.NoError(t, err, "Did not expect an error for filter: %v", tt.filter)
			require.Equal(t, tt.expected, result, "Expected %v, got %v", tt.expected, result)
		}
	}
}

func TestNewWhere(t *testing.T) {
	tests := []struct {
		column   string
		operator Operator
		values   []interface{}
		expected *Where
	}{
		{
			column:   "User.name",
			operator: Equal,
			values:   []interface{}{"John"},
			expected: &Where{
				Table:    "", // Здесь ожидаемое значение пусто, так как "User.name" не разделено на таблицу и колонку
				Column:   "User.name",
				Operator: Equal,
				Value:    "John",
			},
		},
		{
			column:   "age",
			operator: GreaterThan,
			values:   []interface{}{30},
			expected: &Where{
				Table:    "",
				Column:   "age",
				Operator: GreaterThan,
				Value:    30,
			},
		},
	}

	for _, tt := range tests {
		result := NewWhere(tt.column, tt.operator, tt.values...)
		require.Equal(t, tt.expected, result, "Expected %v, got %v", tt.expected, result)
	}
}

func TestNewWhereWithTable(t *testing.T) {
	tests := []struct {
		table    string
		column   string
		operator Operator
		values   []interface{}
		expected *Where
	}{
		{
			table:    "User",
			column:   "name",
			operator: Equal,
			values:   []interface{}{"John"},
			expected: &Where{
				Table:    "User",
				Column:   "name",
				Operator: Equal,
				Value:    "John",
			},
		},
		{
			table:    "User",
			column:   "age",
			operator: GreaterThan,
			values:   []interface{}{30},
			expected: &Where{
				Table:    "User",
				Column:   "age",
				Operator: GreaterThan,
				Value:    30,
			},
		},
		{
			table:    "Orders",
			column:   "id",
			operator: In,
			values:   []interface{}{1, 2, 3},
			expected: &Where{
				Table:    "Orders",
				Column:   "id",
				Operator: In,
				Value:    []interface{}{1, 2, 3},
			},
		},
		{
			table:    "Product",
			column:   "price",
			operator: NotIn,
			values:   []interface{}{100, 200},
			expected: &Where{
				Table:    "Product",
				Column:   "price",
				Operator: NotIn,
				Value:    []interface{}{100, 200},
			},
		},
		{
			table:    "",
			column:   "status",
			operator: NotEqual,
			values:   []interface{}{"active"},
			expected: &Where{
				Table:    "",
				Column:   "status",
				Operator: NotEqual,
				Value:    "active",
			},
		},
	}

	for _, tt := range tests {
		result := NewWhereWithTable(tt.table, tt.column, tt.operator, tt.values...)
		require.Equal(t, tt.expected, result, "Expected %v, got %v", tt.expected, result)
	}
}

//func TestWhere_Where(t *testing.T) {
//	tests := []struct {
//		where    *Where
//		expected string
//	}{
//		{
//			where: &Where{
//				Table:  "user",
//				Column: "name",
//			},
//			expected: `"user"."name"`,
//		},
//		{
//			where: &Where{
//				Table:  "",
//				Column: "age",
//			},
//			expected: `"age"`,
//		},
//		{
//			where: &Where{
//				Table:  "order",
//				Column: "id",
//			},
//			expected: `"order"."id"`,
//		},
//		{
//			where: &Where{
//				Table:  "product",
//				Column: "price",
//			},
//			expected: `"product"."price"`,
//		},
//		{
//			where: &Where{
//				Table:  "",
//				Column: "status",
//			},
//			expected: `"status"`,
//		},
//	}
//
//	for _, tt := range tests {
//		result := tt.where.Where()
//		require.Equal(t, tt.expected, result, "Expected %v, got %v", tt.expected, result)
//	}
//}

//func TestWhere_QueryMod(t *testing.T) {
//	tests := []struct {
//		where    *Where
//		expected string
//		args     []interface{}
//	}{
//		{
//			where:    &Where{Table: "users", Column: "name", Operator: Equal, Value: "John"},
//			expected: "\"users\".\"name\" = ?",
//			args:     []interface{}{"John"},
//		},
//		{
//			where:    &Where{Table: "users", Column: "age", Operator: GreaterThan, Value: 30},
//			expected: "\"users\".\"age\" > ?",
//			args:     []interface{}{30},
//		},
//		{
//			where:    &Where{Table: "users", Column: "age", Operator: LessThanOrEqual, Value: 40},
//			expected: "\"users\".\"age\" <= ?",
//			args:     []interface{}{40},
//		},
//		{
//			where:    &Where{Table: "users", Column: "status", Operator: In, Value: []interface{}{"active", "pending"}},
//			expected: "\"users\".\"status\" IN ?",
//			args:     []interface{}{"active", "pending"},
//		},
//		{
//			where:    &Where{Table: "users", Column: "name", Operator: Contains, Value: "John"},
//			expected: "\"users\".\"name\" LIKE ?",
//			args:     []interface{}{"%John%"},
//		},
//		{
//			where:    &Where{Table: "users", Column: "created_at", Operator: IsNull},
//			expected: "\"users\".\"created_at\" IS NULL",
//		},
//		{
//			where:    &Where{Table: "users", Column: "created_at", Operator: IsNotNull},
//			expected: "\"users\".\"created_at\" IS NOT NULL",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(fmt.Sprintf("%s %s %v", tt.where.Table, tt.where.Operator, tt.where.Value), func(t *testing.T) {
//			mods := tt.where.QueryMod()
//			require.Len(t, mods, 1)
//
//			// Assert that the generated query matches the expected query string.
//			queryMod := mods[0].Apply
//			query := qm.SQL(queryMod)
//			require.Equal(t, tt.expected, query.Query)
//
//			// Assert the query arguments.
//			require.Equal(t, tt.args, query.Args)
//		})
//	}
//}

func TestWhere_ToString(t *testing.T) {
	tests := []struct {
		where    *Where
		expected string
	}{
		{
			where:    &Where{Table: "User", Column: "name", Operator: Equal, Value: "John"},
			expected: "users.name eq John",
		},
		{
			where:    &Where{Table: "User", Column: "age", Operator: GreaterThan, Value: 30},
			expected: "users.age gt 30",
		},
		{
			where:    &Where{Table: "User", Column: "created_at", Operator: IsNull},
			expected: "users.created_at isNull <nil>",
		},
		{
			where:    &Where{Table: "", Column: "status", Operator: NotEqual, Value: "active"},
			expected: "status ne active",
		},
		{
			where:    &Where{Table: "Order", Column: "amount", Operator: LessThanOrEqual, Value: 100.50},
			expected: "orders.amount le 100.5",
		},
		{
			where:    &Where{Table: "Product", Column: "name", Operator: Contains, Value: "Laptop"},
			expected: "products.name contains Laptop",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.where.Table, tt.where.Operator), func(t *testing.T) {
			result := tt.where.ToString()
			require.Equal(t, tt.expected, result)
		})
	}
}
