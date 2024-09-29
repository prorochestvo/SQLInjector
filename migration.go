package sqlinjector

import (
	"crypto/md5"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal/migration"
	"net/http"
	"sort"
	"strings"
	"time"
)

func NewMigrater(m Migration, db *sql.DB) *Migrater {
	return nil
}

type Migrater struct {
	instructions migration.Instruction
	db           *sql.DB
	tableName    string
}

func (m *Migrater) State() ([]string, error) {
	_, _, _, err := migration.State(m.instructions, m.db, m.tableName)
	return nil, err
}

func (m *Migrater) Plan() ([]string, error) {
	return nil, nil
}

func (m *Migrater) Up() error {
	return nil
}

func (m *Migrater) Down() error {
	return nil
}

func (m *Migrater) Clean() error {
	return nil
}

// NewFileMigration creates a new migration from a local folder
func NewFileMigration(folder string) (Migration, error) {
	filesystem := http.Dir(folder)
	items, err := migration.ExtractInstructions(filesystem, "/")
	if err != nil {
		return nil, err
	}
	m := make(Migration, len(items))
	for i, item := range items {
		m[i] = item
	}
	return m, nil
}

// NewEmbedMigration creates a new migration from an embed.FS
func NewEmbedMigration(fs embed.FS, root string) (Migration, error) {
	filesystem := http.FS(fs)
	items, err := migration.ExtractInstructions(filesystem, root)
	if err != nil {
		return nil, err
	}
	m := make(Migration, len(items))
	for i, item := range items {
		m[i] = item
	}
	return m, nil
}

// NewMemoryMigration creates a new migration from a memory value
func NewMemoryMigration(up, down string, id ...string) (Migration, error) {
	mID := strings.Join(id, "_")
	if mID == "" {
		hash := md5.New()
		hash.Write([]byte(up + down))
		h := hash.Sum(nil)
		mID = fmt.Sprintf("%d_%s", time.Now().UTC().Unix(), hex.EncodeToString(h))
	}
	items := migration.NewInstruction(
		mID,
		up,
		down,
	)
	m := make(Migration, len(items))
	for i, item := range items {
		m[i] = item
	}
	return m, nil
}

// NewStructMigration creates a new migration for SQLite3 database from a struct
// The struct must be a pointer to a sqlboiler struct.
func NewStructMigration(boil interface{}, table string) (Migration, error) {
	items, err := migration.MakeTableInstructionForSQLite3(table, boil)
	if err != nil || items == nil {
		if err == nil {
			err = fmt.Errorf("error while creating table instruction for %s", table)
		}
		return nil, err
	}
	m := make(Migration, len(items))
	for i, item := range items {
		m[i] = item
	}
	return m, nil
}

// MultipleMigration combines multiple migrations into one
func MultipleMigration(migrations ...Migration) Migration {
	res := make(Migration, len(migrations))
	for _, m := range migrations {
		res = append(res, m...)
	}
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].ID() < res[j].ID()
	})
	return res
}

type Migration migration.Instruction
