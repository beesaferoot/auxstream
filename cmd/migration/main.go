package main

import (
	"auxstream/db"
	"github.com/beesaferoot/gorm-schema/migration"
	"github.com/beesaferoot/gorm-schema/migration/commands"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type DBModelRegistry struct{}

func (r *DBModelRegistry) GetModels() map[string]any{
	return db.ModelTypeRegistry 
}

func init() {
	migration.GlobalModelRegistry = &DBModelRegistry{}
}

func main() {
	_ = godotenv.Load("app.env") // optionally load environment file
	rootCmd := &cobra.Command{
		Use:   "migration",
		Short: "Database Migration Tool",
	}

	rootCmd.AddCommand(
		commands.RegisterCmd(),
		commands.InitCmd(),
		commands.GenerateCmd(),
		commands.UpCmd(),
		commands.DownCmd(),
		commands.StatusCmd(),
		commands.HistoryCmd(),
		commands.ValidateCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
