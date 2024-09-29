package internal

type dialect string

const (
	dialectSQLite3  dialect = "sqlite3"
	dialectMySQL    dialect = "mysql"
	dialectPostgres dialect = "postgres"
)
