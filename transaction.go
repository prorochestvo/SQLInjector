package sqlinjector

import (
	"context"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/prorochestvo/sqlinjector/internal/transaction"
)

// Transmute executes and transmutes result to expected type.
// If annul is true, then rollback the given actions in the one transaction.
// If annul is false, then commit the given actions in the one transaction.
func Transmute[T any](dispatcher internal.Dispatcher, annul bool, actions ...transaction.Action) (res T, err error) {
	var tmp interface{}

	if annul {
		tmp, err = Rollback(dispatcher, actions...)
	} else {
		tmp, err = Commit(dispatcher, actions...)
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

// Rollback executes and rollbacks the given actions in the one transaction.
func Rollback(d internal.Dispatcher, actions ...transaction.Action) (interface{}, error) {
	return transaction.Rollback(context.Background(), d, actions)
}

// Commit executes and commits the given actions in the one transaction.
func Commit(d internal.Dispatcher, actions ...transaction.Action) (interface{}, error) {
	return transaction.Commit(context.Background(), d, actions)
}
