package main

import (
	"github.com/jailtonjunior94/financial/cmd/server"
	"github.com/jailtonjunior94/financial/pkg/bundle"
	migration "github.com/jailtonjunior94/financial/pkg/database/migrate"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "financial",
		Short: "Financial",
	}

	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "Financial Migrations",
		Run: func(cmd *cobra.Command, args []string) {
			container := bundle.NewContainer()
			migrate, err := migration.NewMigrate(container.DB, container.Config.MigratePath, container.Config.DBName)
			if err != nil {
				panic(err)
			}
			if err = migrate.ExecuteMigration(); err != nil {
				panic(err)
			}
		},
	}

	api := &cobra.Command{
		Use:   "api",
		Short: "Financial API",
		Run: func(cmd *cobra.Command, args []string) {
			server := server.NewApiServe()
			server.ApiServer()
		},
	}

	consumers := &cobra.Command{
		Use:   "consumers",
		Short: "Financial Consumers",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	root.AddCommand(migrate, api, consumers)

	root.Execute()
}
