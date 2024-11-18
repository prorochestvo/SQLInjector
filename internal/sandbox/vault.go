package sandbox

import (
	"errors"
	"github.com/prorochestvo/sqlinjector/internal"
	"io"
)

func NewStubVault(d internal.Dialect, db internal.Vault, id int64, c ...io.Closer) internal.Vault {
	return &sqlVault{dialect: d, closers: c, id: id, Vault: db}
}

type sqlVault struct {
	dialect internal.Dialect
	closers []io.Closer
	id      int64
	internal.Vault
}

func (v *sqlVault) Dialect() internal.Dialect {
	return v.dialect
}

func (v *sqlVault) VaultID() int64 {
	return v.id
}

func (v *sqlVault) Close() error {
	var err error
	if v.Vault != nil {
		err = errors.Join(err, v.Vault.Close())
	}
	for _, c := range v.closers {
		err = errors.Join(err, c.Close())
	}
	return err
}

type releaser struct {
	Pool interface {
		remove(name string) error
	}
	Name string
}

func (r *releaser) Close() error {
	if r.Pool == nil {
		return nil
	}
	return r.Pool.remove(r.Name)
}
