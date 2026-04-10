package database

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meanii/pipebin.dev/services/api/migrations"
)

func New(dsn string) (*pgxpool.Pool, error) {
	slog.Info("connecting to postgres")

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MinConns = 10
	cfg.MaxConns = 100
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	slog.Info("connected to postgres database")

	if err = runMigrations(dsn); err != nil {
		return nil, err
	}

	return pool, nil
}

func runMigrations(dsn string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, toPgx5DSN(dsn))
	if err != nil {
		return err
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		slog.Debug("migrations: no changes")
	} else {
		slog.Info("migrations: applied successfully")
	}
	return nil
}

// toPgx5DSN converts a standard postgres:// DSN to the pgx5:// scheme
// required by the golang-migrate pgx/v5 database driver.
func toPgx5DSN(dsn string) string {
	if len(dsn) > 11 && dsn[:11] == "postgresql:" {
		return "pgx5:" + dsn[11:]
	}
	if len(dsn) > 9 && dsn[:9] == "postgres:" {
		return "pgx5:" + dsn[9:]
	}
	return dsn
}
