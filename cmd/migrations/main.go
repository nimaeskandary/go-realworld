package main

import (
	"context"
	"log"
	"os"

	"github.com/nimaeskandary/go-realworld/cmd/migrations/app"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"
)

func main() {
	ctx := context.Background()
	cleanupManager := util.NewCleanupManager(ctx, true)
	defer cleanupManager.Cleanup()

	args := app.ParseArgs()

	// setup deps

	configData, err := os.ReadFile(args.ConfigPath)
	if err != nil {
		log.Fatalf("failed to read config file at %v: %v", args.ConfigPath, err)
	}

	moduleList, err := app.ModuleList(args.TargetDatabase, configData)
	if err != nil {
		log.Fatalf("failed to create module list: %v", err)
	}

	var migrationRunner db_types.SqlMigrationRunner
	fxApp := util.CreateFxAppAndExtract(moduleList, &migrationRunner)

	if err := fxApp.Start(ctx); err != nil {
		log.Fatalf("dependency injection system failed to start: %v", err)
	}

	cleanupManager.RegisterCleanupFunc(func() {
		if err := fxApp.Stop(ctx); err != nil {
			log.Printf("dependency injection system failed to stop gracefully: %v", err)
		}
	})

	// run migrations

	err = nil
	switch args.Action {

	case app.ActionApplyAll:
		log.Printf("running %v on database %v...", args.Action, args.TargetDatabase)
		if err := migrationRunner.ApplyAll(ctx); err != nil {
			log.Printf("failed to apply migrations: %v", err)
		}

	case app.ActionApply:
		log.Printf("running %v version %v on database %v...", args.Action, *args.Version, args.TargetDatabase)
		if err := migrationRunner.Apply(ctx, *args.Version); err != nil {
			log.Printf("failed to apply migration version %v: %v", *args.Version, err)
		}

	case app.ActionRollback:
		log.Printf("running %v version %v on database %v...", args.Action, *args.Version, args.TargetDatabase)
		if err := migrationRunner.Rollback(ctx, *args.Version); err != nil {
			log.Printf("failed to rollback migration version %v: %v", *args.Version, err)
		}

	case app.ActionRollbackTo:
		log.Printf("running %v to version %v on database %v...", args.Action, *args.Version, args.TargetDatabase)
		if err := migrationRunner.RollbackTo(ctx, *args.Version); err != nil {
			log.Printf("failed to rollback migrations down to version %v: %v", *args.Version, err)
		}

	case app.ActionStatus:
		status, err := migrationRunner.Status(ctx)
		if err != nil {
			log.Printf("failed to get migration status: %v", err)
		}
		log.Printf("%s", status)

	default:
		log.Printf("unknown migration action: %v", args.Action)
	}
	passed := err == nil

	version, err := migrationRunner.CurrentVersion(ctx)
	if err != nil {
		log.Printf("failed to get current migration version: %v", err)
	}
	log.Printf("current migration version: %v", version)

	if !passed {
		cleanupManager.Cleanup()
		os.Exit(1)
	}
}
