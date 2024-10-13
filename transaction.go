package sqlinjector

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal/transaction"
)

// Transmute executes and transmutes result to expected type.
// If annul is true, then rollback the given actions in the one transaction.
// If annul is false, then commit the given actions in the one transaction.
func Transmute[T any](db *sql.DB, annul bool, actions ...transaction.Action) (res T, err error) {
	var tmp interface{}

	if annul {
		tmp, err = Rollback(db, actions...)
	} else {
		tmp, err = Commit(db, actions...)
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
func Rollback(db *sql.DB, actions ...transaction.Action) (interface{}, error) {
	return transaction.Rollback(context.Background(), db, actions)
}

// Commit executes and commits the given actions in the one transaction.
func Commit(db *sql.DB, actions ...transaction.Action) (interface{}, error) {
	return transaction.Commit(context.Background(), db, actions)
}
