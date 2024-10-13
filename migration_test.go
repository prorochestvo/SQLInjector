package sqlinjector

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/prorochestvo/sqlinjector/internal/sandbox"
	"github.com/prorochestvo/sqlinjector/internal/schema"
	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func TestNewMigrater(t *testing.T) {
	m1, _ := NewMemoryMigration("CREATE TABLE M1 (m1_id VARCHAR(250));", "DROP TABLE"+" M1;", "m0001")
	m2, _ := NewMemoryMigration("CREATE TABLE M2 (m1_id VARCHAR(250));", "DROP TABLE"+" M2;", "m0002")
	m3, _ := NewMemoryMigration("CREATE TABLE M3 (m1_id VARCHAR(250));", "DROP TABLE"+" M3;", "m0003")
	m4, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id VARCHAR(250));", "DROP TABLE"+" M4;", "m0004")
	m5, _ := NewMemoryMigration("CREATE TABLE M5 (m1_id VARCHAR(250));", "DROP TABLE"+" M5;", "m0005")
	m6, _ := NewMemoryMigration("CREATE TABLE M6 (m1_id VARCHAR(250));", "DROP TABLE"+" M6;", "m0006")
	m7, _ := NewMemoryMigration("CREATE TABLE M7 (m1_id VARCHAR(250));", "DROP TABLE"+" M7;", "m0007")

	mALL := MultipleMigration(m1, m2, m3, m4, m5, m6, m7)

	m := NewMigrater(mALL)
	require.NotNil(t, m)

	require.Equal(t, m.tableName, defaultTableName)

	require.NotNil(t, m.instructions)
	require.Len(t, m.instructions, 7)
	require.Equal(t, m1[0].ID(), m.instructions[0].ID())
	require.Equal(t, m2[0].ID(), m.instructions[1].ID())
	require.Equal(t, m3[0].ID(), m.instructions[2].ID())
	require.Equal(t, m4[0].ID(), m.instructions[3].ID())
	require.Equal(t, m5[0].ID(), m.instructions[4].ID())
	require.Equal(t, m6[0].ID(), m.instructions[5].ID())
	require.Equal(t, m7[0].ID(), m.instructions[6].ID())
}

func TestMigrater_State(t *testing.T) {
	pool := sandbox.NewPool()
	require.NotNil(t, pool)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(pool)

	m1, _ := NewMemoryMigration("CREATE TABLE M1 (m1_id VARCHAR(250));", "DROP TABLE"+" M1;", "m0001")
	m2, _ := NewMemoryMigration("CREATE TABLE M2 (m1_id VARCHAR(250));", "DROP TABLE"+" M2;", "m0002")
	m3, _ := NewMemoryMigration("CREATE TABLE M3 (m1_id VARCHAR(250));", "DROP TABLE"+" M3;", "m0003")
	m4, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id VARCHAR(250));", "DROP TABLE"+" M4;", "m0004")
	m5, _ := NewMemoryMigration("CREATE TABLE M5 (m1_id VARCHAR(250));", "DROP TABLE"+" M5;", "m0005")
	m6, _ := NewMemoryMigration("CREATE TABLE M6 (m1_id VARCHAR(250));", "DROP TABLE"+" M6;", "m0006")
	m7, _ := NewMemoryMigration("CREATE TABLE M7 (m1_id VARCHAR(250));", "DROP TABLE"+" M7;", "m0007")

	mALL := MultipleMigration(m1, m2, m3, m4, m5, m6, m7)

	t.Run("PostgreSQL", func(t *testing.T) {
		db, err := pool.NewPostgreSQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		mERR, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id TEXT);", "", "m0004")
		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, mERR, m5, m7)).Up(db))

		migrater := NewMigrater(mALL)
		require.NotNil(t, migrater)

		s, err := migrater.State(db)
		require.NoError(t, err)
		require.NotNil(t, s)

		expected := []string{
			fmt.Sprintf("%s %s", "[X]", mALL[0].ID()),
			fmt.Sprintf("%s %s", "[ ]", mALL[1].ID()),
			fmt.Sprintf("%s %s", "[X]", mALL[2].ID()),
			fmt.Sprintf("%s %s (%s)", "ERR", mERR[0].ID(), mERR[0].MD5()),
			fmt.Sprintf("%s %s", "[X]", mALL[4].ID()),
			fmt.Sprintf("%s %s", "[ ]", mALL[5].ID()),
			fmt.Sprintf("%s %s", "[X]", mALL[6].ID()),
		}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(s, "; "))
	})
	t.Run("MySQL", func(t *testing.T) {
		db, err := pool.NewMySQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		mERR, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id VARCHAR(200));", "", "m0004")
		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, mERR, m5, m7)).Up(db))

		migrater := NewMigrater(mALL)
		require.NotNil(t, migrater)

		s, err := migrater.State(db)
		require.NoError(t, err)
		require.NotNil(t, s)

		expected := []string{
			fmt.Sprintf("%s %s", "[X]", mALL[0].ID()),
			fmt.Sprintf("%s %s", "[ ]", mALL[1].ID()),
			fmt.Sprintf("%s %s", "[X]", mALL[2].ID()),
			fmt.Sprintf("%s %s (%s)", "ERR", mERR[0].ID(), mERR[0].MD5()),
			fmt.Sprintf("%s %s", "[X]", mALL[4].ID()),
			fmt.Sprintf("%s %s", "[ ]", mALL[5].ID()),
			fmt.Sprintf("%s %s", "[X]", mALL[6].ID()),
		}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(s, "; "))
	})
	t.Run("SQLite", func(t *testing.T) {
		db, err := poll.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		mERR, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id TEXT);", "", "m0004")
		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, mERR, m5, m7)).Up(db))

		migrater := NewMigrater(mALL)
		require.NotNil(t, migrater)

		s, err := migrater.State(db)
		require.NoError(t, err)
		require.NotNil(t, s)

		expected := []string{
			fmt.Sprintf("%s %s", "[X]", mALL[0].ID()),
			fmt.Sprintf("%s %s", "[ ]", mALL[1].ID()),
			fmt.Sprintf("%s %s", "[X]", mALL[2].ID()),
			fmt.Sprintf("%s %s (%s)", "ERR", mERR[0].ID(), mERR[0].MD5()),
			fmt.Sprintf("%s %s", "[X]", mALL[4].ID()),
			fmt.Sprintf("%s %s", "[ ]", mALL[5].ID()),
			fmt.Sprintf("%s %s", "[X]", mALL[6].ID()),
		}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(s, "; "))
	})
}

