package sqldatabase

import (
	"errors"
	"fiveplus/mektep/api/pkg"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"io"
	"sync"
	"testing"
	"time"
)

var _ pkg.SqlDB = (*SqlBase)(nil)

func TestNewDataBaseSQL(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Hour, 0)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	_, err = db.handle.Exec(sqliteTableCreateScript)
	require.NoError(t, err)

	q, err := db.handle.Query(sqliteTableCountScript)
	require.NoError(t, err)
	require.NotNil(t, q)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(q)

	count := -2
	for q.Next() {
		count = -1
		require.NoError(t, q.Scan(&count))
	}
	require.Equal(t, 0, count)
}

func TestSqlBase_Commit(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Hour, 0)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	// init default migration
	_, err = db.Commit(NewAction(sqliteTableCreateScript))
	require.NoError(t, err)

	userID := uuid.NewV4().String()
	userLogin := uuid.NewV4().String()

	// inset new user
	insertUserAction := func(tx boil.ContextExecutor) (interface{}, error) {
		_, e := tx.Exec(fmt.Sprintf(sqliteTableInsertScript, userID, userLogin, uuid.NewV4().String()))
		if e != nil {
			return nil, e
		}
		q, e := tx.Query(fmt.Sprintf(sqliteTableWhereScript, userID))
		if e != nil {
			return nil, e
		}
		defer func(q io.Closer) { _ = q.Close() }(q)
		if !q.Next() {
			return nil, fmt.Errorf("user not found")
		}
		var uID, uLogin, uNickname string
		e = q.Scan(&uID, &uLogin, &uNickname)
		if e != nil {
			return nil, e
		}
		return &user{ID: uID, Login: uLogin, Nickname: uNickname}, nil
	}
	var res interface{}
	res, err = db.Commit(insertUserAction)
	require.NoError(t, err)
	require.NotNil(t, res)
	u, ok := res.(*user)
	require.True(t, ok)
	require.NotNil(t, u)
	require.Equal(t, userID, u.ID)
	require.Equal(t, userLogin, u.Login)

	// sql request without boiler and transaction
	q, err := db.handle.Query(sqliteTableOrderScript)
	require.NoError(t, err)
	require.NotNil(t, q)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(q)

	var users []*user
	for q.Next() {
		id, login, nickname := "", "", ""
		require.NoError(t, q.Scan(&id, &login, &nickname))
		users = append(users, &user{ID: id, Login: login, Nickname: nickname})
	}

	require.Len(t, users, 1)
	require.Equal(t, userID, users[0].ID)
	require.Equal(t, userLogin, users[0].Login)
}

func TestSqlBase_Commit_Error(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Hour, 0)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	// init default migration
	_, err = db.Commit(NewAction(sqliteTableCreateScript))
	require.NoError(t, err)

	userID := uuid.NewV4().String()
	userLogin := uuid.NewV4().String()
	userEmail := uuid.NewV4().String()

	// inset new user
	insertUserAction := func(tx boil.ContextExecutor) (interface{}, error) {
		errs := make([]error, 0, 3)

		_, err = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, uuid.NewV4().String(), uuid.NewV4().String(), uuid.NewV4().String()))
		if err != nil {
			errs = append(errs, err)
		}

		_, err = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, userID, userLogin, userEmail))
		if err != nil {
			errs = append(errs, err)
		}

		_, err = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, uuid.NewV4().String(), uuid.NewV4().String(), uuid.NewV4().String()))
		if err != nil {
			errs = append(errs, err)
		}

		return nil, errors.Join(errs...)
	}
	_, err = db.Commit(insertUserAction)
	require.NoError(t, err)

	// double insert for create error
	_, err = db.Commit(insertUserAction)
	require.Error(t, err)

	// examine users count in db
	q, err := db.handle.Query(sqliteTableCountScript)
	require.NoError(t, err)
	require.NotNil(t, q)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(q)

	count := -1
	for q.Next() {
		count = -1
		require.NoError(t, q.Scan(&count))
	}

	require.Equal(t, 3, count, "users count is incorrect")
}

