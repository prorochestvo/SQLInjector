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
		{"Single column with default direction", "id", []*OrderBy{{Column: "id", direction: defaulting}}, false},
		{"Single column ascending", "name ASC", []*OrderBy{{Column: "name", direction: Ascending}}, false},
		{"Single column descending", "age DESC", []*OrderBy{{Column: "age", direction: Descending}}, false},
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
		{"Column with table", "users/name", Ascending, &OrderBy{Table: "users", Column: "name", direction: Ascending}},
		{"Column without table", "age", Descending, &OrderBy{Table: "", Column: "age", direction: Descending}},
		{"Another column with table", "orders/price", Ascending, &OrderBy{Table: "orders", Column: "price", direction: Ascending}},
		{"Simple column", "product_id", Ascending, &OrderBy{Table: "", Column: "product_id", direction: Ascending}},
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
		{"Column with table users", "users", "name", Ascending, &OrderBy{Table: "users", Column: "name", direction: Ascending}},
		{"Column with table orders", "orders", "price", Descending, &OrderBy{Table: "orders", Column: "price", direction: Descending}},
		{"Column with table products", "products", "id", Ascending, &OrderBy{Table: "products", Column: "id", direction: Ascending}},
		{"Column without table", "", "created_at", Ascending, &OrderBy{Table: "", Column: "created_at", direction: Ascending}},
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
		{"Ascending order for users name", &OrderBy{Table: "users", Column: "name", direction: Ascending}, "ASC"},
		{"Descending order for orders price", &OrderBy{Table: "orders", Column: "price", direction: Descending}, "DESC"},
		{"No direction for products id", &OrderBy{Table: "products", Column: "id", direction: ""}, ""},
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
		{"Ascending order for users name", &OrderBy{Table: "user", Column: "name", direction: Ascending}, `{"users"."name" ASC []}`},
		{"Descending order for orders price", &OrderBy{Table: "order", Column: "price", direction: Descending}, `{"orders"."price" DESC []}`},
		{"No direction for products id", &OrderBy{Table: "product", Column: "id", direction: defaulting}, `{"products"."id" []}`},
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
		{&OrderBy{Table: "User", Column: "name", direction: Ascending}, "users.name ASC"},
		{&OrderBy{Table: "Order", Column: "price", direction: Descending}, "orders.price DESC"},

		{&OrderBy{Table: "Product", Column: "id", direction: defaulting}, "products.id"},

		{&OrderBy{Table: "", Column: "name", direction: Ascending}, "name ASC"},
		{&OrderBy{Table: "", Column: "id", direction: defaulting}, "id"},
	}

	for _, tt := range tests {
		result := tt.orderBy.ToString()
		require.Equal(t, tt.expected, result)
	}
}
