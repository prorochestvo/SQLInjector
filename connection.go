package sqlinjector

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"time"
)

func NewPostgreSQL(source string, parameters ...Parameter) (internal.Dispatcher, error) {
	return openSqlDB(internal.DialectPostgreSQL, source, parameters...)
}

func NewMySQL(source string, parameters ...Parameter) (internal.Dispatcher, error) {
	return openSqlDB(internal.DialectMySQL, source, parameters...)
}

func NewSQLite3(source string, parameters ...Parameter) (internal.Dispatcher, error) {
	return openSqlDB(internal.DialectSQLite3, source, parameters...)
}

// openSqlDB creates a new database session.
func openSqlDB(dialect internal.Dialect, source string, parameters ...Parameter) (internal.Dispatcher, error) {
	db, err := sql.Open(string(dialect), source)
	if err != nil || db == nil {
		if err == nil {
			err = fmt.Errorf("database %s handle is invalid", dialect)
		}
		return nil, err
	}

	for _, p := range parameters {
		err = errors.Join(err, p.Apply(db))
	}
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, err
}

// Stats returns database statistics.
func Stats(d internal.Dispatcher) string {
	res := ""

	if i, ok := d.(interface {
		Stats() sql.DBStats
	}); ok && i != nil {
		state := i.Stats()
		res += fmt.Sprintf("Connections: %d / %d; ", state.OpenConnections, state.MaxOpenConnections)
		res += fmt.Sprintf("InUse: %v (idle: %d / %d); ", state.InUse, state.Idle, state.MaxIdleClosed)
		res += fmt.Sprintf("WaitCount: %d; ", state.WaitCount)
		res += fmt.Sprintf("WaitDuration: %s; ", state.WaitDuration)
		res += fmt.Sprintf("MaxIdleTimeClosed: %v;", state.MaxIdleTimeClosed)
	}

	return res
}

// Burden returns database capacity of connections.
func Burden(d internal.Dispatcher) float64 {
	res := 0.0

	if i, ok := d.(interface {
		Stats() sql.DBStats
	}); ok && i != nil {
		state := i.Stats()
		res = float64(state.OpenConnections) * 100 / max(1, float64(state.MaxOpenConnections))
	}

	return res
}

// ConnectionMaxLifetime sets the maximum amount of time a connection may be reused.
func ConnectionMaxLifetime(duration time.Duration) Parameter {
	f := func(db interface{}) error {
		if i, ok := db.(interface {
			SetConnMaxLifetime(d time.Duration)
		}); ok && i != nil {
			i.SetConnMaxLifetime(duration)
		}
		return nil
	}
	p := parameter(f)
	return &p
}

// ConnectionMaxIdleTime sets the maximum amount of time a connection may be idle.
func ConnectionMaxIdleTime(duration time.Duration) Parameter {
	f := func(db interface{}) error {
		if i, ok := db.(interface {
			SetConnMaxIdleTime(d time.Duration)
		}); ok && i != nil {
			i.SetConnMaxIdleTime(duration)
		}
		return nil
	}
	p := parameter(f)
	return &p
}

// MaxIdleConnection sets the maximum number of idle connections in the database pool.
func MaxIdleConnection(limit int) Parameter {
	f := func(db interface{}) error {
		if i, ok := db.(interface {
			SetMaxIdleConns(n int)
		}); ok && i != nil {
			i.SetMaxIdleConns(limit)
		}
		return nil
	}
	p := parameter(f)
	return &p
}

// MaxOpenConnection sets the maximum number of open connections to the database pool.
func MaxOpenConnection(limit int) Parameter {
	f := func(db interface{}) error {
		if i, ok := db.(interface {
			SetMaxOpenConns(n int)
		}); ok && i != nil {
			i.SetMaxOpenConns(limit)
		}
		return nil
	}
	p := parameter(f)
	return &p
}

// Parameter is a database connection parameter.
type Parameter interface {
	Apply(internal.Dispatcher) error
}

type parameter func(interface{}) error

func (p *parameter) Apply(db internal.Dispatcher) error {
	return (*p)(db)
}
