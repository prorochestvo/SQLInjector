package expression

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var _ expression = &Where{}

func TestWhere_NewWhereFrom(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		expected []*Where
		hasError bool
	}{
		{
			name:   "Valid filter with eq operator",
			filter: "User.name eq 'John'",
			expected: []*Where{
				{Table: "User", Column: "name", Operator: Equal, Value: "John"},
			},
			hasError: false,
		},
		{
			name:   "Valid filter with gt operator",
			filter: "age gt 30",
			expected: []*Where{
				{Table: "", Column: "age", Operator: GreaterThan, Value: "30"},
			},
			hasError: false,
		},
		{
			name:   "Valid filter with multiple conditions",
			filter: "User.age lt 25, User.name eq 'Alice'",
			expected: []*Where{
				{Table: "User", Column: "age", Operator: LessThan, Value: "25"},
				{Table: "User", Column: "name", Operator: Equal, Value: "Alice"},
			},
			hasError: false,
		},
		{
			name:     "Invalid filter format",
			filter:   "invalidFilter",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewWhereFrom(tt.filter)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestWhere_NewWhere(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		operator Operator
		values   []interface{}
		expected *Where
	}{
		{
			name:     "Valid filter with equal operator",
			column:   "User.name",
			operator: Equal,
			values:   []interface{}{"John"},
			expected: &Where{
				Table:    "",
				Column:   "User.name",
				Operator: Equal,
				Value:    "John",
			},
		},
		{
			name:     "Valid filter with greater-than operator",
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
		t.Run(tt.name, func(t *testing.T) {
			result := NewWhere(tt.column, tt.operator, tt.values...)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestWhere_NewWhereWithTable(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		column   string
		operator Operator
		values   []interface{}
		expected *Where
	}{
		{
			name:     "Valid filter with equal operator and table",
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
			name:     "Valid filter with greater-than operator and table",
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
			name:     "Valid filter with in operator and table",
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
			name:     "Valid filter with not-in operator and table",
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
			name:     "Valid filter with not-equal operator without table",
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
		t.Run(tt.name, func(t *testing.T) {
			result := NewWhereWithTable(tt.table, tt.column, tt.operator, tt.values...)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestWhere_Where(t *testing.T) {
	tests := []struct {
		name     string
		where    *Where
		expected string
	}{
		{
			name:     "Table and column with user prefix",
			where:    &Where{Table: "user", Column: "name"},
			expected: `"users"."name"`,
		},
		{
			name:     "Column without table",
			where:    &Where{Table: "", Column: "age"},
			expected: `"age"`,
		},
		{
			name:     "Table and column with order prefix",
			where:    &Where{Table: "order", Column: "id"},
			expected: `"orders"."id"`,
		},
		{
			name:     "Table and column with product prefix",
			where:    &Where{Table: "product", Column: "price"},
			expected: `"products"."price"`,
		},
		{
			name:     "Column without table (status)",
			where:    &Where{Table: "", Column: "status"},
			expected: `"status"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.where.Where()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestWhere_QueryMod(t *testing.T) {
	tests := []struct {
		name     string
		where    *Where
		expected string
		args     []interface{}
	}{
		{
			name:     "Equal operator with user name",
			where:    &Where{Table: "user", Column: "name", Operator: Equal, Value: "John"},
			expected: "\"users\".\"name\" = ?",
			args:     []interface{}{"John"},
		},
		{
			name:     "GreaterThan operator with user age",
			where:    &Where{Table: "user", Column: "age", Operator: GreaterThan, Value: 30},
			expected: "\"users\".\"age\" > ?",
			args:     []interface{}{30},
		},
		{
			name:     "LessThanOrEqual operator with user age",
			where:    &Where{Table: "user", Column: "age", Operator: LessThanOrEqual, Value: 40},
			expected: "\"users\".\"age\" <= ?",
			args:     []interface{}{40},
		},
		{
			name:     "In operator with user status",
			where:    &Where{Table: "user", Column: "status", Operator: In, Value: []interface{}{"active", "pending"}},
			expected: "\"users\".\"status\" IN ?",
			args:     []interface{}{"active", "pending"},
		},
		{
			name:     "Contains operator with user name",
			where:    &Where{Table: "user", Column: "name", Operator: Contains, Value: "John"},
			expected: "\"users\".\"name\" LIKE ?",
			args:     []interface{}{"%John%"},
		},
		{
			name:     "IsNull operator with user created_at",
			where:    &Where{Table: "user", Column: "created_at", Operator: IsNull},
			expected: "\"users\".\"created_at\" IS NULL",
		},
		{
			name:     "IsNotNull operator with user created_at",
			where:    &Where{Table: "user", Column: "created_at", Operator: IsNotNull},
			expected: "\"users\".\"created_at\" IS NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mods := tt.where.QueryMod()
			require.Len(t, mods, 2)
		})
	}
}

func TestWhere_ToString(t *testing.T) {
	tests := []struct {
		name     string
		where    *Where
		expected string
	}{
		{
			name:     "Equal operator with User name",
			where:    &Where{Table: "User", Column: "name", Operator: Equal, Value: "John"},
			expected: "users.name eq John",
		},
		{
			name:     "GreaterThan operator with User age",
			where:    &Where{Table: "User", Column: "age", Operator: GreaterThan, Value: 30},
			expected: "users.age gt 30",
		},
		{
			name:     "IsNull operator with User created_at",
			where:    &Where{Table: "User", Column: "created_at", Operator: IsNull},
			expected: "users.created_at isNull <nil>",
		},
		{
			name:     "NotEqual operator with status",
			where:    &Where{Table: "", Column: "status", Operator: NotEqual, Value: "active"},
			expected: "status ne active",
		},
		{
			name:     "LessThanOrEqual operator with Order amount",
			where:    &Where{Table: "Order", Column: "amount", Operator: LessThanOrEqual, Value: 100.50},
			expected: "orders.amount le 100.5",
		},
		{
			name:     "Contains operator with Product name",
			where:    &Where{Table: "Product", Column: "name", Operator: Contains, Value: "Laptop"},
			expected: "products.name contains Laptop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.where.ToString()
			require.Equal(t, tt.expected, result)
		})
	}
}
