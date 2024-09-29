package sqldatabase

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"time"
)

// newDataBaseSQL creates a new database session.
func newDataBaseSQL(
	dialect dialect,
	source string,
	lifetime time.Duration,
	capacity int,
) (*SqlBase, error) {
	db, err := sql.Open(string(dialect), source)
	if err != nil || db == nil {
		if err == nil {
			err = errors.New("db handle is invalid")
		}
		return nil, err
	}

	capacity = max(capacity, 2)
	lifetime = max(lifetime, time.Second*15)

	// important settings
	db.SetConnMaxLifetime(lifetime)
	db.SetConnMaxIdleTime(lifetime >> 1)
	db.SetMaxOpenConns(capacity)
	db.SetMaxIdleConns(capacity >> 1)

	s := &SqlBase{
		dialect: dialect,
		handle:  db,
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return s, err
}

// SqlBase is a database session.
type SqlBase struct {
	dialect dialect
	handle  *sql.DB
}

// Dialect returns the database dialect.
func (s *SqlBase) Dialect() string {
	return string(s.dialect)
}

// db returns the default database handle.
func (s *SqlBase) db() *sql.DB {
	return s.handle
}

// ExecRawSQL executes the raw SQL query without any additional processing.
func (s *SqlBase) ExecRawSQL(sqlQuery string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, fmt.Errorf("recovered from panic, details: %s", r))
		}
	}()
	_, err = s.handle.Exec(sqlQuery)
	return
}

// Rollback rolls back the given actions in the one transaction.
func (s *SqlBase) Rollback(actions ...func(tx boil.ContextExecutor) (interface{}, error)) (res interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, fmt.Errorf("recovered from panic, details: %s", r))
		}
	}()

	tx, err := s.handle.Begin()
	if err != nil {
		return
	}
	defer func(tx *sql.Tx) {
		if e := tx.Rollback(); e != nil && !errors.Is(e, sql.ErrTxDone) {
			err = errors.Join(err, e)
		}
	}(tx)

	if l := len(actions); l == 1 {
		res, err = actions[0](tx)
	} else if l > 1 {
		results := make([]interface{}, 0, l)
		for _, action := range actions {
			r, e := action(tx)
			results = append(results, r)
			err = errors.Join(err, e)
		}
		res = results
	}

	return res, err
}

// Commit commits the given actions in the one transaction.
func (s *SqlBase) Commit(actions ...func(tx boil.ContextExecutor) (interface{}, error)) (res interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, fmt.Errorf("recovered from panic, details: %s", r))
		}
	}()

	tx, err := s.handle.Begin()
	if err != nil {
		return
	}
	defer func(tx *sql.Tx) {
		if e := tx.Rollback(); e != nil && !errors.Is(e, sql.ErrTxDone) {
			err = errors.Join(err, e)
		}
	}(tx)

	if l := len(actions); l == 1 {
		res, err = actions[0](tx)
	} else if l > 1 {
		results := make([]interface{}, 0, l)
		for _, action := range actions {
			r, e := action(tx)
			results = append(results, r)
			err = errors.Join(err, e)
		}
		res = results
	}

	if err == nil {
		err = tx.Commit()
	}

	return res, err
}

// Stats returns database statistics.
func (s *SqlBase) Stats() string {
	state := s.handle.Stats()

	tmp := ""
	tmp += fmt.Sprintf("Connections: %d / %d; ", state.OpenConnections, state.MaxOpenConnections)
	tmp += fmt.Sprintf("InUse: %v (idle: %d / %d); ", state.InUse, state.Idle, state.MaxIdleClosed)
	tmp += fmt.Sprintf("WaitCount: %d; ", state.WaitCount)
	tmp += fmt.Sprintf("WaitDuration: %s; ", state.WaitDuration)
	tmp += fmt.Sprintf("MaxIdleTimeClosed: %v;", state.MaxIdleTimeClosed)

	return tmp
}

func (s *SqlBase) Burden() float64 {
	state := s.handle.Stats()
	return float64(state.OpenConnections) * 100 / max(1, float64(state.MaxOpenConnections))
}

// Close closes the database connection.
func (s *SqlBase) Close() error {
	if s.handle == nil {
		return errors.New("db handle is invalid")
	}

	err := s.handle.Close()
	if err == nil {
		s.handle = nil
	}

	return err
}

func init() {
	boil.SetLocation(time.UTC)
}
