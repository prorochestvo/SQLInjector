package expression

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"testing"
)

var _ expression = &OrderBy{}

func TestOrderBy_NewOrderByFrom(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected []*OrderBy
		wantErr  bool
	}{
		{"Empty expression", "", []*OrderBy{{}}, false},
		{"Single column with default direction", "id", []*OrderBy{{Column: "id", Direction: defaulting}}, false},
		{"Single column ascending", "name ASC", []*OrderBy{{Column: "name", Direction: Ascending}}, false},
		{"Single column descending", "age DESC", []*OrderBy{{Column: "age", Direction: Descending}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewOrderByFrom(tt.expr)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestOrderBy_NewOrderBy(t *testing.T) {
	tests := []struct {
		name      string
		column    string
		direction Direction
		expected  *OrderBy
	}{
		{"Column with table", "users/name", Ascending, &OrderBy{Table: "users", Column: "name", Direction: Ascending}},
		{"Column without table", "age", Descending, &OrderBy{Table: "", Column: "age", Direction: Descending}},
		{"Another column with table", "orders/price", Ascending, &OrderBy{Table: "orders", Column: "price", Direction: Ascending}},
		{"Simple column", "product_id", Ascending, &OrderBy{Table: "", Column: "product_id", Direction: Ascending}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewOrderBy(tt.column, tt.direction)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestOrderBy_NewOrderByWithTable(t *testing.T) {
	tests := []struct {
		name      string
		table     string
		column    string
		direction Direction
		expected  *OrderBy
	}{
		{"Column with table users", "users", "name", Ascending, &OrderBy{Table: "users", Column: "name", Direction: Ascending}},
		{"Column with table orders", "orders", "price", Descending, &OrderBy{Table: "orders", Column: "price", Direction: Descending}},
		{"Column with table products", "products", "id", Ascending, &OrderBy{Table: "products", Column: "id", Direction: Ascending}},
		{"Column without table", "", "created_at", Ascending, &OrderBy{Table: "", Column: "created_at", Direction: Ascending}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewOrderByWithTable(tt.table, tt.column, tt.direction)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestOrderBy_OrderBy(t *testing.T) {
	tests := []struct {
		name     string
		orderBy  *OrderBy
		expected string
	}{
		{"Ascending order for users name", &OrderBy{Table: "users", Column: "name", Direction: Ascending}, "ASC"},
		{"Descending order for orders price", &OrderBy{Table: "orders", Column: "price", Direction: Descending}, "DESC"},
		{"No direction for products id", &OrderBy{Table: "products", Column: "id", Direction: ""}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.orderBy.OrderBy()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestOrderBy_QueryMod(t *testing.T) {

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		orderBy  *OrderBy
		expected string
	}{
		{"Ascending order for users name", &OrderBy{Table: "user", Column: "name", Direction: Ascending}, `{"users"."name" ASC []}`},
		{"Descending order for orders price", &OrderBy{Table: "order", Column: "price", Direction: Descending}, `{"orders"."price" DESC []}`},
		{"No direction for products id", &OrderBy{Table: "product", Column: "id", Direction: defaulting}, `{"products"."id" []}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mods := tt.orderBy.QueryMod()
			require.Len(t, mods, 1)

			modStr := fmt.Sprintf("%v", mods[0])
			t.Logf("Generated mod string: %s", modStr)
			require.Equal(t, tt.expected, modStr)
		})
	}

	dbConn, _ := db.DB()
	dbConn.Close()
}

func TestOrderBy_ToString(t *testing.T) {
	tests := []struct {
		orderBy  *OrderBy
		expected string
	}{
		{&OrderBy{Table: "User", Column: "name", Direction: Ascending}, "users.name ASC"},
		{&OrderBy{Table: "Order", Column: "price", Direction: Descending}, "orders.price DESC"},

		{&OrderBy{Table: "Product", Column: "id", Direction: defaulting}, "products.id"},

		{&OrderBy{Table: "", Column: "name", Direction: Ascending}, "name ASC"},
		{&OrderBy{Table: "", Column: "id", Direction: defaulting}, "id"},
	}

	for _, tt := range tests {
		result := tt.orderBy.ToString()
		require.Equal(t, tt.expected, result)
	}
}
