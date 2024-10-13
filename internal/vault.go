package internal

import (
	"errors"
	"io"
)

type Vault interface {
	Dialect() Dialect
	VaultID() int64
	Dispatcher
}

func NewSQLVault(d Dialect, db Dispatcher, id int64, c ...io.Closer) Vault {
	return &sqlVault{dialect: d, closers: c, id: id, Dispatcher: db}
}

type sqlVault struct {
	dialect Dialect
	closers []io.Closer
	id      int64
	Dispatcher
}

func (v *sqlVault) Dialect() Dialect {
	return v.dialect
}

func (v *sqlVault) VaultID() int64 {
	return v.id
}

func (v *sqlVault) Close() error {
	var err error
	if v.Dispatcher != nil {
		err = errors.Join(err, v.Dispatcher.Close())
	}
	for _, c := range v.closers {
		err = errors.Join(err, c.Close())
	}
	return err
}
