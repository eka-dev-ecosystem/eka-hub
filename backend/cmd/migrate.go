package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"golang/config"
	"golang/database/migrate"
	"golang/pkg/logging"
)

func main() {
	// Define flags
	migrationsPath := flag.String("path", "./migrations", "Path to migrations directory")
	action := flag.String("action", "up", "Migration action: up, down, version, force, drop, create")
	steps := flag.Int("steps", 1, "Number of steps for rollback")
	version := flag.Int("version", -1, "Version to force (for force action)")
	name := flag.String("name", "", "Migration name (for create action)")

	flag.Parse()

	// Load configuration and logger
	cfg := config.Load()
	logging.Setup(cfg.Log.Level, cfg.Log.File)

	logger := logging.Logger()

	switch *action {
	case "up":
		logger.Info("running migrations", slog.String("path", *migrationsPath))
		if err := migrate.RunMigrations(&cfg.Database, *migrationsPath); err != nil {
			logger.Error("failed to run migrations", logging.Err(err))
			os.Exit(1)
		}
		logger.Info("migrations completed successfully")

	case "down":
		logger.Info("rolling back migrations", slog.Int("steps", *steps))
		if err := migrate.RollbackMigrations(&cfg.Database, *migrationsPath, *steps); err != nil {
			logger.Error("failed to rollback migrations", logging.Err(err))
			os.Exit(1)
		}
		logger.Info("rollback completed successfully")

	case "version":
		version, dirty, err := migrate.MigrationVersion(&cfg.Database, *migrationsPath)
		if err != nil {
			logger.Error("failed to get migration version", logging.Err(err))
			os.Exit(1)
		}
		logger.Info("current migration version",
			slog.Uint64("version", uint64(version)),
			slog.Bool("dirty", dirty),
		)

	case "force":
		if *version < 0 {
			logger.Error("please specify a version with -version flag")
			os.Exit(1)
		}
		logger.Info("forcing migration version", slog.Int("version", *version))
		if err := migrate.ForceMigrationVersion(&cfg.Database, *migrationsPath, *version); err != nil {
			logger.Error("failed to force migration version", logging.Err(err))
			os.Exit(1)
		}
		logger.Info("version forced successfully")

	case "drop":
		logger.Warn("this will drop all tables")
		fmt.Print("Are you sure? Type 'yes' to confirm: ")
		var confirm string
		_, _ = fmt.Scanln(&confirm)
		if confirm != "yes" {
			logger.Info("aborted")
			return
		}
		if err := migrate.DropAllTables(&cfg.Database, *migrationsPath); err != nil {
			logger.Error("failed to drop tables", logging.Err(err))
			os.Exit(1)
		}
		logger.Info("all tables dropped successfully")

	case "create":
		if *name == "" {
			logger.Error("please specify a migration name with -name flag")
			os.Exit(1)
		}
		createMigrationFiles(logger, *name)

	default:
		logger.Error("unknown action, available actions: up, down, version, force, drop, create",
			slog.String("action", *action))
		os.Exit(1)
	}
}

func createMigrationFiles(logger *slog.Logger, name string) {
	// Get next migration number
	files, err := os.ReadDir("./migrations")
	if err != nil {
		logger.Error("failed to read migrations directory", logging.Err(err))
		os.Exit(1)
	}

	nextNumber := 1
	if len(files) > 0 {
		// Find the highest number
		for _, file := range files {
			if !file.IsDir() {
				var num int
				_, _ = fmt.Sscanf(file.Name(), "%d", &num)
				if num >= nextNumber {
					nextNumber = num + 1
				}
			}
		}
	}

	// Create up migration
	upFilename := fmt.Sprintf("./migrations/%06d_%s.up.sql", nextNumber, name)
	upContent := fmt.Sprintf("-- Migration: %s (up)\n\n-- Write your UP migration here\n", name)
	if err := os.WriteFile(upFilename, []byte(upContent), 0600); err != nil {
		logger.Error("failed to create up migration", logging.Err(err))
		os.Exit(1)
	}
	logger.Info("created up migration", slog.String("file", upFilename))

	// Create down migration
	downFilename := fmt.Sprintf("./migrations/%06d_%s.down.sql", nextNumber, name)
	downContent := fmt.Sprintf("-- Migration: %s (down)\n\n-- Write your DOWN migration here\n", name)
	if err := os.WriteFile(downFilename, []byte(downContent), 0600); err != nil {
		logger.Error("failed to create down migration", logging.Err(err))
		os.Exit(1)
	}
	logger.Info("created down migration", slog.String("file", downFilename))

	logger.Info("migration files created successfully")
}
