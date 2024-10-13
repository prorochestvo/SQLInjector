package internal

type Dialect string

const (
	DialectSQLite3    Dialect = "sqlite"
	DialectMySQL      Dialect = "mysql"
	DialectPostgreSQL Dialect = "postgres"
)
