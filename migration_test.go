package sqlinjector

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal/migration"
	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
	"os"
	"path"
	"testing"
	"time"
)

func TestNewFileMigration(t *testing.T) {
	folder := t.TempDir()
	tableName := "os_files_01"
	createTable := fmt.Sprintf("%sUp\nCREATE TABLE %s (id TEXT PRIMARY KEY);\n\n%sDown\nDROP TABLE %s;\n\n", migration.MigrationCommandPrefix, tableName, migration.MigrationCommandPrefix, tableName)
	insertTable01 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", migration.MigrationCommandPrefix, tableName, uuid.NewV4().String(), migration.MigrationCommandPrefix, tableName)
	insertTable02 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", migration.MigrationCommandPrefix, tableName, uuid.NewV4().String(), migration.MigrationCommandPrefix, tableName)
	insertTable03 := fmt.Sprintf("%sUp\nINSERT INTO %s (id) VALUES ('%s');\n\n%sDown\nDELETE FROM %s;\n\n", migration.MigrationCommandPrefix, tableName, uuid.NewV4().String(), migration.MigrationCommandPrefix, tableName)
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
		actually := fmt.Sprintf("%sUp\n%s\n\n%sDown\n%s\n\n", migration.MigrationCommandPrefix, m[i].(instructionUp).Up(), migration.MigrationCommandPrefix, m[i].(instructionDown).Down())
		require.Equal(t, expected, actually)
	}
}

func TestNewEmbedMigration(t *testing.T) {
	m, err := NewEmbedMigration(internalMigrations, "internal/migration")
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
	}{}, "demo")
	unix := time.Now().UTC().Unix()
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Len(t, m, 1)
	require.Equal(t, fmt.Sprintf("%d_demo_create", unix), m[0].ID())
	require.Equal(t, "CREATE"+" TABLE IF NOT EXISTS demo (\n   id INTEGER NOT NULL PRIMARY KEY);", m[0].(instructionUp).Up())
	require.Equal(t, "DROP"+" TABLE IF EXISTS demo;", m[0].(instructionDown).Down())
	require.Equal(t, makeMD5(m[0].(instructionUp).Up()+"\n"+m[0].(instructionDown).Down()), m[0].MD5())
}

//go:embed internal/migration/*.sql
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
