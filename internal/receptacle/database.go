package receptacle

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/prorochestvo/sqlinjector/internal"
	"strings"
	"time"
)

func NewPostgreSQL(port int, userLogin, userPassword, databaseName string) (*DBContainer, error) {
	defaultPort := 5432
	exposedPort := ""
	if defaultPort == port {
		exposedPort = fmt.Sprintf("%d/tcp", port)
	} else {
		exposedPort = fmt.Sprintf("%d:%d", port, defaultPort)
	}
	internalPort := fmt.Sprintf("%d/tcp", defaultPort)

	container, err := NewTestContainers(
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

	db, err := sql.Open(string(internal.DialectPostgreSQL), strings.Join(args, " "))
	if err != nil || db == nil {
		if err == nil {
			err = errors.New("db handle is invalid")
		}
		err = errors.Join(err, container.Close())
		return nil, err
	}

	db.SetConnMaxLifetime(time.Second * 5)
	db.SetConnMaxIdleTime(time.Second)
	db.SetMaxOpenConns(13)
	db.SetMaxIdleConns(3)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DBContainer{DB: db, container: container, dialect: internal.DialectPostgreSQL}, nil
}

func NewMySQL(port int, userLogin, userPassword, databaseName string) (*DBContainer, error) {
	defaultPort := 3306
	exposedPort := ""
	if defaultPort == port {
		exposedPort = fmt.Sprintf("%d/tcp", port)
	} else {
		exposedPort = fmt.Sprintf("%d:%d", port, defaultPort)
	}
	internalPort := fmt.Sprintf("%d/tcp", defaultPort)

	container, err := NewTestContainers(
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", userLogin, userPassword, cHost, cPort, databaseName)

	db, err := sql.Open(string(internal.DialectMySQL), dsn)
	if err != nil || db == nil {
		if err == nil {
			err = errors.New("db handle is invalid")
		}
		err = errors.Join(err, container.Close())
		return nil, err
	}

	db.SetConnMaxLifetime(time.Second * 5)
	db.SetConnMaxIdleTime(time.Second)
	db.SetMaxOpenConns(13)
	db.SetMaxIdleConns(3)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DBContainer{DB: db, container: container, dialect: internal.DialectMySQL}, nil
}

func NewSQLite3() (*DBContainer, error) {
	db, err := sql.Open(string(internal.DialectSQLite3), ":memory:")
	if err != nil || db == nil {
		if err == nil {
			err = errors.New("db handle is invalid")
		}
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute)
	db.SetMaxOpenConns(13)
	db.SetMaxIdleConns(3)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DBContainer{DB: db, container: nil, dialect: internal.DialectSQLite3}, nil
}

type DBContainer struct {
	dialect   internal.Dialect
	container *TestContainer
	*sql.DB
}

func (e *DBContainer) Dialect() internal.Dialect {
	return e.dialect
}

func (e *DBContainer) DataBase() *sql.DB {
	return e.DB
}

func (e *DBContainer) Close() error {
	var err error
	if e.container != nil {
		err = errors.Join(err, e.DB.Close())
	}
	if e.container != nil {
		err = errors.Join(err, e.container.Close())
	}
	return err
}
