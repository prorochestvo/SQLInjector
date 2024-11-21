package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Rollback executes and rollbacks the given actions in the one transaction of the given options.
func Rollback(ctx context.Context, vault internal.Vault, a actions) (res interface{}, err error) {
	t, err := vault.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	defer func(t internal.Transaction) { err = errors.Join(err, ignoreChanges(t)) }(t)

	res, err = a.exec(t)

	if err == nil {
		err = t.Rollback()
	}

	return
}

// Commit executes and commits the given actions in the one transaction of the given options.
func Commit(ctx context.Context, vault internal.Vault, a actions) (res interface{}, err error) {
	t, err := vault.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	defer func(t internal.Transaction) { err = errors.Join(err, ignoreChanges(t)) }(t)

	res, err = a.exec(t)

	if err == nil {
		err = t.Commit()
	}

	return
}

type Action func(boil.ContextExecutor) (interface{}, error)

type actions []Action

func (a actions) exec(executor boil.ContextExecutor) (res interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, fmt.Errorf("recovered from panic, details: %s", r))
		}
	}()

	if l := len(a); l == 1 {
		res, err = a[0](executor)
	} else if l > 1 {
		results := make([]interface{}, 0, l)
		for _, act := range a {
			r, e := act(executor)
			results = append(results, r)
			err = errors.Join(err, e)
		}
		res = results
	}
	return
}

func ignoreChanges(t internal.Transaction) error {
	if err := t.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	return nil
}
