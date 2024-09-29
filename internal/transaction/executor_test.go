package transaction

import (
	"context"
	"database/sql"
	"encoding/hex"
	"github.com/prorochestvo/sqlinjector/internal/receptacle"
	"github.com/prorochestvo/sqlinjector/internal/schema"
	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"io"
	"strings"
	"testing"
)

func TestCommit(t *testing.T) {
	v1 := "V001"
	v2 := "V002"
	v3 := "V003"
	v4 := "V004"
	v5 := "V005"
	v6 := "V006"
	v7 := "V007"
	table := "testCommit1"
	initial := schema.NewInstruction("m0000", "CREATE TABLE "+table+" (id VARCHAR(50));", "DROP TABLE"+" "+table+";")

	insert := func(value string) Action {
		return func(executor boil.ContextExecutor) (interface{}, error) {
			sqlScript := "INSERT INTO" + " " + table + " (id) VALUES ('" + strings.ReplaceAll(value, "'", "") + "');"
			_, err := executor.Exec(sqlScript)
			return 0, err
		}
	}

	extractActualIDs := func(t *testing.T, db *sql.DB) []string {
		rows, err := db.Query("SELECT id FROM" + " " + table + " ORDER BY id;")
		require.NoError(t, err)
		require.NotNil(t, rows)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(rows)

		res := []string{}

		for rows.Next() {
			var id string
			require.NoError(t, rows.Scan(&id))
			require.NotEmpty(t, id)
			res = append(res, id)
		}

		return res
	}

	t.Run("PostgreSQL", func(t *testing.T) {
		container, err := receptacle.NewPostgreSQL(20011, "test", "test", "test")
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		_, err = Commit(context.Background(), container.DB, actions{insert(v1), insert(v3), insert(v5), insert(v7)})
		require.NoError(t, err)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v1, v3, v5, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		_, err = Commit(context.Background(), container.DB, actions{insert(v2), insert(v4), insert(v6)})
		require.NoError(t, err)
		actually = extractActualIDs(t, container.DB)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("MySQL", func(t *testing.T) {
		container, err := receptacle.NewMySQL(20012, "test", "test", "test")
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		_, err = Commit(context.Background(), container.DB, actions{insert(v1), insert(v3), insert(v5), insert(v7)})
		require.NoError(t, err)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v1, v3, v5, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		_, err = Commit(context.Background(), container.DB, actions{insert(v2), insert(v4), insert(v6)})
		require.NoError(t, err)
		actually = extractActualIDs(t, container.DB)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("SQLite", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		_, err = Commit(context.Background(), container.DB, actions{insert(v1), insert(v3), insert(v5), insert(v7)})
		require.NoError(t, err)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v1, v3, v5, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		_, err = Commit(context.Background(), container.DB, actions{insert(v2), insert(v4), insert(v6)})
		require.NoError(t, err)
		actually = extractActualIDs(t, container.DB)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("Error", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		insertError := func(msg string) Action {
			return func(executor boil.ContextExecutor) (interface{}, error) {
				sqlScript := "INSERT INTO" + " " + table + " (id) VALUES (" + msg + ",'V');"
				_, sqlError := executor.Exec(sqlScript)
				return 0, sqlError
			}
		}
		errorMSG := hex.EncodeToString(uuid.NewV4().Bytes())

		_, err = Commit(context.Background(), container.DB, actions{insertError(errorMSG)})
		require.Error(t, err)
		require.Contains(t, err.Error(), errorMSG)
	})
	t.Run("Panic", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		insertPanic := func(msg string) Action {
			return func(_ boil.ContextExecutor) (interface{}, error) { panic(msg) }
		}
		panicMSG := hex.EncodeToString(uuid.NewV4().Bytes())

		_, err = Commit(context.Background(), container.DB, actions{insertPanic(panicMSG)})
		require.Error(t, err)
		require.Contains(t, err.Error(), panicMSG)
	})
	t.Run("RollbackAfterError", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		insertError := func(msg string) Action {
			return func(executor boil.ContextExecutor) (interface{}, error) {
				sqlScript := "INSERT INTO" + " " + table + " (id) VALUES (" + msg + ",'V');"
				_, sqlError := executor.Exec(sqlScript)
				return 0, sqlError
			}
		}
		errorMSG := hex.EncodeToString(uuid.NewV4().Bytes())

		_, err = Commit(context.Background(), container.DB, actions{insert(v7)})
		require.NoError(t, err)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		_, err = Commit(context.Background(), container.DB, actions{insert(v1), insert(v2), insertError(errorMSG), insert(v4), insert(v5)})
		require.Error(t, err)
		require.Contains(t, err.Error(), errorMSG)
		actually = extractActualIDs(t, container.DB)
		expected = []string{v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		_, err = Commit(context.Background(), container.DB, actions{insert(v6)})
		require.NoError(t, err)
		actually = extractActualIDs(t, container.DB)
		expected = []string{v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("BringDataset", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		retrieve := func(executor boil.ContextExecutor) (res interface{}, err error) {
			rows, err := executor.Query("SELECT id FROM" + " " + table + " ORDER BY id;")
			if err != nil {
				return
			}
			defer func(c io.Closer) { require.NoError(t, c.Close()) }(rows)

			ids := make([]string, 0)

			for rows.Next() {
				var id string
				err = rows.Scan(&id)
				if err != nil {
					return
				}
				ids = append(ids, id)
			}

			res = ids

			return
		}

		_, err = Commit(context.Background(), container.DB, actions{insert(v3), insert(v6), insert(v7)})
		require.NoError(t, err)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v3, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		rows, err := Commit(context.Background(), container.DB, actions{insert(v1), insert(v2), retrieve, insert(v4), insert(v5)})
		require.NoError(t, err)
		require.NotNil(t, rows)
		require.Len(t, rows, 5)
		require.Equal(t, 0, rows.([]interface{})[0])
		require.Equal(t, 0, rows.([]interface{})[1])
		require.Equal(t, []string{v1, v2, v3, v6, v7}, rows.([]interface{})[2])
		require.Equal(t, 0, rows.([]interface{})[3])
		require.Equal(t, 0, rows.([]interface{})[4])
		actually = extractActualIDs(t, container.DB)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
}

func TestRollback(t *testing.T) {
	v1 := "V001"
	v2 := "V002"
	v3 := "V003"
	table := "testRollback1"
	initial := schema.NewInstruction("m0000", "CREATE TABLE "+table+" (id VARCHAR(50));", "DROP TABLE"+" "+table+";")
	insert1 := schema.NewInstruction("m0001", "INSERT"+" INTO "+table+" (id) VALUES ('"+v1+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v1+"';")
	insert2 := schema.NewInstruction("m0002", "INSERT"+" INTO "+table+" (id) VALUES ('"+v2+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v2+"';")
	insert3 := schema.NewInstruction("m0003", "INSERT"+" INTO "+table+" (id) VALUES ('"+v3+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v3+"';")

	insert := func(value string) Action {
		return func(executor boil.ContextExecutor) (interface{}, error) {
			sqlScript := "INSERT INTO" + " " + table + " (id) VALUES ('" + strings.ReplaceAll(value, "'", "") + "');"
			_, err := executor.Exec(sqlScript)
			return 0, err
		}
	}
	insertNew := insert(hex.EncodeToString(uuid.NewV4().Bytes()))

	extractActualIDs := func(t *testing.T, db *sql.DB) []string {
		rows, err := db.Query("SELECT id FROM" + " " + table + " ORDER BY id;")
		require.NoError(t, err)
		require.NotNil(t, rows)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(rows)

		res := []string{}

		for rows.Next() {
			var id string
			require.NoError(t, rows.Scan(&id))
			require.NotEmpty(t, id)
			res = append(res, id)
		}

		return res
	}

	t.Run("PostgreSQL", func(t *testing.T) {
		container, err := receptacle.NewPostgreSQL(20021, "test", "test", "test")
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert1, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert2, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert3, container.DB, migrationTable))

		_, err = Rollback(context.Background(), container.DB, actions{insertNew, insertNew, insertNew})
		require.NoError(t, err)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v1, v2, v3}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("MySQL", func(t *testing.T) {
		container, err := receptacle.NewMySQL(20022, "test", "test", "test")
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert1, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert2, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert3, container.DB, migrationTable))

		_, err = Rollback(context.Background(), container.DB, actions{insertNew, insertNew, insertNew})
		require.NoError(t, err)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v1, v2, v3}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("SQLite", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert1, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert2, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert3, container.DB, migrationTable))

		_, err = Rollback(context.Background(), container.DB, actions{insertNew, insertNew, insertNew})
		require.NoError(t, err)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v1, v2, v3}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("Error", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		insertError := func(msg string) Action {
			return func(executor boil.ContextExecutor) (interface{}, error) {
				sqlScript := "INSERT INTO" + " " + table + " (id) VALUES (" + msg + ",'V');"
				_, sqlError := executor.Exec(sqlScript)
				return 0, sqlError
			}
		}
		errorMSG := hex.EncodeToString(uuid.NewV4().Bytes())

		_, err = Rollback(context.Background(), container.DB, actions{insertError(errorMSG)})
		require.Error(t, err)
		require.Contains(t, err.Error(), errorMSG)
	})
	t.Run("Panic", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))

		insertPanic := func(msg string) Action {
			return func(_ boil.ContextExecutor) (interface{}, error) { panic(msg) }
		}
		panicMSG := hex.EncodeToString(uuid.NewV4().Bytes())

		_, err = Rollback(context.Background(), container.DB, actions{insertPanic(panicMSG)})
		require.Error(t, err)
		require.Contains(t, err.Error(), panicMSG)
	})
	t.Run("RollbackAfterError", func(t *testing.T) {
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert1, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert2, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert3, container.DB, migrationTable))

		insertError := func(msg string) Action {
			return func(executor boil.ContextExecutor) (interface{}, error) {
				sqlScript := "INSERT INTO" + " " + table + " (id) VALUES (" + msg + ",'V');"
				_, sqlError := executor.Exec(sqlScript)
				return 0, sqlError
			}
		}
		errorMSG := hex.EncodeToString(uuid.NewV4().Bytes())

		_, err = Rollback(context.Background(), container.DB, actions{insertNew, insertError(errorMSG), insertNew})
		require.Error(t, err)
		require.Contains(t, err.Error(), errorMSG)
		actually := extractActualIDs(t, container.DB)
		expected := []string{v1, v2, v3}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("BringDataset", func(t *testing.T) {
		v4 := "V004"
		v5 := "V005"
		container, err := receptacle.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, container)
		defer func(c io.Closer) { require.NoError(t, c.Close()) }(container)

		require.NoError(t, schema.Up(initial, container.DB, migrationTable))
		require.NoError(t, schema.Up(insert3, container.DB, migrationTable))

		retrieve := func(executor boil.ContextExecutor) (res interface{}, err error) {
			rows, err := executor.Query("SELECT id FROM" + " " + table + " ORDER BY id;")
			if err != nil {
				return
			}
			defer func(c io.Closer) { require.NoError(t, c.Close()) }(rows)

			ids := make([]string, 0)

			for rows.Next() {
				var id string
				err = rows.Scan(&id)
				if err != nil {
					return
				}
				ids = append(ids, id)
			}

			res = ids

			return
		}

		rows, err := Rollback(context.Background(), container.DB, actions{insert(v1), insert(v2), retrieve, insert(v4), insert(v5)})
		require.NoError(t, err)
		require.NotNil(t, rows)
		require.Len(t, rows, 5)
		require.Equal(t, 0, rows.([]interface{})[0])
		require.Equal(t, 0, rows.([]interface{})[1])
		require.Equal(t, []string{v1, v2, v3}, rows.([]interface{})[2])
		require.Equal(t, 0, rows.([]interface{})[3])
		require.Equal(t, 0, rows.([]interface{})[4])
		actually := extractActualIDs(t, container.DB)
		expected := []string{v3}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
}

const migrationTable = "__migrations"
