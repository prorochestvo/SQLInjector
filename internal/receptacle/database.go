package receptacle

import (
	"database/sql"
	"errors"
	_ "github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/prorochestvo/sqlinjector/internal"
	"time"
)

func NewPostgreSQL(port int, userLogin, userPassword, databaseName string) (*DBContainer, error) {
	container, src, err := RunPostgreSqlContainer(port, userLogin, userPassword, databaseName)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(string(internal.DialectPostgreSQL), src)
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
	container, src, err := RunMySqlContainer(port, userLogin, userPassword, databaseName)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(string(internal.DialectMySQL), src)
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
