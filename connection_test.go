package sqlinjector

import (
	"database/sql"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal/sandbox"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	"testing"
	"time"
)

func TestNewPostgreSQL(t *testing.T) {
	args := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "127.0.0.1", 20000, "test", "test", "test")

	c, err := sandbox.RunPostgreSqlContainer(20000, "test", "test", "test")
	require.NoError(t, err)
	require.NotNil(t, c)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(c)

	db, err := NewPostgreSQL(args, MaxOpenConnection(13))
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	dbStats, err := extractDBStats(db)
	require.NoError(t, err)
	require.NotNil(t, dbStats)
	require.Equal(t, 13, dbStats.MaxOpenConnections)
}

func TestNewMySQL(t *testing.T) {
	args := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", "test", "test", "127.0.0.1", 20000, "test")

	c, err := sandbox.RunMySqlContainer(20000, "test", "test", "test")
	require.NoError(t, err)
	require.NotNil(t, c)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(c)

	db, err := NewMySQL(args, MaxOpenConnection(13))
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	dbStats, err := extractDBStats(db)
	require.NoError(t, err)
	require.NotNil(t, dbStats)
	require.Equal(t, 13, dbStats.MaxOpenConnections)
}

func TestNewSQLite3(t *testing.T) {
	db, err := NewSQLite3(":memory:", MaxOpenConnection(13))
	require.NoError(t, err)
	require.NotNil(t, db)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

	dbStats, err := extractDBStats(db)
	require.NoError(t, err)
	require.NotNil(t, dbStats)
	require.Equal(t, 13, dbStats.MaxOpenConnections)
}

func TestOpenSqlDB(t *testing.T) {
	t.Run("Successful Creation", func(t *testing.T) {
		db, err := openSqlDB("sqlite", ":memory:")
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)
	})
	t.Run("CheckParameter:ConnectionMaxLifetime", func(t *testing.T) {
		agr := time.Duration(rand.Int63n(1000)) * time.Microsecond
		db, err := openSqlDB("sqlite", ":memory:", ConnectionMaxLifetime(agr))
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		t.Skip("could not obtain the value from the *sql.db")
	})
	t.Run("CheckParameter:ConnectionMaxIdleTime", func(t *testing.T) {
		agr := time.Duration(rand.Int63n(1000)) * time.Microsecond
		db, err := openSqlDB("sqlite", ":memory:", ConnectionMaxIdleTime(agr))
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		t.Skip("could not obtain the value from the *sql.db")
	})
	t.Run("CheckParameter:MaxIdleConnection", func(t *testing.T) {
		agr := rand.Intn(1000)
		db, err := openSqlDB("sqlite", ":memory:", MaxIdleConnection(agr))
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		t.Skip("could not obtain the value from the *sql.db")
	})
	t.Run("CheckParameter:MaxOpenConnection", func(t *testing.T) {
		agr := rand.Intn(1000)
		db, err := openSqlDB("sqlite", ":memory:", MaxOpenConnection(agr))
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		dbStats, err := extractDBStats(db)
		require.NoError(t, err)
		require.NotNil(t, dbStats)
		require.Equal(t, agr, dbStats.MaxOpenConnections)
	})
}

func TestBurden(t *testing.T) {
	t.Skip("not implemented")
}

func TestStats(t *testing.T) {
	t.Skip("not implemented")
}

func extractDBStats(db interface{}) (sql.DBStats, error) {
	var res sql.DBStats
	if i, ok := db.(interface {
		Stats() sql.DBStats
	}); ok && i != nil {
		res = i.Stats()
	} else {
		return res, fmt.Errorf("could not extract db stats")
	}
	return res, nil
}
