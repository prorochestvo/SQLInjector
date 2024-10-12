package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"time"
)

// NewConnection creates a new database session.
func NewConnection(dialect internal.Dialect, source string, parameters ...Parameter) (*sql.DB, error) {
	db, err := sql.Open(string(dialect), source)
	if err != nil || db == nil {
		if err == nil {
			err = errors.New("database handle is invalid")
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

// ConnectionStats returns database statistics.
func ConnectionStats(db *sql.DB) string {
	state := db.Stats()

	tmp := ""
	tmp += fmt.Sprintf("Connections: %d / %d; ", state.OpenConnections, state.MaxOpenConnections)
	tmp += fmt.Sprintf("InUse: %v (idle: %d / %d); ", state.InUse, state.Idle, state.MaxIdleClosed)
	tmp += fmt.Sprintf("WaitCount: %d; ", state.WaitCount)
	tmp += fmt.Sprintf("WaitDuration: %s; ", state.WaitDuration)
	tmp += fmt.Sprintf("MaxIdleTimeClosed: %v;", state.MaxIdleTimeClosed)

	return tmp
}

// ConnectionBurden returns database capacity of connections.
func ConnectionBurden(db *sql.DB) float64 {
	state := db.Stats()
	return float64(state.OpenConnections) * 100 / max(1, float64(state.MaxOpenConnections))
}

// ConnectionMaxLifetime sets the maximum amount of time a connection may be reused.
func ConnectionMaxLifetime(duration time.Duration) Parameter {
	f := func(db *sql.DB) error {
		db.SetConnMaxLifetime(duration)
		return nil
	}
	p := parameter(f)
	return &p
}

// ConnectionMaxIdleTime sets the maximum amount of time a connection may be idle.
func ConnectionMaxIdleTime(duration time.Duration) Parameter {
	f := func(db *sql.DB) error {
		db.SetConnMaxIdleTime(duration)
		return nil
	}
	p := parameter(f)
	return &p
}

// MaxIdleConnection sets the maximum number of idle connections in the database pool.
func MaxIdleConnection(limit int) Parameter {
	f := func(db *sql.DB) error {
		db.SetMaxIdleConns(limit)
		return nil
	}
	p := parameter(f)
	return &p
}

// MaxOpenConnection sets the maximum number of open connections to the database pool.
func MaxOpenConnection(limit int) Parameter {
	f := func(db *sql.DB) error {
		db.SetMaxOpenConns(limit)
		return nil
	}
	p := parameter(f)
	return &p
}

// Parameter is a database connection parameter.
type Parameter interface {
	Apply(*sql.DB) error
}

type parameter func(*sql.DB) error

func (p *parameter) Apply(db *sql.DB) error {
	return (*p)(db)
}
