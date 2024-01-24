package configs

import (
	"github.com/spf13/viper"
)

type Config struct {
	DevelopmentMode  bool   `mapstructure:"DEVELOPMENT_MODE"`
	HttpServerPort   string `mapstructure:"HTTP_SERVER_PORT"`
	DBDriver         string `mapstructure:"DB_DRIVER"`
	DBHost           string `mapstructure:"DB_HOST"`
	DBPort           string `mapstructure:"DB_PORT"`
	DBUser           string `mapstructure:"DB_USER"`
	DBPassword       string `mapstructure:"DB_PASSWORD"`
	DBName           string `mapstructure:"DB_NAME"`
	DBMaxIdleConns   int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	MigratePath      string `mapstructure:"MIGRATE_PATH"`
	AuthExpirationAt int    `mapstructure:"AUTH_EXPIRATION_AT"`
	AuthSecretKey    string `mapstructure:"AUTH_SECRET_KEY"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg *Config

	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
