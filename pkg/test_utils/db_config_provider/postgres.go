package db_config_provider

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/nimaeskandary/go-realworld/pkg/database"
	"github.com/nimaeskandary/go-realworld/pkg/database/types"
)

// random constant
const postgresAdvisoryLockId = 5436123451

// metadataTableName is the name of the table used to track template and test database creation times
const metadataTableName = "template_metadata"

// templateTTLMinutes - things could get wonky if a template db is never torn down, e.g.
// developers changing migration code for an in development migration version. If you see this and
// are facing an issue like that, don't need to wait, can restart the container via docker compose
// since it has no volume and will be wiped.
const templateTTLMinutes = 5

// testDbTTLMinutes is the time after which test databases are considered expired and can be cleaned up. This is for stray
// databases that were never cleaned up for some reason.
const testDbTTLMinutes = 180

var getLockSql = fmt.Sprintf("SELECT pg_advisory_lock(%v)", postgresAdvisoryLockId)

var unlockSql = fmt.Sprintf("SELECT pg_advisory_unlock(%v)", postgresAdvisoryLockId)

func createDatabaseFromTemplateSql(templateName string, freshDbName string) string {
	return fmt.Sprintf("CREATE DATABASE %v TEMPLATE %v", freshDbName, templateName)
}

func databaseExistsSql(dbName string) string {
	return fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%v'", dbName)
}

func createMetadataTableSql() string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v (db_name TEXT PRIMARY KEY, created_at TIMESTAMP)", metadataTableName)
}

func getCreatedAtSql(templateDbName string) string {
	return fmt.Sprintf("SELECT created_at FROM %v WHERE db_name = '%v'", metadataTableName, templateDbName)
}

func dropDatabaseSql(dbName string) string {
	return fmt.Sprintf("DROP DATABASE IF EXISTS %v WITH (FORCE)", dbName)
}

func createDatabaseSql(dbName string) string {
	return fmt.Sprintf("CREATE DATABASE %v", dbName)
}

func insertCreatedAtSql(metadataTableName string, templateDbName string) string {
	return fmt.Sprintf("INSERT INTO %v (db_name, created_at) VALUES ('%v', NOW()) ON CONFLICT (db_name) DO UPDATE SET created_at = EXCLUDED.created_at", metadataTableName, templateDbName)
}

func deleteMetadataSql(metadataTableName string, dbName string) string {
	return fmt.Sprintf("DELETE FROM %v WHERE db_name = '%v'", metadataTableName, dbName)
}

func getExpiredTestDatabasesSql(metadataTableName string) string {
	return fmt.Sprintf("SELECT db_name FROM %v WHERE created_at < NOW() - INTERVAL '%v minutes'", metadataTableName, testDbTTLMinutes)
}

// postgresSqlDbConfigProvider implements SqlDbConfigProvider for Postgres databases. The general strategy used here is:
// * maintain a template database with all migrations already run for the requested db config
// * each call to GetFreshDb clones a new database from the template database
// * periodically, the template database is rebuilt from scratch
//
// The goal here is to provide an isolated test database for each requestor, and speed things up
// by using template dbs instead of running migrations each time. Should err on the side of using this
// once per test file, and not per sub test case, due to the overhead of creating databases
type postgresSqlDbConfigProvider struct {
	cfg             db_types.SqlDbConfig
	runMigrationsFn func(ctx context.Context, db db_types.SQLDatabase) error
}

func NewPostgresSqlDbConfigProvider(
	cfg db_types.SqlDbConfig,
	runMigrationsFn func(ctx context.Context, db db_types.SQLDatabase) error,
) SqlDbConfigProvider {
	return &postgresSqlDbConfigProvider{
		cfg:             cfg,
		runMigrationsFn: runMigrationsFn,
	}
}

// GetFreshDb provides a config for a fresh instance of the db. To achieve this, a template database with migrations is maintaned, and then
// re used to clone fresh databases if the template is still valid. If the template is stale, it is dropped and recreated.
// It is safe for concurrent use
func (p *postgresSqlDbConfigProvider) GetFreshDbConfig(ctx context.Context) (db_types.SqlDbConfig, error) {
	mainDb, err := createPostgresMaintenanceConnection(ctx, p.cfg)
	if err != nil {
		return db_types.SqlDbConfig{}, fmt.Errorf("failed to create maintenance connection: %v", err)
	}
	defer func() {
		_ = mainDb.Close()
	}()

	// Blocking call to aquire application level lock, this is what makes this process safe for concurrent use.
	// This gets released by us in the defer call, but as a failsafe, gets released by postgres when the connection is closed
	_, _ = mainDb.ExecContext(ctx, getLockSql)
	defer func() {
		_, _ = mainDb.ExecContext(ctx, unlockSql)
	}()

	// create metadata table for age tracking
	_, err = mainDb.ExecContext(ctx, createMetadataTableSql())
	if err != nil {
		return db_types.SqlDbConfig{}, fmt.Errorf("failed to create template metadata table: %v", err)
	}

	// Cleanup expired test dbs that never got torn down
	err = cleanupExpiredTestDatabases(ctx, mainDb)
	if err != nil {
		return db_types.SqlDbConfig{}, fmt.Errorf("failed to cleanup expired test dbs: %v", err)
	}

	// check if template needs rebuild
	rebuild, err := p.needsRebuild(ctx, mainDb, p.cfg.DBName)
	if err != nil {
		return db_types.SqlDbConfig{}, fmt.Errorf("failed to check if template needs rebuild: %v", err)
	}

	if rebuild {
		err = p.rebuildTemplate(ctx, mainDb, p.cfg)
		if err != nil {
			return db_types.SqlDbConfig{}, fmt.Errorf("failed to rebuild template: %v", err)
		}
	}

	// ensure migrations are up to date on template
	err = p.runTemplateMigrations(ctx, p.cfg)
	if err != nil {
		return db_types.SqlDbConfig{}, fmt.Errorf("failed to run migrations on template: %v", err)
	}

	// create fresh db from template
	freshDbName := fmt.Sprintf("test_%v_%v_%v", p.cfg.DBName, rand.IntN(10000), time.Now().UnixMilli())
	if _, err := mainDb.ExecContext(ctx, createDatabaseFromTemplateSql(p.cfg.DBName, freshDbName)); err != nil {
		return db_types.SqlDbConfig{}, fmt.Errorf("failed to clone template: %v", err)
	}

	// track this db in the metadata table
	_, err = mainDb.ExecContext(ctx, insertCreatedAtSql(metadataTableName, freshDbName))
	if err != nil {
		return db_types.SqlDbConfig{}, fmt.Errorf("failed to insert fresh db %v metadata: %v", freshDbName, err)
	}

	newCfg := p.cfg
	newCfg.DBName = freshDbName
	return newCfg, nil
}

