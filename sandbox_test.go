package sqlinjector

import (
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
	"io"
	"testing"
)

func TestNewSandboxOfPostgreSQL(t *testing.T) {
	t.Run("Successful Creation", func(t *testing.T) {
		var actually string

		db1, err := NewSandboxOfPostgreSQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db1)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db1)

		db2, err := NewSandboxOfPostgreSQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db2)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db2)

		rows, err := db1.Query("SELECT current_schema();")
		require.NoError(t, err)
		require.NotNil(t, rows)

		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&actually))
		require.Equal(t, fmt.Sprintf("schema_%0.3d", 1), actually)

		rows, err = db2.Query("SELECT current_schema();")
		require.NoError(t, err)
		require.NotNil(t, rows)

		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&actually))
		require.Equal(t, fmt.Sprintf("schema_%0.3d", 2), actually)
	})
	t.Run("Successful Migration", func(t *testing.T) {
		tableName01 := "t_01_" + hex.EncodeToString(uuid.NewV4().Bytes()[:4])
		tableName02 := "t_02_" + hex.EncodeToString(uuid.NewV4().Bytes()[:4])
		expected := uuid.NewV4().String()
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val VARCHAR(250));", "DROP TABLE"+" "+tableName01+";", "m0001")
		m2, _ := NewMemoryMigration("CREATE TABLE "+tableName02+" (m2_val VARCHAR(250));", "DROP TABLE"+" "+tableName02+";", "m0002")
		m3, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expected+"')", "DELETE FROM"+" "+tableName01+";", "m0003")

		db, err := NewSandboxOfPostgreSQL(20000, m2, m3, m1)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		var value string
		err = db.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&value)
		require.NoError(t, err)
		require.Equal(t, expected, value)

		var count int
		err = db.QueryRow("SELECT COUNT(*)" + " FROM " + tableName02 + ";").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})
	t.Run("Successful Multiple Creation", func(t *testing.T) {
		var actually string

		tableName01 := "t_01"
		expectedValue01 := uuid.NewV4().String()
		expectedValue02 := uuid.NewV4().String()
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val VARCHAR(250));", "DROP TABLE"+" "+tableName01+";", "m0001")
		m2, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expectedValue01+"')", "DELETE FROM"+" "+tableName01+";", "m0002")
		m3, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expectedValue02+"')", "DELETE FROM"+" "+tableName01+";", "m0003")

		db1, err := NewSandboxOfPostgreSQL(20000, m1, m2)
		require.NoError(t, err)
		require.NotNil(t, db1)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db1)

		db2, err := NewSandboxOfPostgreSQL(20000, m1, m3)
		require.NoError(t, err)
		require.NotNil(t, db2)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db2)

		err = db1.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&actually)
		require.NoError(t, err)
		require.Equal(t, expectedValue01, actually)

		err = db2.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&actually)
		require.NoError(t, err)
		require.Equal(t, expectedValue02, actually)
	})
	t.Run("Failed Migration", func(t *testing.T) {
		tableName01 := "t_01"
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val: STR_VAL);", "DROP TABLE"+" "+tableName01+";", "m0001")

		db, err := NewSandboxOfPostgreSQL(20000, m1)
		require.Error(t, err)
		require.Nil(t, db)
	})
}

func TestNewSandboxOfMySQL(t *testing.T) {
	t.Run("Successful Creation", func(t *testing.T) {
		var actually string

		db1, err := NewSandboxOfMySQL(20001)
		require.NoError(t, err)
		require.NotNil(t, db1)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db1)

		db2, err := NewSandboxOfMySQL(20001)
		require.NoError(t, err)
		require.NotNil(t, db2)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db2)

		rows, err := db1.Query("SELECT DATABASE();")
		require.NoError(t, err)
		require.NotNil(t, rows)

		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&actually))
		require.Equal(t, fmt.Sprintf("schema_%0.3d", 1), actually)

		rows, err = db2.Query("SELECT DATABASE();")
		require.NoError(t, err)
		require.NotNil(t, rows)

		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&actually))
		require.Equal(t, fmt.Sprintf("schema_%0.3d", 2), actually)
	})
	t.Run("Successful Migration", func(t *testing.T) {
		tableName01 := "t_01_" + hex.EncodeToString(uuid.NewV4().Bytes()[:4])
		tableName02 := "t_02_" + hex.EncodeToString(uuid.NewV4().Bytes()[:4])
		expected := uuid.NewV4().String()
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val VARCHAR(250));", "DROP TABLE"+" "+tableName01+";", "m0001")
		m2, _ := NewMemoryMigration("CREATE TABLE "+tableName02+" (m2_val VARCHAR(250));", "DROP TABLE"+" "+tableName02+";", "m0002")
		m3, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expected+"')", "DELETE FROM"+" "+tableName01+";", "m0003")

		db, err := NewSandboxOfMySQL(20001, m2, m3, m1)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		var value string
		err = db.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&value)
		require.NoError(t, err)
		require.Equal(t, expected, value)

		var count int
		err = db.QueryRow("SELECT COUNT(*)" + " FROM " + tableName02 + ";").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})
	t.Run("Successful Multiple Creation", func(t *testing.T) {
		var actually string

		tableName01 := "t_01"
		expectedValue01 := uuid.NewV4().String()
		expectedValue02 := uuid.NewV4().String()
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val VARCHAR(250));", "DROP TABLE"+" "+tableName01+";", "m0001")
		m2, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expectedValue01+"')", "DELETE FROM"+" "+tableName01+";", "m0002")
		m3, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expectedValue02+"')", "DELETE FROM"+" "+tableName01+";", "m0003")

		db1, err := NewSandboxOfMySQL(20001, m1, m2)
		require.NoError(t, err)
		require.NotNil(t, db1)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db1)

		db2, err := NewSandboxOfMySQL(20001, m1, m3)
		require.NoError(t, err)
		require.NotNil(t, db2)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db2)

		err = db1.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&actually)
		require.NoError(t, err)
		require.Equal(t, expectedValue01, actually)

		err = db2.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&actually)
		require.NoError(t, err)
		require.Equal(t, expectedValue02, actually)
	})
	t.Run("Failed Migration", func(t *testing.T) {
		tableName01 := "t_01"
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val: STR_VAL);", "DROP TABLE"+" "+tableName01+";", "m0001")

		db, err := NewSandboxOfMySQL(20001, m1)
		require.Error(t, err)
		require.Nil(t, db)
	})
}