func TestMigrater_Plan(t *testing.T) {
	pool := sandbox.NewPool()
	require.NotNil(t, pool)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(pool)

	m1, _ := NewMemoryMigration("CREATE TABLE M1 (m1_id VARCHAR(250));", "DROP TABLE"+" M1;", "m0001")
	m2, _ := NewMemoryMigration("CREATE TABLE M2 (m1_id VARCHAR(250));", "DROP TABLE"+" M2;", "m0002")
	m3, _ := NewMemoryMigration("CREATE TABLE M3 (m1_id VARCHAR(250));", "DROP TABLE"+" M3;", "m0003")
	m4, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id VARCHAR(250));", "DROP TABLE"+" M4;", "m0004")
	m5, _ := NewMemoryMigration("CREATE TABLE M5 (m1_id VARCHAR(250));", "DROP TABLE"+" M5;", "m0005")
	m6, _ := NewMemoryMigration("CREATE TABLE M6 (m1_id VARCHAR(250));", "DROP TABLE"+" M6;", "m0006")
	m7, _ := NewMemoryMigration("CREATE TABLE M7 (m1_id VARCHAR(250));", "DROP TABLE"+" M7;", "m0007")

	mALL := MultipleMigration(m1, m2, m3, m4, m5, m6, m7)

	t.Run("PostgreSQL", func(t *testing.T) {
		db, err := poll.NewPostgreSQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		mERR, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id TEXT);", "", "m0004")
		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, mERR, m5, m7)).Up(db))

		migrater := NewMigrater(mALL)
		require.NotNil(t, migrater)

		s, err := migrater.Plan(db)
		require.NoError(t, err)
		require.NotNil(t, s)

		expected := []string{
			mALL[1].ID(),
			mALL[5].ID(),
		}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(s, "; "))
	})
	t.Run("MySQL", func(t *testing.T) {
		db, err := poll.NewMySQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		mERR, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id VARCHAR(200));", "", "m0004")
		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, mERR, m5, m7)).Up(db))

		migrater := NewMigrater(mALL)
		require.NotNil(t, migrater)

		s, err := migrater.Plan(db)
		require.NoError(t, err)
		require.NotNil(t, s)

		expected := []string{
			mALL[1].ID(),
			mALL[5].ID(),
		}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(s, "; "))
	})
	t.Run("SQLite", func(t *testing.T) {
		db, err := pool.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		mERR, _ := NewMemoryMigration("CREATE TABLE M4 (m1_id TEXT);", "", "m0004")
		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, mERR, m5, m7)).Up(db))

		migrater := NewMigrater(mALL)
		require.NotNil(t, migrater)

		s, err := migrater.Plan(db)
		require.NoError(t, err)
		require.NotNil(t, s)

		expected := []string{
			mALL[1].ID(),
			mALL[5].ID(),
		}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(s, "; "))
	})
}

