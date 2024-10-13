package sqlinjector

import (
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/prorochestvo/sqlinjector/internal/sandbox"
)

// NewSandboxOfPostgreSQL create new connection PostgreSQL via docker container.
// technical function for testing purposes of external packages.
// closing the container is necessary to free the docker resources.
func NewSandboxOfPostgreSQL(port int, migrations ...Migration) (internal.Dispatcher, error) {
	db, err := poll.NewPostgreSQL(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgresql container: %w", err)
	}

	// make schema
	schemaID := fmt.Sprintf("schema_%0.3d", db.VaultID())
	schemaSQL := fmt.Sprintf("CREATE SCHEMA %s; SET search_path TO %s;", schemaID, schemaID)
	if _, err = db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("failed to switch schema: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// NewSandboxOfMySQL create new connection MySQL via docker container.
// technical function for testing purposes of external packages.
// closing the container is necessary to free the docker resources.
func NewSandboxOfMySQL(port int, migrations ...Migration) (internal.Dispatcher, error) {
	db, err := poll.NewMySQL(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgresql container: %w", err)
	}

	// make database
	schemaID := fmt.Sprintf("schema_%0.3d", db.VaultID())
	if _, err = db.Exec(fmt.Sprintf("CREATE DATABASE"+" %s;", schemaID)); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	if _, err = db.Exec(fmt.Sprintf("USE %s;", schemaID)); err != nil {
		return nil, fmt.Errorf("failed to switch database: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// NewSandboxOfSQLite3 create new connection SQLite into memory.
// technical function for testing purposes of external packages
// closing the container is necessary to free the docker resources.
func NewSandboxOfSQLite3(migrations ...Migration) (internal.Dispatcher, error) {
	db, err := poll.NewSQLite3()
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite container: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

//type containerPollManager struct {
//	containers *sandbox.Pool
//	counter    map[string]int64
//	mutex      sync.RWMutex
//	number     uint64
//}
//
//func (cpm *containerPollManager) NewSandBox(key string, creator func() (*sandbox.Container, string, string, error), parameters ...Parameter) (*sql.DB, uint64, error) {
//	cpm.mutex.RLock()
//	defer cpm.mutex.RUnlock()
//
//	var c *sandbox.Container
//	var d, s string
//
//	if _, exists := cpm.counter[key]; exists {
//		var err error
//		_, d, s, err = cpm.containers.Retrieve(key)
//		if err != nil {
//			return nil, 0, err
//		}
//		cpm.counter[key]++
//	} else {
//		var err error
//		c, d, s, err = creator()
//		if err != nil {
//			return nil, 0, err
//		}
//		err = cpm.containers.Append(key, c, d, s)
//		if err != nil {
//			return nil, 0, err
//		}
//		cpm.counter[key]++
//	}
//	if d == "" || s == "" {
//		return nil, 0, fmt.Errorf("invalid source for container %s", key)
//	}
//
//	db, err := openSqlDB(internal.Dialect(d), s, parameters...)
//	if err != nil {
//		return nil, 0, err
//	}
//
//	cpm.number++
//
//	return db, cpm.number, nil
//}
//
//func (cpm *containerPollManager) Append(key string, container *sandbox.Container, driverName, dataSourceName string) error {
//	cpm.mutex.Lock()
//	defer cpm.mutex.Unlock()
//
//	err := cpm.containers.Append(key, container, driverName, dataSourceName)
//	if err != nil {
//		return err
//	}
//
//	cpm.counter[key]++
//
//	return nil
//}
//
//func (cpm *containerPollManager) Release(key string) error {
//	cpm.mutex.Lock()
//	defer cpm.mutex.Unlock()
//
//	count, ok := cpm.counter[key]
//	if ok && count > 0 {
//		count--
//		cpm.counter[key] = count
//	}
//
//	if count <= 0 {
//		err := cpm.containers.Remove(key)
//		delete(cpm.counter, key)
//		if err != nil && !errors.Is(err, sandbox.ErrContainerNotFound) {
//			return err
//		}
//	}
//
//	return nil
//}
//
//var pool = containerPollManager{
//	containers: sandbox.NewPool(),
//	counter:    make(map[string]int64),
//	number:     0,
//}
//
//type SandBox struct {
//	dialect internal.Dialect
//	key     string
//	manager *containerPollManager
//	*sql.DB
//}
//
//func (s *SandBox) Dialect() internal.Dialect {
//	return s.dialect
//}
//
//func (s *SandBox) DataBase() *sql.DB {
//	return s.DB
//}
//
//func (s *SandBox) Close() error {
//	err := s.DB.Close()
//	if s.key != "" && s.manager != nil {
//		err = errors.Join(err, s.manager.Release(s.key))
//	}
//	return err
//}

var poll = sandbox.NewPool()
