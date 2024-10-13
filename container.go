package sqlinjector

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/prorochestvo/sqlinjector/internal/receptacle"
	"github.com/twinj/uuid"
	"sync"
)

// NewPostgreSQLSandbox create new connection PostgreSQL via docker container.
// technical function for testing purposes of external packages.
// closing the container is necessary to free the docker resources.
func NewPostgreSQLSandbox(port int, userLogin, userPassword, databaseName string, migrations ...Migration) (*SandBox, error) {
	keySandbox := fmt.Sprintf("%s://%s:%s@127.0.0.1:%d/%s", internal.DialectPostgreSQL, userLogin, userPassword, port, databaseName)

	// get or create container
	dbHndl, dbID, err := pool.NewDB(keySandbox, func() (*receptacle.TestContainer, []string, error) {
		c, s, err := receptacle.RunPostgreSqlContainer(port, userLogin, userPassword, databaseName)
		return c, []string{string(internal.DialectPostgreSQL), s}, err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// make schema
	schemaID := fmt.Sprintf("schema_%0.3d_%s", dbID, hex.EncodeToString(uuid.NewV4().Bytes()))
	schemaSQL := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s; SET search_path TO %s;", schemaID, schemaID)
	if _, err = dbHndl.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("failed to switch schema: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(dbHndl)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &SandBox{dialect: internal.DialectPostgreSQL, DB: dbHndl, key: keySandbox, manager: &pool}, nil
}

// NewMySQLSandBox create new connection MySQL via docker container.
// technical function for testing purposes of external packages.
// closing the container is necessary to free the docker resources.
func NewMySQLSandBox(port int, userLogin, userPassword, databaseName string, migrations ...Migration) (*SandBox, error) {
	keySandbox := fmt.Sprintf("%s://%s:%s@127.0.0.1:%d/%s", internal.DialectMySQL, userLogin, userPassword, port, databaseName)

	// get or create container
	dbHndl, dbID, err := pool.NewDB(keySandbox, func() (*receptacle.TestContainer, []string, error) {
		c, s, err := receptacle.RunMySqlContainer(port, userLogin, userPassword, databaseName)
		return c, []string{string(internal.DialectMySQL), s}, err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// make schema
	schemaID := fmt.Sprintf("schema_%0.3d_%s", dbID, hex.EncodeToString(uuid.NewV4().Bytes()))
	schemaSQL := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s; SET search_path TO %s;", schemaID, schemaID)
	if _, err = dbHndl.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("failed to switch schema: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(dbHndl)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &SandBox{dialect: internal.DialectMySQL, DB: dbHndl, key: keySandbox, manager: &pool}, nil
}

// NewSQLiteSandBox create new connection SQLite into memory.
// technical function for testing purposes of external packages
// closing the container is necessary to free the docker resources.
func NewSQLiteSandBox(migrations ...Migration) (*SandBox, error) {
	c, err := receptacle.NewSQLite3()
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(c.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &SandBox{dialect: internal.DialectSQLite3, DB: c.DB}, nil
}

type containerPollManager struct {
	containers *receptacle.ContainerPoll
	counter    map[string]int64
	m          sync.RWMutex
	number     uint64
}

func (cpm *containerPollManager) NewDB(key string, creator func() (*receptacle.TestContainer, []string, error), parameters ...Parameter) (*sql.DB, uint64, error) {
	cpm.m.RLock()
	defer cpm.m.RUnlock()

	var c *receptacle.TestContainer
	var a []string

	if _, exists := cpm.counter[key]; exists {
		var err error
		c, a, err = cpm.containers.Get(key)
		if err != nil {
			return nil, 0, err
		}
	} else {
		var err error
		c, a, err = creator()
		if err != nil {
			return nil, 0, err
		}
		err = pool.Append(key, c, a...)
		if err != nil {
			return nil, 0, err
		}
	}
	if a == nil || len(a) != 2 {
		return nil, 0, fmt.Errorf("invalid source for container %s", key)
	}

	db, err := newConnection(internal.Dialect(a[0]), a[1], parameters...)
	if err != nil {
		return nil, 0, err
	}

	cpm.number++

	return db, cpm.number, nil
}

func (cpm *containerPollManager) Append(key string, container *receptacle.TestContainer, args ...string) error {
	cpm.m.Lock()
	defer cpm.m.Unlock()

	err := cpm.containers.Append(key, container, args...)
	if err != nil {
		return err
	}

	cpm.counter[key]++

	return nil
}

func (cpm *containerPollManager) Release(key string) error {
	cpm.m.Lock()
	defer cpm.m.Unlock()

	count, ok := cpm.counter[key]
	if ok && count > 0 {
		count--
		cpm.counter[key] = count
	}

	if count <= 0 {
		err := cpm.containers.Remove(key)
		if err != nil && !errors.Is(err, receptacle.ErrContainerNotFound) {
			return err
		}
	}

	return nil
}

var pool = containerPollManager{
	containers: receptacle.NewContainerPoll(),
	counter:    make(map[string]int64),
}

type SandBox struct {
	dialect internal.Dialect
	key     string
	manager *containerPollManager
	*sql.DB
}

func (s *SandBox) Dialect() internal.Dialect {
	return s.dialect
}

func (s *SandBox) DataBase() *sql.DB {
	return s.DB
}

func (s *SandBox) Close() error {
	if s.key != "" || s.manager == nil {
		return s.DB.Close()
	}
	return errors.Join(s.DB.Close(), s.manager.Release(s.key))
}
