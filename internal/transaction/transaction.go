package transaction

import (
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Rollbacker interface {
	Rollback(actions ...func(tx boil.ContextExecutor) (interface{}, error)) (interface{}, error)
}

// Rollback executes the given actions and rolls back any changes.
func Rollback[T any](s Rollbacker, a func(tx boil.ContextExecutor) (interface{}, error)) (T, error) {
	var res T

	i, err := s.Rollback(a)
	if err != nil {
		return res, err
	}

	res, ok := i.(T)
	if !ok {
		return res, fmt.Errorf("could not cast %T to %T", i, res)
	}

	return res, nil
}

type Commiter interface {
	Commit(actions ...func(tx boil.ContextExecutor) (interface{}, error)) (interface{}, error)
}

// Commit executes the given actions and try commit any changes in one transaction.
func Commit[T any](s Commiter, a func(tx boil.ContextExecutor) (interface{}, error)) (T, error) {
	var res T

	i, err := s.Commit(a)
	if err != nil {
		return res, err
	}

	res, ok := i.(T)
	if !ok {
		return res, fmt.Errorf("could not cast %T to %T", i, res)
	}

	return res, nil
}

// NewAction creates a new Action from a SQL script.
func NewAction(sqlScript string) func(tx boil.ContextExecutor) (interface{}, error) {
	return func(tx boil.ContextExecutor) (interface{}, error) {
		_, err := tx.Exec(sqlScript)
		return nil, err
	}
}
