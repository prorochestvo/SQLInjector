package sqlinjector

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/prorochestvo/sqlinjector/internal/schema"
	"hash/crc64"
	"net/http"
	"sort"
	"strings"
	"time"
)

func NewMigrater(m Migration) *Migrater {
	instructions := make(schema.Instruction, len(m))
	for i, item := range m {
		instructions[i] = item
	}

	return &Migrater{
		instructions: instructions,
		tableName:    defaultTableName,
	}
}

type Migrater struct {
	instructions schema.Instruction
	tableName    string
}

func (m *Migrater) SetTableName(n string) {
	m.tableName = n
}

func (m *Migrater) State(db *sql.DB) ([]string, error) {
	if len(m.instructions) == 0 {
		return nil, nil
	}

	type INSTRUCTION struct {
		ID    string
		MD5   string
		State string
	}

	exists, nonExists, undefined, err := schema.State(m.instructions, db, m.tableName)
	if err != nil {
		return nil, err
	}

	instructions := make([]INSTRUCTION, 0, len(exists)+len(nonExists)+len(undefined))

	for _, item := range exists {
		instruction := INSTRUCTION{
			ID:    item.ID(),
			MD5:   "",
			State: "[X]",
		}
		instructions = append(instructions, instruction)
	}

	for _, item := range nonExists {
		instruction := INSTRUCTION{
			ID:    item.ID(),
			MD5:   "",
			State: "[ ]",
		}
		instructions = append(instructions, instruction)
	}

	for _, item := range undefined {
		instruction := INSTRUCTION{
			ID:    item.ID(),
			MD5:   item.MD5(),
			State: "ERR",
		}
		instructions = append(instructions, instruction)
	}

	sort.SliceStable(instructions, func(i, j int) bool {
		return instructions[i].ID < instructions[j].ID
	})

	items := make([]string, len(instructions))
	for i, instruction := range instructions {
		if instruction.MD5 != "" {
			items[i] = fmt.Sprintf("%s %s (%s)", instruction.State, instruction.ID, instruction.MD5)
		} else {
			items[i] = fmt.Sprintf("%s %s", instruction.State, instruction.ID)
		}
	}

	return items, nil
}

func (m *Migrater) Plan(db *sql.DB) ([]string, error) {
	if len(m.instructions) == 0 {
		return nil, nil
	}

	instructions, err := schema.Plan(m.instructions, db, m.tableName)
	if err != nil {
		return nil, err
	}

	items := make([]string, len(instructions))
	for i, instruction := range instructions {
		items[i] = instruction.ID()
	}

	return items, nil
}

func (m *Migrater) Up(db *sql.DB) error {
	if len(m.instructions) == 0 {
		return nil
	}

	return schema.Up(m.instructions, db, m.tableName)
}

func (m *Migrater) Down(db *sql.DB) error {
	if len(m.instructions) == 0 {
		return nil
	}

	exists, _, _, err := schema.State(m.instructions, db, m.tableName)
	if err != nil {
		return err
	}

	var lastInstruction schema.Instruction = nil
	for i := len(m.instructions) - 1; i >= 0; i-- {
		for _, existInstruction := range exists {
			if existInstruction.ID() == m.instructions[i].ID() {
				lastInstruction = m.instructions[i:]
				i = 0
				break
			}
		}
	}

	if lastInstruction == nil {
		return nil
	}

	return schema.Down(lastInstruction, db, m.tableName)
}

func (m *Migrater) Clean(db *sql.DB) error {
	if len(m.instructions) == 0 {
		return nil
	}

	return schema.Down(m.instructions, db, m.tableName)
}

// NewFileMigration creates a new migration from a local folder
func NewFileMigration(folder string) (Migration, error) {
	filesystem := http.Dir(folder)
	items, err := schema.ExtractInstructions(filesystem, "/")
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
	items, err := schema.ExtractInstructions(filesystem, root)
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
		checksum := crc64.Checksum([]byte(up+"\n"+down), crcTable)
		mID = fmt.Sprintf("%d_%X", time.Now().UTC().Unix(), checksum)
	}
	items := schema.NewInstruction(
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
func NewStructMigration(boil interface{}, table string, dialect internal.Dialect) (Migration, error) {
	items, err := schema.MakeTableInstruction(table, boil, dialect)
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
	res := make(Migration, 0, len(migrations))
	for _, m := range migrations {
		res = append(res, m...)
	}
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].ID() < res[j].ID()
	})
	return res
}

type Migration schema.Instruction

const defaultTableName = "_migrations"

var crcTable = crc64.MakeTable(crc64.ECMA)
