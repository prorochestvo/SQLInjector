package sandbox

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	_ "github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"strings"
	"sync"
	"time"
)

func RunPostgreSqlContainer(port int, userLogin, userPassword, databaseName string) (*Container, error) {
	defaultPort := 5432
	exposedPort := ""
	if defaultPort == port {
		exposedPort = fmt.Sprintf("%d/tcp", port)
	} else {
		exposedPort = fmt.Sprintf("%d:%d", port, defaultPort)
	}
	internalPort := fmt.Sprintf("%d/tcp", defaultPort)

	container, err := RunTestContainer(
		"postgres:latest",
		[]string{exposedPort},
		map[string]string{
			"POSTGRES_USER":     userLogin,
			"POSTGRES_PASSWORD": userPassword,
			"POSTGRES_DB":       databaseName,
		},
		internalPort,
	)
	if err != nil {
		return nil, err
	}

	cHost, cPort, err := container.MappedHostPort(internalPort)
	if err != nil {
		err = errors.Join(err, container.Close())
		return nil, err
	}

	args := make([]string, 0, 5)
	args = append(args, fmt.Sprintf("host=%s", cHost))
	args = append(args, fmt.Sprintf("port=%s", cPort))
	args = append(args, fmt.Sprintf("user=%s", userLogin))
	args = append(args, fmt.Sprintf("password=%s", userPassword))
	args = append(args, fmt.Sprintf("dbname=%s", databaseName))
	args = append(args, "sslmode=disable")

	container.dialectName = internal.DialectPostgreSQL
	container.dialectSource = strings.Join(args, " ")

	return container, err
}

func RunMySqlContainer(port int, userLogin, userPassword, databaseName string) (*Container, error) {
	defaultPort := 3306
	exposedPort := ""
	if defaultPort == port {
		exposedPort = fmt.Sprintf("%d/tcp", port)
	} else {
		exposedPort = fmt.Sprintf("%d:%d", port, defaultPort)
	}
	internalPort := fmt.Sprintf("%d/tcp", defaultPort)

	container, err := RunTestContainer(
		"mysql:latest",
		[]string{exposedPort},
		map[string]string{
			"MYSQL_ROOT_PASSWORD": userPassword,
			"MYSQL_DATABASE":      databaseName,
			"MYSQL_USER":          userLogin,
			"MYSQL_PASSWORD":      userPassword,
		},
		internalPort,
	)
	if err != nil {
		return nil, err
	}

	cHost, cPort, err := container.MappedHostPort(internalPort)
	if err != nil {
		err = errors.Join(err, container.Close())
		return nil, err
	}

	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", userPassword, cHost, cPort, databaseName)

	container.dialectName = internal.DialectMySQL
	container.dialectSource = args

	return container, err
}

func RunTestContainer(image string, ports []string, env map[string]string, waitPorts ...string) (tc *Container, err error) {
	defer func() {
		if err != nil && tc != nil && tc.logs != nil {
			println(tc.logs.Sting())
		}
	}()

	tc = &Container{logs: &logger{}}

	var strategies []wait.Strategy
	for _, p := range waitPorts {
		strategies = append(strategies, wait.ForListeningPort(nat.Port(p)).WithStartupTimeout(30*time.Second))
	}
	if strategies == nil {
		strategies = append(strategies, wait.ForHealthCheck())
	}

	request := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: ports,
		Env:          env,
		WaitingFor:   wait.ForAll(strategies...),
	}

	c, err := run(request, tc.logs)
	if err != nil || c == nil {
		if err == nil {
			err = fmt.Errorf("container is nil")
		}
		return nil, err
	}

	tc.container = c

	return tc, nil
}

type Container struct {
	container     testcontainers.Container
	closers       []io.Closer
	dialectName   internal.Dialect
	dialectSource string
	logs          *logger
	mutex         sync.RWMutex
}

func (tc *Container) Dialect() internal.Dialect {
	return tc.dialectName
}

func (tc *Container) NewConnection() (internal.Dispatcher, error) {
	db, err := sql.Open(string(tc.dialectName), tc.dialectSource)
	if err != nil || db == nil {
		if err == nil {
			err = errors.New("db handle is invalid")
		}
		return nil, err
	}

	tc.closers = append(tc.closers, db)

	return db, nil
}

func (tc *Container) MappedHostPort(port string) (string, string, error) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	if tc.container == nil {
		return "", "", fmt.Errorf("container is closed")
	}

	ctx := context.Background()

	host, err := tc.container.Host(ctx)
	if err != nil {
		return "", "", err
	}

	p, err := tc.container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return "", "", err
	}

	return host, p.Port(), nil
}

func (tc *Container) Close() error {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if tc.container == nil {
		return nil
	}

	for _, c := range tc.closers {
		_ = c.Close()
	}

	err := tc.container.Terminate(context.Background())
	if err != nil {
		return err
	}

	tc.container = nil

	return nil
}

func run(request testcontainers.ContainerRequest, logging testcontainers.Logging) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
		Logger:           logging,
	})
	if err != nil || container == nil {
		if err == nil {
			err = fmt.Errorf("container is nil")
		}
		return nil, err
	}
	return container, nil
}

type logger struct {
	buffer bytes.Buffer
	m      sync.RWMutex
}

func (l *logger) Sting() string {
	l.m.RLock()
	defer l.m.RUnlock()

	return l.buffer.String()
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.m.Lock()
	defer l.m.Unlock()

	_, _ = fmt.Fprintf(&l.buffer, format, v...)
}