// Cleanup drops the test database created with GetFreshDbConfig
func (p *postgresSqlDbConfigProvider) Cleanup(ctx context.Context, cfg db_types.SqlDbConfig) error {
	mainDb, err := createPostgresMaintenanceConnection(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create maintenance connection: %v", err)
	}
	defer func() {
		_ = mainDb.Close()
	}()

	_, err = mainDb.ExecContext(ctx, dropDatabaseSql(cfg.DBName))
	if err != nil {
		return fmt.Errorf("failed to drop test database: %v", err)
	}

	_, err = mainDb.ExecContext(ctx, deleteMetadataSql(metadataTableName, cfg.DBName))
	if err != nil {
		return fmt.Errorf("failed to delete metadata for test database %v: %v", cfg.DBName, err)
	}

	return nil
}

// needsRebuilds checks if the template database needs to be rebuilt
func (p *postgresSqlDbConfigProvider) needsRebuild(ctx context.Context, mainDb *sql.DB, templateDbName string) (bool, error) {
	var exists bool
	query := databaseExistsSql(templateDbName)
	err := mainDb.QueryRowContext(ctx, query).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, fmt.Errorf("failed to check if template db exists: %v", err)
	}

	if !exists {
		return true, nil
	}

	// check if template is old
	var createdAt time.Time
	err = mainDb.QueryRowContext(ctx, getCreatedAtSql(templateDbName)).Scan(&createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, fmt.Errorf("failed to get template creation time: %v", err)
	}

	if time.Since(createdAt) > templateTTLMinutes*time.Minute {
		return true, nil
	}

	return false, nil
}

func (p *postgresSqlDbConfigProvider) rebuildTemplate(ctx context.Context, mainDb *sql.DB, cfg db_types.SqlDbConfig) error {
	_, _ = mainDb.ExecContext(ctx, dropDatabaseSql(cfg.DBName))
	_, _ = mainDb.ExecContext(ctx, createDatabaseSql(cfg.DBName))

	_, err := mainDb.ExecContext(ctx, insertCreatedAtSql(metadataTableName, cfg.DBName))
	if err != nil {
		return fmt.Errorf("failed to insert template metadata: %v", err)
	}
	return nil
}

func (p *postgresSqlDbConfigProvider) runTemplateMigrations(ctx context.Context, cfg db_types.SqlDbConfig) error {
	tempDb, err := database.NewPostgresSQLDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to temp template db: %v", err)
	}
	err = tempDb.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start temp template db: %v", err)
	}
	defer func() { _ = tempDb.Stop(ctx) }()

	err = p.runMigrationsFn(ctx, tempDb)
	if err != nil {
		return fmt.Errorf("failed to run migrations on template db: %v", err)
	}

	return nil
}

func cleanupExpiredTestDatabases(ctx context.Context, mainDb *sql.DB) error {
	rows, err := mainDb.QueryContext(ctx, getExpiredTestDatabasesSql(metadataTableName))
	defer func() {
		_ = rows.Close()
	}()

	if err != nil {
		return fmt.Errorf("failed to query expired test databases: %v", err)
	}

	for rows.Next() {
		var dbName string
		err := rows.Scan(&dbName)
		if err != nil {
			return fmt.Errorf("failed to scan expired test database name: %v", err)
		}

		_, err = mainDb.ExecContext(ctx, dropDatabaseSql(dbName))
		if err != nil {
			return fmt.Errorf("failed to drop expired test database %v: %v", dbName, err)
		}

		_, err = mainDb.ExecContext(ctx, deleteMetadataSql(metadataTableName, dbName))
		if err != nil {
			return fmt.Errorf("failed to delete metadata for expired test database %v: %v", dbName, err)
		}
	}
	return nil
}

// createPostgresMaintenanceConnection creates a connection to the "postgres" maintenance database, this is needed to create and drop other databases
func createPostgresMaintenanceConnection(ctx context.Context, cfg db_types.SqlDbConfig) (*sql.DB, error) {
	mainCfg := cfg
	mainCfg.DBName = "postgres"
	db, err := database.NewPostgresSQLDatabase(mainCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to maintenance DB: %v", err)
	}
	err = db.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start maintenance DB connection: %v", err)
	}
	return db.GetDB(), nil
}