func TestMigrater_Up(t *testing.T) {
	pool := sandbox.NewPool()
	require.NotNil(t, pool)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(pool)

	v1 := "V001"
	v2 := "V002"
	v3 := "V003"
	v4 := "V004"
	v5 := "V005"
	v6 := "V006"
	v7 := "V007"
	table := "m1"
	m0, _ := NewMemoryMigration("CREATE TABLE "+table+" (id VARCHAR(50));", "DROP TABLE"+" "+table+";", "m0000")
	m1, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v1+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v1+"';", "m0001")
	m2, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v2+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v2+"';", "m0002")
	m3, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v3+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v3+"';", "m0003")
	m4, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v4+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v4+"';", "m0004")
	m5, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v5+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v5+"';", "m0005")
	m6, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v6+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v6+"';", "m0006")
	m7, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v7+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v7+"';", "m0007")

	mALL := MultipleMigration(m1, m2, m3, m4, m5, m6, m7)

	extractActualIDs := func(t *testing.T, d internal.Dispatcher) []string {
		rows, err := d.Query("SELECT id FROM" + " " + table + " ORDER BY id;")
		require.NoError(t, err)
		require.NotNil(t, rows)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(rows)

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
		db, err := pool.NewPostgreSQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		require.NoError(t, NewMigrater(m0).Up(db))

		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, m5, m7)).Up(db))
		actually := extractActualIDs(t, db)
		expected := []string{v1, v3, v5, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(MultipleMigration(m2, m4, m6, m7)).Up(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Up(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("MySQL", func(t *testing.T) {
		db, err := pool.NewMySQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		require.NoError(t, NewMigrater(m0).Up(db))

		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, m5, m7)).Up(db))
		actually := extractActualIDs(t, db)
		expected := []string{v1, v3, v5, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(MultipleMigration(m2, m4, m6, m7)).Up(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Up(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("SQLite", func(t *testing.T) {
		db, err := pool.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		require.NoError(t, NewMigrater(m0).Up(db))

		require.NoError(t, NewMigrater(MultipleMigration(m1, m3, m5, m7)).Up(db))
		actually := extractActualIDs(t, db)
		expected := []string{v1, v3, v5, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(MultipleMigration(m2, m4, m6, m7)).Up(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Up(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
}

func TestMigrater_Down(t *testing.T) {
	pool := sandbox.NewPool()
	require.NotNil(t, pool)
	defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(pool)

	v1 := "V001"
	v2 := "V002"
	v3 := "V003"
	v4 := "V004"
	v5 := "V005"
	v6 := "V006"
	v7 := "V007"
	table := "m1"
	m0, _ := NewMemoryMigration("CREATE TABLE "+table+" (id VARCHAR(50));", "DROP TABLE"+" "+table+";", "m0000")
	m1, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v1+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v1+"';", "m0001")
	m2, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v2+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v2+"';", "m0002")
	m3, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v3+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v3+"';", "m0003")
	m4, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v4+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v4+"';", "m0004")
	m5, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v5+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v5+"';", "m0005")
	m6, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v6+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v6+"';", "m0006")
	m7, _ := NewMemoryMigration("INSERT"+" INTO "+table+" (id) VALUES ('"+v7+"');", "DELETE FROM"+" "+table+" WHERE id = '"+v7+"';", "m0007")

	mALL := MultipleMigration(m1, m2, m3, m4, m5, m6, m7)

	extractActualIDs := func(t *testing.T, d internal.Dispatcher) []string {
		rows, err := d.Query("SELECT id FROM" + " " + table + " ORDER BY id;")
		require.NoError(t, err)
		require.NotNil(t, rows)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(rows)

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
		db, err := pool.NewPostgreSQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		require.NoError(t, NewMigrater(MultipleMigration(m0, mALL)).Up(db))
		actually := extractActualIDs(t, db)
		expected := []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(MultipleMigration(m2, m4, m6, m7)).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("MySQL", func(t *testing.T) {
		db, err := pool.NewMySQL(20000)
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		require.NoError(t, NewMigrater(MultipleMigration(m0, mALL)).Up(db))
		actually := extractActualIDs(t, db)
		expected := []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(MultipleMigration(m2, m4, m6, m7)).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
	t.Run("SQLite", func(t *testing.T) {
		db, err := pool.NewSQLite3()
		require.NoError(t, err)
		require.NotNil(t, db)
		defer func(closer io.Closer) { require.NoError(t, closer.Close()) }(db)

		require.NoError(t, NewMigrater(MultipleMigration(m0, mALL)).Up(db))
		actually := extractActualIDs(t, db)
		expected := []string{v1, v2, v3, v4, v5, v6, v7}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5, v6}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4, v5}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3, v4}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2, v3}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1, v2}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{v1}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(mALL).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))

		require.NoError(t, NewMigrater(MultipleMigration(m2, m4, m6, m7)).Down(db))
		actually = extractActualIDs(t, db)
		expected = []string{}
		require.Equal(t, strings.Join(expected, "; "), strings.Join(actually, "; "))
	})
}

