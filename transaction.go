package sqlinjector

import (
	"context"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/prorochestvo/sqlinjector/internal/transaction"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Rollback executes and rollbacks the given actions in the one transaction.
func Rollback[T any](vault Vault, actions ...Action[T]) (T, error) {
	return transmute(vault, true, actions...)
}

// TransactionRollback executes and rollbacks the given actions in the one transaction.
func TransactionRollback(vault Vault, actions ...transaction.Action) (interface{}, error) {
	return transaction.Rollback(context.Background(), vault, actions)
}

// Commit executes and commits the given actions in the one transaction.
func Commit[T any](vault Vault, actions ...Action[T]) (T, error) {
	return transmute(vault, false, actions...)
}

// TransactionCommit executes and commits the given actions in the one transaction.
func TransactionCommit(vault Vault, actions ...transaction.Action) (interface{}, error) {
	return transaction.Commit(context.Background(), vault, actions)
}

type Action[T any] func(boil.ContextExecutor) (T, error)

// transmute executes and transmutes result to expected type.
// If revoke is true, then rollback the given actions in the one transaction.
// If revoke is false, then commit the given actions in the one transaction.
func transmute[T any](vault internal.Vault, revoke bool, actions ...Action[T]) (res T, err error) {
	var tmp interface{}

	a := make([]transaction.Action, len(actions))
	for i, action := range actions {
		a[i] = func(a Action[T]) transaction.Action {
			return func(executor boil.ContextExecutor) (interface{}, error) {
				return a(executor)
			}
		}(action)
	}

	if revoke {
		tmp, err = TransactionRollback(vault, a...)
	} else {
		tmp, err = TransactionCommit(vault, a...)
	}

	if err != nil {
		return
	}

	res, ok := tmp.(T)
	if !ok {
		err = fmt.Errorf("could not convert %T to %T", tmp, res)
		return
	}

	return
}
