package main

import (
	"context"
	"log"

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
			container := bundle.NewContainer(context.Background())
			switch container.Config.DBDriver {
			case "postgres":
				{
					migrate, err := migration.NewMigrateCockroachDB(container.Logger, container.DB, container.Config.MigratePath, container.Config.DBName)
					if err != nil {
						log.Fatal(err)
					}
					if err = migrate.Execute(); err != nil {
						log.Fatal(err)
					}
				}
			case "mysql":
				{
					migrate, err := migration.NewMigrateMySql(container.Logger, container.DB, container.Config.MigratePath, container.Config.DBName)
					if err != nil {
						log.Fatal(err)
					}
					if err = migrate.Execute(); err != nil {
						log.Fatal(err)
					}
				}
			}
		},
	}

	api := &cobra.Command{
		Use:   "api",
		Short: "Financial API",
		Run: func(cmd *cobra.Command, args []string) {
			server.NewApiServer().Server()
		},
	}

	consumers := &cobra.Command{
		Use:   "consumers",
		Short: "Financial Consumers",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("not implement")
		},
	}

	root.AddCommand(migrate, api, consumers)
	root.Execute()
}