func TestSqlBase_Commit_Complicated(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Hour, 0)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	// init default migration
	_, err = db.Commit(NewAction(sqliteTableCreateScript))
	require.NoError(t, err)

	// insert new user and select it
	insertUserAction := func(tx boil.ContextExecutor) (interface{}, error) {
		userID := uuid.NewV4().String()
		_, e := tx.Exec(fmt.Sprintf(sqliteTableInsertScript, userID, uuid.NewV4().String(), uuid.NewV4().String()))
		if e != nil {
			return nil, e
		}
		q, e := tx.Query(fmt.Sprintf(sqliteTableWhereScript, userID))
		if e != nil {
			return nil, e
		}
		defer func(q io.Closer) { _ = q.Close() }(q)
		if !q.Next() {
			return nil, fmt.Errorf("user not found")
		}
		var uID, uLogin, uNickname string
		e = q.Scan(&uID, &uLogin, &uNickname)
		if e != nil {
			return nil, e
		}
		return &user{ID: uID, Login: uLogin, Nickname: uNickname}, nil
	}
	var res interface{}
	res, err = db.Commit(insertUserAction, insertUserAction, insertUserAction)
	require.NoError(t, err)
	require.NotNil(t, res)
	u, ok := res.([]interface{})
	require.True(t, ok)
	require.NotNil(t, u)
	require.Len(t, u, 3)
	require.NotNil(t, u[0])
	require.NotNil(t, u[1])
	require.NotNil(t, u[2])
}

func TestSqlBase_Rollback(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Hour, 0)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	// init default migration
	_, err = db.Commit(NewAction(sqliteTableCreateScript))
	require.NoError(t, err)

	userID := uuid.NewV4().String()

	// insert new user and discard it
	_, err = db.Commit(func(tx boil.ContextExecutor) (interface{}, error) {
		_, err = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, userID, uuid.NewV4().String(), uuid.NewV4().String()))
		if err != nil {
			return nil, err
		}
		return true, nil
	})
	require.NoError(t, err)

	// try to find user
	selectUserAction := func(tx boil.ContextExecutor) (interface{}, error) {
		q, e := tx.Query(fmt.Sprintf(sqliteTableWhereScript, userID))
		if e != nil {
			return nil, e
		}
		defer func(q io.Closer) { _ = q.Close() }(q)
		if !q.Next() {
			return nil, fmt.Errorf("user not found")
		}
		var uID, uLogin, uNickname string
		e = q.Scan(&uID, &uLogin, &uNickname)
		if e != nil {
			return nil, e
		}
		return &user{ID: uID, Login: uLogin, Nickname: uNickname}, nil
	}
	var res interface{}
	res, err = db.Rollback(selectUserAction)
	require.NoError(t, err)
	require.NotNil(t, res)
	u, ok := res.(*user)
	require.True(t, ok)
	require.NotNil(t, u)
	require.Equal(t, userID, u.ID)
}

func TestSqlBase_Rollback_Complicated(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Hour, 0)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	// init default migration
	_, err = db.Commit(NewAction(sqliteTableCreateScript))
	require.NoError(t, err)

	// inset new user
	insertUserAction := func(tx boil.ContextExecutor) (interface{}, error) {
		_, err = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, uuid.NewV4().String(), uuid.NewV4().String(), uuid.NewV4().String()))
		if err != nil {
			return nil, err
		}
		return true, nil
	}
	var res interface{}
	res, err = db.Rollback(insertUserAction, insertUserAction)
	require.NoError(t, err)
	require.NotNil(t, res)
	u, ok := res.([]interface{})
	require.True(t, ok)
	require.NotNil(t, u)
	require.Len(t, u, 2)
	require.NotNil(t, u[0])
	require.NotNil(t, u[1])
}

func TestSqlBase_Close(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Hour, 0)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(db io.Closer) { require.Error(t, db.Close()) }(db)

	innerDbHandle := db.handle

	_, err = db.handle.Exec(sqliteTableCreateScript)
	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)
	require.Nil(t, db.handle, "session handle is not nil")

	_, err = innerDbHandle.Query(sqliteTableCountScript)
	require.Error(t, err, "expected behavior of closed session")
}

