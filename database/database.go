package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Driver for migrating with files
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite3
	"github.com/viddrobnic/sparovec/config"
)

const defaultMigrationsPath = "migrations"

func New(conf *config.Config) (*sqlx.DB, error) {
	return sqlx.Open("sqlite3", conf.Database.Location)
}

func MigrateWithPath(db *sqlx.DB, path string) error {
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance("file://"+path, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

func Migrate(db *sqlx.DB) error {
	return MigrateWithPath(db, defaultMigrationsPath)
}
