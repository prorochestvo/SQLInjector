package sandbox

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
	"io"
	"testing"
)

func TestNewPool(t *testing.T) {
	pool := NewPool()
	require.NotNil(t, pool)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(pool)

	require.NotNil(t, pool.containers)
	require.Len(t, pool.containers, 0)

	t.Run("NewPostgreSQL", func(t *testing.T) {
		port := 10001
		login, password, database := "u_"+hex.EncodeToString(uuid.NewV4().Bytes()[:4]), "p_"+hex.EncodeToString(uuid.NewV4().Bytes()), "db_"+hex.EncodeToString(uuid.NewV4().Bytes()[:4])
		psql, err := pool.NewPostgreSQL(port, login, password, database)
		require.NoError(t, err)
		require.NotNil(t, psql)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(psql)

		require.Len(t, pool.containers, 1)

		args := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "127.0.0.1", port, login, password, database)
		db, err := sql.Open(string(internal.DialectPostgreSQL), args)
		require.NoError(t, err)
		require.NotNil(t, db)
		require.NoError(t, db.Ping())
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		rows, err := psql.Query("SELECT now()")
		require.NoError(t, err)
		require.NotNil(t, rows)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(rows)

		var now string
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&now))
		require.NotEmpty(t, now)

		psql, err = pool.NewPostgreSQL(port, login, password, database)
		require.NoError(t, err)
		require.NotNil(t, psql)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(psql)

		require.Len(t, pool.containers, 1)
	})
	t.Run("NewMySQL", func(t *testing.T) {
		port := 10002
		login, password, database := "u_"+hex.EncodeToString(uuid.NewV4().Bytes()[:4]), "p_"+hex.EncodeToString(uuid.NewV4().Bytes()), "db_"+hex.EncodeToString(uuid.NewV4().Bytes()[:4])
		msql, err := pool.NewMySQL(port, login, password, database)
		require.NoError(t, err)
		require.NotNil(t, msql)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(msql)

		require.Len(t, pool.containers, 1)

		args := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", login, password, "127.0.0.1", port, database)
		db, err := sql.Open(string(internal.DialectMySQL), args)
		require.NoError(t, err)
		require.NotNil(t, db)
		require.NoError(t, db.Ping())
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		rows, err := msql.Query("SELECT now()")
		require.NoError(t, err)
		require.NotNil(t, rows)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(rows)

		var now string
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&now))
		require.NotEmpty(t, now)

		msql, err = pool.NewMySQL(port, login, password, database)
		require.NoError(t, err)
		require.NotNil(t, msql)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(msql)

		require.Len(t, pool.containers, 1)
	})
	t.Run("NewSQLite3", func(t *testing.T) {
		lsql, err := pool.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, lsql)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(lsql)

		require.Len(t, pool.containers, 0)

		rows, err := lsql.Query("SELECT '1'")
		require.NoError(t, err)
		require.NotNil(t, rows)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(rows)

		var now string
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&now))
		require.NotEmpty(t, now)

		lsql, err = pool.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, lsql)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(lsql)

		require.Len(t, pool.containers, 0)
	})
}
