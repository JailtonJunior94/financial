package main

import (
	"log"

	"github.com/jailtonjunior94/financial/cmd/server"
	"github.com/jailtonjunior94/financial/pkg/bundle"
	migration "github.com/jailtonjunior94/financial/pkg/database/migrate"
	databaseMySql "github.com/jailtonjunior94/financial/pkg/database/mysql"

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

			mySQL, _ := databaseMySql.NewMySqlDatabase(nil)

			migrateMySql, err := migration.NewMigrateMySql(mySQL, container.Config.MigratePath, container.Config.DBName)
			if err != nil {
				panic(err)
			}
			if err = migrateMySql.ExecuteMigration(); err != nil {
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
			log.Println("not implement")
		},
	}

	root.AddCommand(migrate, api, consumers)

	root.Execute()
}