func TestSqlBase_PoolConnections(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Millisecond*1000, 4)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	wg := sync.WaitGroup{}

	// inset new user
	aSleep300 := func(tx boil.ContextExecutor) (interface{}, error) {
		_, _ = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, uuid.NewV4().String(), uuid.NewV4().String(), uuid.NewV4().String()))
		time.Sleep(time.Millisecond * 300)
		wg.Done()
		return false, nil
	}
	aSleep200 := func(tx boil.ContextExecutor) (interface{}, error) {
		_, _ = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, uuid.NewV4().String(), uuid.NewV4().String(), uuid.NewV4().String()))
		time.Sleep(time.Millisecond * 200)
		wg.Done()
		return false, nil
	}
	aSleep100 := func(tx boil.ContextExecutor) (interface{}, error) {
		_, _ = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, uuid.NewV4().String(), uuid.NewV4().String(), uuid.NewV4().String()))
		time.Sleep(time.Millisecond * 100)
		wg.Done()
		return false, nil
	}
	aSleep050 := func(tx boil.ContextExecutor) (interface{}, error) {
		_, _ = tx.Exec(fmt.Sprintf(sqliteTableInsertScript, uuid.NewV4().String(), uuid.NewV4().String(), uuid.NewV4().String()))
		time.Sleep(time.Millisecond * 50)
		wg.Done()
		return false, nil
	}

	// init default migration
	_, err = db.Commit(NewAction(sqliteTableCreateScript))
	require.NoError(t, err)

	wg.Add(8)
	go func(a func(tx boil.ContextExecutor) (interface{}, error), s *SqlBase) { _, _ = s.Rollback(a) }(aSleep300, db)
	go func(a func(tx boil.ContextExecutor) (interface{}, error), s *SqlBase) { _, _ = s.Rollback(a) }(aSleep200, db)
	go func(a func(tx boil.ContextExecutor) (interface{}, error), s *SqlBase) { _, _ = s.Rollback(a) }(aSleep100, db)
	go func(a func(tx boil.ContextExecutor) (interface{}, error), s *SqlBase) { _, _ = s.Rollback(a) }(aSleep050, db)
	go func(a func(tx boil.ContextExecutor) (interface{}, error), s *SqlBase) { _, _ = s.Rollback(a) }(aSleep300, db)
	go func(a func(tx boil.ContextExecutor) (interface{}, error), s *SqlBase) { _, _ = s.Rollback(a) }(aSleep200, db)
	go func(a func(tx boil.ContextExecutor) (interface{}, error), s *SqlBase) { _, _ = s.Rollback(a) }(aSleep100, db)
	go func(a func(tx boil.ContextExecutor) (interface{}, error), s *SqlBase) { _, _ = s.Rollback(a) }(aSleep050, db)
	wg.Wait()
}

func TestSqlBase_PanicRecovering(t *testing.T) {
	db, err := newDataBaseSQL(dialectSQLite3, ":memory:", time.Millisecond*1000, 4)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.handle, "session handle is invalid")
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	errorMessage := "error message"

	aPanic := func(tx boil.ContextExecutor) (interface{}, error) {
		panic(errorMessage)
		return nil, nil
	}

	errorMessage = uuid.NewV4().String()
	_, err = db.Rollback(aPanic)
	require.Error(t, err)
	require.Contains(t, err.Error(), errorMessage, "panic recovering is failed")

	errorMessage = uuid.NewV4().String()
	_, err = db.Commit(aPanic)
	require.Error(t, err)
	require.Contains(t, err.Error(), errorMessage, "panic recovering is failed")
}

const (
	columnUserID       = "id"
	columnUserLogin    = "login"
	columnUserNickname = "nickname"
)

type user struct {
	ID       string
	Login    string
	Nickname string
}

var sqliteTableCreateScript = `CREATE TABLE ` + tableName + fmt.Sprintf(` (%s TEXT PRIMARY KEY NOT NULL, %s TEXT, %s TEXT);`, columnUserID, columnUserLogin, columnUserNickname)
var sqliteTableInsertScript = `INSERT INTO` + ` ` + tableName + fmt.Sprintf(` (%s, %s, %s)`, columnUserID, columnUserLogin, columnUserNickname) + ` VALUES ('%s', '%s', '%s');`
var sqliteTableCountScript = `SELECT COUNT(*) FROM` + ` ` + tableName
var sqliteTableSelectScript = `SELECT ` + fmt.Sprintf(`%s, %s, %s`, columnUserID, columnUserLogin, columnUserNickname) + ` FROM` + ` ` + tableName
var sqliteTableOrderScript = sqliteTableSelectScript + fmt.Sprintf(` ORDER BY %s ASC;`, columnUserID)
var sqliteTableWhereScript = sqliteTableSelectScript + fmt.Sprintf(` WHERE %s = `, columnUserID) + `'%s';`