func TestNewFileMigration(t *testing.T) {
	folder := t.TempDir()
	tableName := "os_files_01"
	createTable := fmt.Sprintf("%sUp\nCREATE TABLE %s (id TEXT PRIMARY KEY);\n\n%sDown\nDROP TABLE %s;\n\n", schema.MigrationCommandPrefix, tableName, schema.MigrationCommandPrefix, tableName)
	insertTable01 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", schema.MigrationCommandPrefix, tableName, uuid.NewV4().String(), schema.MigrationCommandPrefix, tableName)
	insertTable02 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", schema.MigrationCommandPrefix, tableName, uuid.NewV4().String(), schema.MigrationCommandPrefix, tableName)
	insertTable03 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", schema.MigrationCommandPrefix, tableName, uuid.NewV4().String(), schema.MigrationCommandPrefix, tableName)
	require.NoError(t, os.WriteFile(path.Join(folder, "003.insert_table_02.sql"), []byte(insertTable02), 0666))
	require.NoError(t, os.WriteFile(path.Join(folder, "001.create_table.sql"), []byte(createTable), 0666))
	require.NoError(t, os.WriteFile(path.Join(folder, "001.create_table.txt"), []byte(createTable), 0666))
	require.NoError(t, os.WriteFile(path.Join(folder, "002.insert_table_01.sql"), []byte(insertTable01), 0666))
	require.NoError(t, os.WriteFile(path.Join(folder, "004.insert_table_03.sql"), []byte(insertTable03), 0666))

	m, err := NewFileMigration(folder)
	require.NoError(t, err)
	require.NotNil(t, m)

	require.Len(t, m, 4)
	for i, expected := range []string{createTable, insertTable01, insertTable02, insertTable03} {
		actually := fmt.Sprintf("%sUp\n%s\n\n%sDown\n%s\n\n", schema.MigrationCommandPrefix, m[i].(instructionUp).Up(), schema.MigrationCommandPrefix, m[i].(instructionDown).Down())
		require.Equal(t, expected, actually)
	}
}

func TestNewEmbedMigration(t *testing.T) {
	m, err := NewEmbedMigration(internalMigrations, "internal/schema")
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Len(t, m, 2)
	require.Equal(t, "instruction_test.01.sql", m[0].ID())
	require.Equal(t, "CREATE"+" TABLE os_files_01 (id TEXT PRIMARY KEY);", m[0].(instructionUp).Up())
	require.Equal(t, "DROP"+" TABLE os_files_01;", m[0].(instructionDown).Down())
	require.Equal(t, "instruction_test.02.sql", m[1].ID())
	require.Equal(t, "INSERT"+" INTO os_files_01 (id) VALUES ('bd5dc0fa-db1a-4e15-bea4-34c4fcc2133b');", m[1].(instructionUp).Up())
	require.Equal(t, "DELETE"+" FROM os_files_01;", m[1].(instructionDown).Down())
}

func TestNewMemoryMigration(t *testing.T) {
	up := "CREATE" + " TABLE demo (id INTEGER);"
	down := "DROP" + " TABLE demo;"
	m, err := NewMemoryMigration(up, down, "demo", "table")
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Len(t, m, 1)
	require.Equal(t, "demo_table", m[0].ID())
	require.Equal(t, up, m[0].(instructionUp).Up())
	require.Equal(t, down, m[0].(instructionDown).Down())
	require.Equal(t, makeMD5(up+"\n"+down), m[0].MD5())
}

func TestNewStructMigration(t *testing.T) {
	m, err := NewStructMigration(struct {
		ID int `boil:"id" `
	}{}, "demo", internal.DialectSQLite3)
	unix := time.Now().UTC().Unix()
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Len(t, m, 1)
	require.Equal(t, fmt.Sprintf("%d_demo_create", unix), m[0].ID())
	require.Equal(t, "CREATE"+" TABLE IF NOT EXISTS demo (\n   id INTEGER NOT NULL DEFAULT 0 PRIMARY KEY);", m[0].(instructionUp).Up())
	require.Equal(t, "DROP"+" TABLE IF EXISTS demo;", m[0].(instructionDown).Down())
	require.Equal(t, makeMD5(m[0].(instructionUp).Up()+"\n"+m[0].(instructionDown).Down()), m[0].MD5())
}

//go:embed internal/schema/*.sql
var internalMigrations embed.FS

func makeMD5(src string) string {
	hash := md5.New()
	hash.Write([]byte(src))
	h := hash.Sum(nil)
	return hex.EncodeToString(h)
}

type instructionDown interface {
	Down() string
}

type instructionUp interface {
	Up() string
}
