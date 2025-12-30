package main

import (
	"log"

	"github.com/jailtonjunior94/financial/cmd/server"

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
			// migrate, err := migration.NewMigrateCockroachDB(
			// 	container.DB,
			// 	container.Config.DBConfig.MigratePath,
			// 	container.Config.DBConfig.Name,
			// )
			// if err != nil {
			// 	log.Fatalf("error initializing migrations: %v", err)
			// }

			// if err = migrate.Execute(); err != nil {
			// 	log.Fatalf("error executing migrations: %v", err)
			// }
		},
	}

	api := &cobra.Command{
		Use:   "api",
		Short: "Financial API",
		Run: func(cmd *cobra.Command, args []string) {
			server.Run()
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
	if err := root.Execute(); err != nil {
		log.Fatalf("error executing command: %v", err)
	}
}
