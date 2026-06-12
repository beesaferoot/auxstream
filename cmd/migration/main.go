package main

import (
	"fmt"
	"os"
	"strings"

	"auxstream/internal/db"

	"github.com/beesaferoot/gorm-migrate/migration"
	"github.com/beesaferoot/gorm-migrate/migration/commands"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type DBModelRegistry struct{}

func (r *DBModelRegistry) GetModels() map[string]any {
	return db.ModelTypeRegistry
}

func init() {
	migration.GlobalModelRegistry = &DBModelRegistry{}
}

// migrationsDir mirrors the gorm-migrate loader's resolution: MIGRATIONS_PATH or
// the default "migrations" directory, relative to the working directory.
func migrationsDir() string {
	if dir := os.Getenv("MIGRATIONS_PATH"); dir != "" {
		return dir
	}
	return "migrations"
}

// requireMigrationFiles fails loudly when no migration source files are present.
// gorm-migrate parses migration SQL from the *.go files on disk at runtime, so a
// missing/empty directory makes `up` report "No pending migrations" and exit 0 —
// a silent success that lets a deployment start against an unmigrated database
// (e.g. a runtime image that forgot to ship the migrations/ directory). Refuse to
// proceed instead, so the deploy's migrate step blocks the api/worker from starting.
func requireMigrationFiles(cmd *cobra.Command, _ []string) error {
	dir := migrationsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("migrations directory %q is not readable: %w (is it shipped in the deployment image?)", dir, err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
			return nil
		}
	}
	return fmt.Errorf("no migration files (*.go) found in %q — refusing to continue; the deployment image is likely missing the migrations directory", dir)
}

func main() {
	_ = godotenv.Load("app.env") // optionally load environment file
	rootCmd := &cobra.Command{
		Use:   "migration",
		Short: "Database Migration Tool",
	}

	upCmd := commands.UpCmd()
	// Guard against the silent-success case; usage text is irrelevant for this
	// operational failure, so suppress it and let the error speak for itself.
	upCmd.SilenceUsage = true
	upCmd.PreRunE = requireMigrationFiles

	rootCmd.AddCommand(
		commands.RegisterCmd(),
		commands.InitCmd(),
		commands.GenerateCmd(),
		upCmd,
		commands.DownCmd(),
		commands.StatusCmd(),
		commands.HistoryCmd(),
		commands.ValidateCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		// cobra has already printed the error; exit non-zero (cleanly, no stack
		// trace) so a CI/compose deploy step halts here.
		os.Exit(1)
	}
}
