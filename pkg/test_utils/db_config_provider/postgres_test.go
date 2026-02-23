package db_config_provider_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"

	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/db_config_provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Test_PostgresSqlDbConfigProvider(t *testing.T) {
	templateName := "db_config_provider_test_template"
	testCfg := db_types.SqlDbConfig{
		Host:     "localhost",
		Port:     15432,
		Username: "testpostgres",
		Password: "testpassword",
		DBName:   templateName,
		SslMode:  "disable",
	}

	migrationCount := atomic.Int32{}

	migrationFn := func(ctx context.Context, db db_types.SQLDatabase) error {
		migrationCount.Add(1)
		_, err := db.GetDB().ExecContext(ctx, "CREATE TABLE IF NOT EXISTS migration_marker (id INT)")
		return err
	}

	provider := db_config_provider.NewPostgresSqlDbConfigProvider(testCfg, migrationFn)
	cleanupDatabase(t.Context(), t, testCfg)

	// template should not exist yet
	assert.False(t, databaseExists(t, testCfg, testCfg.DBName))

	t.Cleanup(func() { cleanupDatabase(context.Background(), t, testCfg) })

	t.Run("should make databases off a shared template and handle concurrency", func(t *testing.T) {
		eg, egCtx := errgroup.WithContext(t.Context())

		const concurrentRequests = 5
		cfgsChan := make(chan db_types.SqlDbConfig, concurrentRequests)

		for range concurrentRequests {
			eg.Go(func() error {
				cfg, err := provider.GetFreshDbConfig(egCtx)
				t.Cleanup(func() { _ = provider.Cleanup(context.Background(), cfg) })

				if err != nil {
					return err
				}

				cfgsChan <- cfg
				return nil
			})
		}

		assert.NoError(t, eg.Wait())
		close(cfgsChan)

		assert.True(t, databaseExists(t, testCfg, testCfg.DBName), "Template database should be created")
		assert.Equal(t, int32(5), migrationCount.Load(), "Migrations are run on the template db for each fresh db call")

		seen := make(map[string]any)

		for cfg := range cfgsChan {
			assert.True(t, databaseExists(t, testCfg, cfg.DBName))
			assert.True(t, tableExists(t, cfg, "migration_marker"), "Cloned DB should have the migration table")
			_, seenBefore := seen[cfg.DBName]
			assert.False(t, seenBefore, "Each config should have a unique DBName")
			seen[cfg.DBName] = struct{}{}
		}
	})
}

func getMaintenanceDB(t *testing.T, testCfg db_types.SqlDbConfig) *sql.DB {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		testCfg.Username, testCfg.Password, testCfg.Host, testCfg.Port, "postgres", testCfg.SslMode)
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	return db
}

func cleanupDatabase(ctx context.Context, t *testing.T, testCfg db_types.SqlDbConfig) {
	db := getMaintenanceDB(t, testCfg)
	defer func() { _ = db.Close() }()
	_, _ = db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", testCfg.DBName))
}

func databaseExists(t *testing.T, testCfg db_types.SqlDbConfig, dbName string) bool {
	db := getMaintenanceDB(t, testCfg)
	defer func() { _ = db.Close() }()
	var exists bool
	err := db.QueryRowContext(t.Context(), "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	require.NoError(t, err)
	return exists
}

func tableExists(t *testing.T, cfg db_types.SqlDbConfig, tableName string) bool {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SslMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return false
	}
	defer func() { _ = db.Close() }()
	var exists bool
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)"
	err = db.QueryRowContext(t.Context(), query, tableName).Scan(&exists)
	return err == nil && exists
}
