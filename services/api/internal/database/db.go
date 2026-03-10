package database

import (
	"time"

	"github.com/meanii/pipebin.dev/services/api/repository"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(dsn string) (*gorm.DB, error) {

	zap.S().Infof("connecting to postgres %s", dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	zap.S().Info("connected to postgres database")
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	// TODO: add auto migration
	db.AutoMigrate(
		repository.PasteModel{},
	)

	return db, nil
}
