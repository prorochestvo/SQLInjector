package sqlinjector

import (
	"github.com/prorochestvo/sqlinjector/internal/receptacle"
)

// NewPostgreSqlContainer create new connection PostgreSQL via docker container.
// technical function for testing purposes of external packages.
// closing the container is necessary to free the docker resources.
func NewPostgreSqlContainer(port int) (*receptacle.DBContainer, error) {
	return receptacle.NewPostgreSQL(port, "user", "pwd", "dbase")
}

// NewMySqlContainer create new connection MySQL via docker container.
// technical function for testing purposes of external packages.
// closing the container is necessary to free the docker resources.
func NewMySqlContainer(port int) (*receptacle.DBContainer, error) {
	return receptacle.NewMySQL(port, "usr", "pwd", "dbase")
}

// NewSqLiteContainer create new connection SQLite into memory.
// technical function for testing purposes of external packages
// closing the container is necessary to free the docker resources.
func NewSqLiteContainer() (*receptacle.DBContainer, error) {
	return receptacle.NewSQLite3()
}
