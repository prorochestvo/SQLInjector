package sandbox

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"sync"
	"time"
)

func NewPool() *Pool {
	return &Pool{
		containers: make(map[string]*Container, 2),
		counter:    make(map[string]int64, 2),
	}
}

type Pool struct {
	containers map[string]*Container
	counter    map[string]int64
	mutex      sync.Mutex
}

func (p *Pool) NewPostgreSQL(port int, options ...string) (internal.Vault, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	n := fmt.Sprintf("%s://%s:%d", internal.DialectPostgreSQL, "localhost", port)

	container, ok := p.containers[n]
	if !ok {
		userLogin, userPassword, databaseName := "username", "password", "postgres"
		for i, o := range options {
			switch i {
			case 0:
				userLogin = o
			case 1:
				userPassword = o
			case 2:
				databaseName = o
			}
		}
		c, err := RunPostgreSqlContainer(port, userLogin, userPassword, databaseName)
		if err != nil {
			return nil, err
		}
		container = c
		p.containers[n] = container
	}

	sqlBase, err := container.NewConnection()
	if err != nil {
		return nil, err
	}

	if i, ok := sqlBase.(interface {
		Ping() error
	}); ok && i != nil {
		err = i.Ping()
		if err != nil {
			return nil, err
		}
	}

	p.counter[n]++

	v := NewStubVault(
		container.Dialect(),
		sqlBase,
		p.counter[n],
		&releaser{Pool: p, Name: n},
	)

	return v, nil
}

func (p *Pool) NewMySQL(port int, options ...string) (internal.Vault, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	n := fmt.Sprintf("%s://%s:%d", internal.DialectMySQL, "localhost", port)

	container, ok := p.containers[n]
	if !ok {
		userLogin, userPassword, databaseName := "username", "password", "mysql"
		for i, o := range options {
			switch i {
			case 0:
				userLogin = o
			case 1:
				userPassword = o
			case 2:
				databaseName = o
			}
		}
		c, err := RunMySqlContainer(port, userLogin, userPassword, databaseName)
		if err != nil {
			return nil, err
		}
		container = c
		p.containers[n] = container
	}

	sqlBase, err := container.NewConnection()
	if err != nil {
		return nil, err
	}

	if i, ok := sqlBase.(interface {
		Ping() error
	}); ok && i != nil {
		err = i.Ping()
		if err != nil {
			return nil, err
		}
	}

	p.counter[n]++

	v := NewStubVault(
		container.Dialect(),
		sqlBase,
		p.counter[n],
		&releaser{Pool: p, Name: n},
	)

	return v, nil
}

func (p *Pool) NewSQLite3() (internal.Vault, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	n := fmt.Sprintf("%s://%s:%s", internal.DialectMySQL, "localhost", "memory")

	sqlBase, err := sql.Open(string(internal.DialectSQLite3), ":memory:")
	if err != nil || sqlBase == nil {
		if err == nil {
			err = errors.New("sqlBase handle is invalid")
		}
		return nil, err
	}

	sqlBase.SetConnMaxLifetime(time.Minute * 5)
	sqlBase.SetConnMaxIdleTime(time.Minute)
	sqlBase.SetMaxOpenConns(13)
	sqlBase.SetMaxIdleConns(3)

	err = sqlBase.Ping()
	if err != nil {
		return nil, err
	}

	p.counter[n]++

	v := NewStubVault(
		internal.DialectSQLite3,
		sqlBase,
		p.counter[n],
		&releaser{Pool: p, Name: n},
	)

	return v, nil
}

func (p *Pool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var err error
	for n, container := range p.containers {
		err = errors.Join(err, container.Close())
		p.containers[n] = nil
		p.counter[n] = 0
	}

	p.containers = make(map[string]*Container)
	p.counter = make(map[string]int64)

	return err
}

func (p *Pool) remove(n string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if c, ok := p.counter[n]; !ok {
		return ErrContainerNotFound
	} else if c > 1 {
		p.counter[n] = c - 1
		return nil
	}

	var err error
	if c, ok := p.containers[n]; ok && c != nil {
		err = c.Close()
	}

	p.containers[n] = nil // garbage collector optimization

	delete(p.containers, n)
	delete(p.counter, n)

	if len(p.containers) == 0 {
		p.containers = make(map[string]*Container) // garbage collector optimization
	}

	return err
}

var ErrContainerNotFound = errors.New("container not found")
