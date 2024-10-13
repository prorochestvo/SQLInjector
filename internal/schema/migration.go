package schema

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

func State(m Instruction, c sqlConnection, tableName string) (exists Instruction, nonExists Instruction, undefined Instruction, err error) {
	t, err := c.Begin()
	if err != nil {
		return
	}
	defer func(t transaction) {
		err = errors.Join(err, t.Rollback())
	}(t)

	err = createMigrationTable(t, tableName)
	if err != nil {
		return
	}

	notExistsMigration, err := selectMigrationTable(t, tableName)
	if err != nil {
		return
	}

	exists = make(Instruction, 0, len(m))
	for _, i := range m {
		md5, ok := notExistsMigration[i.ID()]
		if ok {
			if md5 == i.MD5() {
				delete(notExistsMigration, i.ID())
				exists = append(exists, i)
			}
		} else {
			nonExists = append(nonExists, i)
		}
	}

	undefined = make(Instruction, 0, len(notExistsMigration))
	for id, md5 := range notExistsMigration {
		undefined = append(undefined, &migration{id: id, md5: md5})
	}

	sort.Slice(exists, func(i, j int) bool {
		return exists[i].ID() < exists[j].ID()
	})
	sort.Slice(nonExists, func(i, j int) bool {
		return nonExists[i].ID() < nonExists[j].ID()
	})
	sort.Slice(undefined, func(i, j int) bool {
		return undefined[i].ID() < undefined[j].ID()
	})

	return
}

func Plan(m Instruction, c sqlConnection, tableName string) (items Instruction, err error) {
	t, err := c.Begin()
	if err != nil {
		return
	}
	defer func(t transaction) {
		err = errors.Join(err, t.Rollback())
	}(t)

	err = createMigrationTable(t, tableName)
	if err != nil {
		return
	}

	notExistsMigration, err := selectMigrationTable(t, tableName)
	if err != nil {
		return
	}

	items = make(Instruction, 0, len(m))
	for _, i := range m {
		_, exists := notExistsMigration[i.ID()]
		if exists {
			delete(notExistsMigration, i.ID())
			continue
		}
		items = append(items, i)
	}

	if len(notExistsMigration) != 0 {
		migrationIDs := make([]string, 0, len(notExistsMigration))
		for id := range notExistsMigration {
			migrationIDs = append(migrationIDs, id)
		}
		err = fmt.Errorf("unexpected migrations in database: %v", strings.Join(migrationIDs, ", "))
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID() < items[j].ID()
	})

	return
}

func Up(m Instruction, c sqlConnection, tableName string) (err error) {
	t, err := c.Begin()
	if err != nil {
		return err
	}
	defer func(t transaction) {
		if err != nil {
			err = errors.Join(err, t.Rollback())
		} else {
			err = t.Commit()
		}
	}(t)

	err = createMigrationTable(t, tableName)
	if err != nil {
		return err
	}

	exists, err := selectMigrationTable(t, tableName)
	if err != nil {
		return
	}

	type sqlInstructionUP interface {
		Up() string
	}

	sort.Slice(m, func(i, j int) bool {
		return m[i].ID() < m[j].ID()
	})

	for _, i := range m {
		if _, ok := exists[i.ID()]; ok {
			continue
		}

		var sqlScript string
		if s, ok := i.(sqlInstructionUP); ok && s != nil {
			sqlScript = s.Up()
		}

		println("<<< sqlScript:", sqlScript)

		err = insertMigrationTable(t, tableName, i.ID(), i.MD5())
		if err != nil {
			err = fmt.Errorf("failed to keep migration %s hash, reason: %w", i.ID(), err)
			return
		}

		if sqlScript == "" {
			continue
		}

		_, err = t.Exec(sqlScript)
		if err != nil {
			err = fmt.Errorf("failed to execute up migration %s, reason: %w", i.ID(), err)
			return
		}
	}

	return
}

func Down(m Instruction, c sqlConnection, tableName string) (err error) {
	t, err := c.Begin()
	if err != nil {
		return err
	}
	defer func(t transaction) {
		if err != nil {
			err = errors.Join(err, t.Rollback())
		} else {
			err = t.Commit()
		}
	}(t)

	err = createMigrationTable(t, tableName)
	if err != nil {
		return err
	}

	exists, err := selectMigrationTable(t, tableName)
	if err != nil {
		return
	}

	type sqlInstructionDOWN interface {
		Down() string
	}

	sort.Slice(m, func(i, j int) bool {
		return m[i].ID() > m[j].ID()
	})

	for _, i := range m {
		if _, ok := exists[i.ID()]; !ok {
			continue
		}

		var sqlScript string
		if s, ok := i.(sqlInstructionDOWN); ok && s != nil {
			sqlScript = s.Down()
		}

		if err = deleteMigrationTable(t, tableName, i.ID(), i.MD5()); err != nil {
			err = fmt.Errorf("failed to keep migration %s hash, reason: %w", i.ID(), err)
			return
		}

		if sqlScript == "" {
			continue
		}

		_, err = t.Exec(sqlScript)
		if _, err = t.Exec(sqlScript); err != nil {
			err = fmt.Errorf("failed to execute up migration %s, reason: %w", i.ID(), err)
			return
		}
	}

	return
}

func createMigrationTable(e executor, table string) error {
	_, err := e.Exec("CREATE TABLE IF NOT EXISTS " + table + " (id VARCHAR(50) NOT NULL PRIMARY KEY, md5 VARCHAR(50) NOT NULL, applied_at VARCHAR(50) NOT NULL);")
	return err
}

func insertMigrationTable(e executor, table string, id, md5 string) error {
	sqlScript := fmt.Sprintf(
		"INSERT"+" INTO "+table+" (id, md5, applied_at) VALUES ('%s', '%s', '%s');",
		strings.ReplaceAll(id, "'", ""),
		strings.ReplaceAll(md5, "'", ""),
		time.Now().UTC().Format(time.RFC3339),
	)
	r, err := e.Exec(sqlScript)
	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("unexpected number of rows affected: %d", rows)
	}
	return err
}

func deleteMigrationTable(e executor, table string, id, md5 string) error {
	sqlScript := fmt.Sprintf(
		"DELETE"+" FROM "+table+" WHERE id = '%s' AND md5 = '%s';",
		strings.ReplaceAll(id, "'", ""),
		strings.ReplaceAll(md5, "'", ""),
	)
	r, err := e.Exec(sqlScript)
	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("unexpected number of rows affected: %d", rows)
	}
	return err
}

func selectMigrationTable(e executor, table string) (m map[string]string, err error) {
	rows, err := e.Query("SELECT id, md5" + " FROM " + table + " ORDER BY applied_at;")
	if err != nil {
		return
	}

	defer func(c io.Closer) { err = errors.Join(err, c.Close()) }(rows)

	m = make(map[string]string, 0)
	for rows.Next() {
		var id, md5 string
		if err = rows.Scan(&id, &md5); err != nil {
			return
		}
		m[id] = md5
	}

	return
}

type sqlConnection interface {
	Begin() (*sql.Tx, error)
}

type executor interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
}

type transaction interface {
	Commit() error
	Rollback() error
}
