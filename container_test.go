package sqlinjector

import (
	"errors"
	"github.com/prorochestvo/sqlinjector/internal/receptacle"
	"github.com/stretchr/testify/require"
	"testing"
)

// Test for NewPostgreSQLSandbox function
func TestNewPostgreSQLSandbox(t *testing.T) {
	port := 5432
	userLogin := "user"
	userPassword := "password"
	databaseName := "testdb"

	// Test case: Successful creation
	db1, err := NewPostgreSQLSandbox(port, userLogin, userPassword, databaseName)
	require.NoError(t, err)
	require.NotNil(t, db1)

	db2, err := NewPostgreSQLSandbox(port, userLogin, userPassword, databaseName)
	require.NoError(t, err)
	require.NotNil(t, db2)
}

// Mock pool with error scenario
type MockPoolWithError struct{}

func (mp *MockPoolWithError) GetOrCreate(name string, creator func() (*receptacle.DBContainer, error)) (*receptacle.DBContainer, string, error) {
	return nil, "", errors.New("mock error")
}
