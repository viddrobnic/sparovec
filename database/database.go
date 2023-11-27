package database

import (
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Driver for migrating with files
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // Driver for sqlite3
	"github.com/viddrobnic/sparovec/config"
)

func New(conf *config.Config) (*sqlx.DB, error) {
	return sqlx.Open("sqlite3", conf.Database.Location)
}

func Migrate(db *sqlx.DB, migrations fs.FS) error {
	migrationsFs, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration fs instance: %w", err)
	}

	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migration, err := migrate.NewWithInstance("iofs", migrationsFs, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
