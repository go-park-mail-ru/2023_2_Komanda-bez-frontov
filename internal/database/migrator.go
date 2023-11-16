package database

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"go-form-hub/internal/config"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type migration struct {
	SQL     string
	Version int
}

type version struct {
	Version int  `db:"version"`
	Dirty   bool `db:"dirty"`
}

func ConnectDatabaseWithRetry(
	cfg *config.Config,
) (ConnPool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	uri, err := ParseURI(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse uri error: %e", err)
	}

	poolConfig, err := pgxpool.ParseConfig(uri.String())

	schema := uri.Query().Get("search_path")

	for i := 1; i <= cfg.DatabaseConnectMaxRetries; i++ {
		db, connectErr := pgxpool.NewWithConfig(context.Background(), poolConfig)
		if connectErr != nil {
			log.Info().Msgf("trying to connect database with retry, retries %d, error %e", i, connectErr)
			err = fmt.Errorf("connect to database failed after %d retries: %e", i, connectErr)
			time.Sleep(cfg.DatabaseConnectRetryTimeout)
			continue
		}

		return NewConnPool(db, schema), nil
	}

	return nil, err
}

func ParseURI(uri string) (*url.URL, error) {
	dbURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("db connect failed: parse uri error %e", err)
	}

	dbQueryParams := dbURL.Query()
	schema := dbQueryParams.Get("search_path")
	if schema == "" {
		schema = "nofronts"
		dbQueryParams.Set("search_path", schema)
		dbURL.RawQuery = dbQueryParams.Encode()
	}
	return dbURL, nil
}

func Migrate(db ConnPool, cfg *config.Config, builder squirrel.StatementBuilderType) (int, error) {
	if cfg == nil {
		return 0, fmt.Errorf("config is nil")
	}

	ctx := context.Background()
	uri, err := ParseURI(cfg.DatabaseURL)
	if err != nil {
		return 0, fmt.Errorf("parse schema error: %e", err)
	}

	schema := uri.Query().Get("search_path")

	if schema != "public" {
		tx, err := db.Begin(ctx)
		if err != nil {
			return 0, fmt.Errorf("create schema begin transaction error: %e", err)
		}

		if _, err = tx.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s;", schema)); err != nil {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("create schema %s error: %e", schema, err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			return 0, fmt.Errorf("user_repository find_by_username failed to commit transaction: %e", err)
		}
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.schema_migrations (
		version BIGINT NOT NULL PRIMARY KEY,
		dirty BOOLEAN NOT NULL DEFAULT false
	)`, schema)

	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("create schema_migrations table begin transaction error: %e", err)
	}
	if _, err = tx.Exec(ctx, query); err != nil {
		_ = tx.Rollback(ctx)
		return 0, fmt.Errorf("failed to create schema_migrations table: %e", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("user_repository find_by_username failed to commit transaction: %e", err)
	}

	currentVersion := version{
		Version: 0,
		Dirty:   false,
	}

	query, _, err = builder.Select("version", "dirty").From(fmt.Sprintf("%s.schema_migrations", schema)).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %e", err)
	}

	tx, err = db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction error: %e", err)
	}
	row := tx.QueryRow(ctx, query)
	if err = row.Scan(&currentVersion.Version, &currentVersion.Dirty); err != nil {
		if err != pgx.ErrNoRows {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("failed to get current version: %e", err)
		}

		q, _, err := builder.Insert(fmt.Sprintf("%s.schema_migrations", schema)).Columns("version", "dirty").Values(0, false).ToSql()
		_, insertErr := tx.Exec(ctx, q, 0, false)
		if insertErr != nil {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("failed to insert current version: %e, sql: %s", err, q)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("user_repository find_by_username failed to commit transaction: %e", err)
	}

	currentVersionInt := currentVersion.Version

	files, err := os.ReadDir(cfg.DatabaseMigrationsDir)
	if err != nil {
		return 0, fmt.Errorf("read migrations dir error: %e", err)
	}

	lastMigrationsVersion := 0
	migrations := make([]migration, 0)
	for _, v := range files {
		filename := fmt.Sprintf("%s/%s", cfg.DatabaseMigrationsDir, v.Name())
		content, errRead := os.ReadFile(filename)
		if errRead != nil {
			return 0, fmt.Errorf("failed to read migration file: %e, filename: %s", errRead, filename)
		}

		version, errParse := strconv.Atoi(v.Name()[:6])
		if errParse != nil {
			return 0, fmt.Errorf("failed to parse version: %e", errParse)
		}

		lastMigrationsVersion = version
		if version <= currentVersionInt {
			continue
		}

		migrations = append(migrations, migration{
			SQL:     strings.ReplaceAll(string(content), "nofronts.", fmt.Sprintf("%s.", schema)),
			Version: version,
		})
	}

	for _, m := range migrations {
		tx, errTx := db.Begin(ctx)
		if errTx != nil {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("failed to start transaction error: %e, version: %d, sql: %s", errTx, m.Version, m.SQL)
		}

		_, errTx = tx.Exec(ctx, m.SQL)
		if errTx != nil {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("failed to run migration error: %e, version: %d, sql: %s", errTx, m.Version, m.SQL)
		}

		q, args, buildErr := builder.Update(fmt.Sprintf("%s.schema_migrations", schema)).Set("version", m.Version).Set("dirty", false).ToSql()
		if buildErr != nil {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("failed to build query: %e", buildErr)
		}

		_, errTx = tx.Exec(ctx, q, args...)
		if errTx != nil {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("failed to set current version error: %e, version: %d, sql: %s", err, m.Version, m.SQL)
		}

		errTx = tx.Commit(ctx)
		if errTx != nil {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("commit migration error: %e", errTx)
		}
	}

	query, _, err = builder.Select("version", "dirty").From(fmt.Sprintf("%s.schema_migrations", schema)).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %e", err)
	}

	tx, err = db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction error: %e", err)
	}

	row = tx.QueryRow(ctx, query)
	if err = row.Scan(&currentVersion.Version, &currentVersion.Dirty); err != nil {
		if err != pgx.ErrNoRows {
			_ = tx.Rollback(ctx)
			return 0, fmt.Errorf("failed to get current version: %e", err)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("user_repository find_by_username failed to commit transaction: %e", err)
	}
	currentVersionInt = currentVersion.Version
	if currentVersionInt < lastMigrationsVersion {
		return currentVersionInt, fmt.Errorf("failed migration, current version: %d, available version: %d", currentVersionInt, lastMigrationsVersion)
	}

	return currentVersionInt, nil
}
