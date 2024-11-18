package internal

import (
	"context"
	"database/sql"
	"io"
)

type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
}

type Extractor interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type Vault interface {
	Begin() (*sql.Tx, error)
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	Executor
	Extractor
	io.Closer
}

type Transaction interface {
	Commit() error
	Rollback() error
	Executor
	Extractor
}
