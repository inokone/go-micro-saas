package db

import (
	"database/sql"
	"embed"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/inokone/go-micro-saas/internal/common"
	log "github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var fs embed.FS

// InitDB sets up the database connection pool
func InitDB(conf *common.RDBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", conf.String())
	if err != nil {
		log.WithError(err).Error("Failed to connect to database.")
		return nil, err
	}
	return db, nil
}

// ForceMigration applies the database migration based on the update scripts in migrations folder
// If migration fails it resolves the dirty flag in the database before shuttin down the application
func ForceMigration(conf *common.RDBConfig) {
	log.Info("Db migration started.")
	db, err := sql.Open("postgres", conf.String())
	if err != nil {
		os.Exit(1)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.WithError(err).Error("Failed to create database driver.")
		os.Exit(1)
	}

	migrations, err := iofs.New(fs, "migrations") // List migrations from db/migrations folder
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance("iofs", migrations, conf.DBName, driver)
	if err != nil {
		log.WithError(err).Error("Failed to create database migration engine.")
		os.Exit(1)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.WithError(err).Error("Failed to collect database version.")
		os.Exit(1)
	}

	if dirty {
		log.WithField("version", version).Warn("Failed migration detected, fixing version.")
		err = m.Force(int(version) - 1)
		if err != nil {
			log.WithError(err).Error("Failed to execute database migrations.")
			os.Exit(1)
		}
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.WithError(err).Error("Failed to execute database migrations.")
		os.Exit(1)
	}
	log.Info("Db migration finished successfully.")
}
