package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/nimaeskandary/go-realworld/pkg/database/types"
	"github.com/pressly/goose/v3"
	"github.com/samber/lo"
)

type gooseMigrationRunner struct {
	db                db_types.SQLDatabase
	migrationsProvder db_types.MigrationsProvider
	gooseProvider     *goose.Provider
}

func NewGooseMigrationRunner(db db_types.SQLDatabase, migrationsProvider db_types.MigrationsProvider) (db_types.SqlMigrationRunner, error) {
	return &gooseMigrationRunner{
		gooseProvider:     nil,
		db:                db,
		migrationsProvder: migrationsProvider,
	}, nil
}

func (r *gooseMigrationRunner) Start(ctx context.Context) error {
	mp := r.migrationsProvder
	gooseCodeMigrations := lo.Map(*mp.GetCodeMigrations(), func(cm db_types.GoMigration, _ int) *goose.Migration {
		return goose.NewGoMigration(
			int64(cm.Version()),
			&goose.GoFunc{RunTx: cm.Up()},
			&goose.GoFunc{RunTx: cm.Down()},
		)
	})

	var dialect goose.Dialect
	switch r.db.GetDialect() {
	case "postgres":
		dialect = goose.DialectPostgres
	default:
		return fmt.Errorf("unknown dialect for goose migration runner: %v", r.db.GetDialect())
	}

	gooseProvider, err := goose.NewProvider(dialect, r.db.GetDB(), mp.GetSqlMigrationsFs(), goose.WithGoMigrations(gooseCodeMigrations...))
	if err != nil {
		return fmt.Errorf("failed to create goose provider: %w", err)
	}

	r.gooseProvider = gooseProvider
	return nil
}

func (r *gooseMigrationRunner) Stop(_ context.Context) error {
	r.gooseProvider = nil
	return nil
}

func (r *gooseMigrationRunner) Apply(ctx context.Context, version int64) error {
	result, err := r.gooseProvider.ApplyVersion(ctx, version, true)
	if err != nil {
		return fmt.Errorf("failed to apply migration version %v: %w", version, err)
	}
	if result.Error != nil {
		return fmt.Errorf("failed to apply migration version %v: %w", version, result.Error)
	}

	return nil
}

func (r *gooseMigrationRunner) ApplyAll(ctx context.Context) error {
	result, err := r.gooseProvider.Up(ctx)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	resultErrs := lo.Reduce(result, func(agg error, item *goose.MigrationResult, _ int) error {
		if item.Error != nil {
			return errors.Join(agg, fmt.Errorf("failed to apply migration version %v: %w", item.Source.Version, item.Error))
		}
		return agg
	}, nil)

	return resultErrs
}

func (r *gooseMigrationRunner) Rollback(ctx context.Context, version int64) error {
	result, err := r.gooseProvider.ApplyVersion(ctx, version, false)
	if err != nil {
		return fmt.Errorf("failed to rollback migration %v: %w", version, err)
	}

	if result.Error != nil {
		return fmt.Errorf("failed to rollback migration %v: %w", version, result.Error)
	}

	return nil
}

func (r *gooseMigrationRunner) RollbackTo(ctx context.Context, version int64) error {
	result, err := r.gooseProvider.DownTo(ctx, version)
	if err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	resultErrs := lo.Reduce(result, func(agg error, item *goose.MigrationResult, _ int) error {
		if item.Error != nil {
			return errors.Join(agg, fmt.Errorf("failed to rollback migration version %v: %w", item.Source.Version, item.Error))
		}
		return agg
	}, nil)

	return resultErrs
}

func (r *gooseMigrationRunner) Status(ctx context.Context) (string, error) {
	result, err := r.gooseProvider.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get migration status: %w", err)
	}

	return lo.Reduce(result, func(agg string, item *goose.MigrationStatus, _ int) string {
		s := fmt.Sprintf("%v - Version %v", item.State, item.Source.Version)
		return fmt.Sprintf("%v\n%v", agg, s)
	}, ""), nil
}

func (r *gooseMigrationRunner) CurrentVersion(ctx context.Context) (int64, error) {
	version, err := r.gooseProvider.GetDBVersion(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current migration version: %w", err)
	}

	return version, nil
}