func TestNewSandboxOfSQLite3(t *testing.T) {
	t.Run("Successful Creation", func(t *testing.T) {
		var actually string

		db1, err := NewSandboxOfSQLite3()
		require.NoError(t, err)
		require.NotNil(t, db1)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db1)

		db2, err := NewSandboxOfSQLite3()
		require.NoError(t, err)
		require.NotNil(t, db2)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db2)

		rows, err := db1.Query("SELECT 1;")
		require.NoError(t, err)
		require.NotNil(t, rows)

		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&actually))
		require.NotEmpty(t, actually)

		rows, err = db2.Query("SELECT 2;")
		require.NoError(t, err)
		require.NotNil(t, rows)

		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&actually))
		require.NotEmpty(t, actually)
	})
	t.Run("Successful Migration", func(t *testing.T) {
		tableName01 := "t_01_" + hex.EncodeToString(uuid.NewV4().Bytes()[:4])
		tableName02 := "t_02_" + hex.EncodeToString(uuid.NewV4().Bytes()[:4])
		expected := uuid.NewV4().String()
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val VARCHAR(250));", "DROP TABLE"+" "+tableName01+";", "m0001")
		m2, _ := NewMemoryMigration("CREATE TABLE "+tableName02+" (m2_val VARCHAR(250));", "DROP TABLE"+" "+tableName02+";", "m0002")
		m3, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expected+"')", "DELETE FROM"+" "+tableName01+";", "m0003")

		db, err := NewSandboxOfSQLite3(m2, m3, m1)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		var value string
		err = db.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&value)
		require.NoError(t, err)
		require.Equal(t, expected, value)

		var count int
		err = db.QueryRow("SELECT COUNT(*)" + " FROM " + tableName02 + ";").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})
	t.Run("Successful Multiple Creation", func(t *testing.T) {
		var actually string

		tableName01 := "t_01"
		expectedValue01 := uuid.NewV4().String()
		expectedValue02 := uuid.NewV4().String()
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val VARCHAR(250));", "DROP TABLE"+" "+tableName01+";", "m0001")
		m2, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expectedValue01+"')", "DELETE FROM"+" "+tableName01+";", "m0002")
		m3, _ := NewMemoryMigration("INSERT INTO"+" "+tableName01+" (m1_val) VALUES ('"+expectedValue02+"')", "DELETE FROM"+" "+tableName01+";", "m0003")

		db1, err := NewSandboxOfSQLite3(m1, m2)
		require.NoError(t, err)
		require.NotNil(t, db1)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db1)

		db2, err := NewSandboxOfSQLite3(m1, m3)
		require.NoError(t, err)
		require.NotNil(t, db2)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db2)

		err = db1.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&actually)
		require.NoError(t, err)
		require.Equal(t, expectedValue01, actually)

		err = db2.QueryRow("SELECT m1_val" + " FROM " + tableName01 + " LIMIT 1;").Scan(&actually)
		require.NoError(t, err)
		require.Equal(t, expectedValue02, actually)
	})
	t.Run("Failed Migration", func(t *testing.T) {
		tableName01 := "t_01"
		m1, _ := NewMemoryMigration("CREATE TABLE "+tableName01+" (m1_val: STR_VAL);", "DROP TABLE"+" "+tableName01+";", "m0001")

		db, err := NewSandboxOfSQLite3(m1)
		require.Error(t, err)
		require.Nil(t, db)
	})
}
