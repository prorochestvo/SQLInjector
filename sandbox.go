package sqlinjector

import (
	"fmt"
	"github.com/prorochestvo/sqlinjector/internal"
	"github.com/prorochestvo/sqlinjector/internal/sandbox"
	"math/rand"
)

// NewSandboxOfPostgreSQL create new connection PostgreSQL via docker container.
// technical function for testing purposes of external packages.
// closing the container is necessary to free the docker resources.
func NewSandboxOfPostgreSQL(port int, migrations ...Migration) (internal.Vault, error) {
	db, err := poll.NewPostgreSQL(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgresql container: %w", err)
	}

	var vaultID int64
	if vID, ok := db.(interface{ VaultID() int64 }); ok {
		vaultID = vID.VaultID()
	} else {
		vaultID = rand.Int63()
	}

	// make schema
	schemaID := fmt.Sprintf("schema_%0.3d", vaultID)
	schemaSQL := fmt.Sprintf("CREATE SCHEMA %s; SET search_path TO %s;", schemaID, schemaID)
	if _, err = db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("failed to switch schema: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// NewSandboxOfMySQL create new connection MySQL via docker container.
// technical function for testing purposes of external packages.
// closing the container is necessary to free the docker resources.
func NewSandboxOfMySQL(port int, migrations ...Migration) (internal.Vault, error) {
	db, err := poll.NewMySQL(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgresql container: %w", err)
	}

	var vaultID int64
	if vID, ok := db.(interface{ VaultID() int64 }); ok {
		vaultID = vID.VaultID()
	} else {
		vaultID = rand.Int63()
	}

	// make database
	schemaID := fmt.Sprintf("schema_%0.3d", vaultID)
	if _, err = db.Exec(fmt.Sprintf("CREATE DATABASE"+" %s;", schemaID)); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	if _, err = db.Exec(fmt.Sprintf("USE %s;", schemaID)); err != nil {
		return nil, fmt.Errorf("failed to switch database: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// NewSandboxOfSQLite3 create new connection SQLite into memory.
// technical function for testing purposes of external packages
// closing the container is necessary to free the docker resources.
func NewSandboxOfSQLite3(migrations ...Migration) (internal.Vault, error) {
	db, err := poll.NewSQLite3()
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite container: %w", err)
	}

	// migrate database
	err = NewMigrater(MultipleMigration(migrations...)).Up(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

var poll = sandbox.NewPool()
