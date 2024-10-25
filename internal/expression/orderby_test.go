package expression

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var _ expression = &OrderBy{}

func TestNewOrderByFrom(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected []*OrderBy
		wantErr  bool
	}{
		{"Empty expression", "", []*OrderBy{{}}, false}, // Пустое выражение ожидает пустой результат, если функция возвращает пустой слайс
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

func TestNewOrderBy(t *testing.T) {
	tests := []struct {
		column    string
		direction Direction
		expected  *OrderBy
	}{
		{"users/name", Ascending, &OrderBy{Table: "users", Column: "name", direction: Ascending}},
		{"age", Descending, &OrderBy{Table: "", Column: "age", direction: Descending}},
		{"orders/price", Ascending, &OrderBy{Table: "orders", Column: "price", direction: Ascending}},
		{"product_id", Ascending, &OrderBy{Table: "", Column: "product_id", direction: Ascending}},
	}

	for _, tt := range tests {
		result := NewOrderBy(tt.column, tt.direction)

		require.Equal(t, tt.expected, result)
	}
}

func TestNewOrderByWithTable(t *testing.T) {
	tests := []struct {
		table     string
		column    string
		direction Direction
		expected  *OrderBy
	}{
		{"users", "name", Ascending, &OrderBy{Table: "users", Column: "name", direction: Ascending}},
		{"orders", "price", Descending, &OrderBy{Table: "orders", Column: "price", direction: Descending}},
		{"products", "id", Ascending, &OrderBy{Table: "products", Column: "id", direction: Ascending}},
		{"", "created_at", Ascending, &OrderBy{Table: "", Column: "created_at", direction: Ascending}},
	}

	for _, tt := range tests {
		result := NewOrderByWithTable(tt.table, tt.column, tt.direction)

		require.Equal(t, tt.expected, result)
	}
}

func TestOrderBy_OrderBy(t *testing.T) {
	tests := []struct {
		orderBy  *OrderBy
		expected string
	}{
		{&OrderBy{Table: "users", Column: "name", direction: Ascending}, "ASC"},
		{&OrderBy{Table: "orders", Column: "price", direction: Descending}, "DESC"},
		{&OrderBy{Table: "products", Column: "id", direction: ""}, ""},
	}

	for _, tt := range tests {
		result := tt.orderBy.OrderBy()
		require.Equal(t, tt.expected, result)
	}
}

//func TestOrderBy_QueryMod(t *testing.T) {
//	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
//	require.NoError(t, err, "Failed to connect to the in-memory database")
//
//	tests := []struct {
//		orderBy  *OrderBy
//		expected string
//	}{
//		{&OrderBy{Table: "users", Column: "name", direction: Ascending}, `"users"."name" ASC`},
//		{&OrderBy{Table: "orders", Column: "price", direction: Descending}, `"orders"."price" DESC`},
//		{&OrderBy{Table: "products", Column: "id", direction: defaulting}, `"products"."id"`},
//	}
//
//	for _, tt := range tests {
//		mods := tt.orderBy.QueryMod()
//
//		require.Len(t, mods, 1, "Expected 1 mod")
//
//		modStr := fmt.Sprintf("%v", mods[0])
//		t.Logf("Generated mod string: %s", modStr)
//
//
//		require.Equal(t, tt.expected, modStr, "The generated query mod should match the expected string")
//	}
//
//	dbConn, _ := db.DB()
//	dbConn.Close()
//}

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
		require.Equal(t, tt.expected, result, "Expected %v, got %v", tt.expected, result)
	}
}
