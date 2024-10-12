package internal

type Dialect string

const (
	DialectSQLite3  Dialect = "sqlite"
	DialectMySQL    Dialect = "mysql"
	DialectPostgres Dialect = "postgres"
)
