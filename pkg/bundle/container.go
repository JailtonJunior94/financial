package bundle

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/configs"
	database "github.com/jailtonjunior94/financial/pkg/database/postgres"
)

type container struct {
	Config *configs.Config
	DB     *sql.DB
}

func NewContainer() *container {
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	dbConnection, err := database.NewPostgresDatabase(config)
	if err != nil {
		panic(err)
	}

	return &container{
		Config: config,
		DB:     dbConnection,
	}
}
