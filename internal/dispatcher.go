package internal

import (
	"context"
	"database/sql"
)

type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
}

type Extractor interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type Dispatcher interface {
	Begin() (*sql.Tx, error)
	Executor
	Extractor
}

type DispatcherEx interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	Executor
	Extractor
}

type Transaction interface {
	Commit() error
	Rollback() error
	Executor
	Extractor
}
