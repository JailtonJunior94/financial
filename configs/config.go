package configs

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceName          string `mapstructure:"SERVICE_NAME"`
	ServiceVersion       string `mapstructure:"SERVICE_VERSION"`
	DevelopmentMode      bool   `mapstructure:"DEVELOPMENT_MODE"`
	HttpServerPort       string `mapstructure:"HTTP_SERVER_PORT"`
	DBDriver             string `mapstructure:"DB_DRIVER"`
	DBHost               string `mapstructure:"DB_HOST"`
	DBPort               string `mapstructure:"DB_PORT"`
	DBUser               string `mapstructure:"DB_USER"`
	DBPassword           string `mapstructure:"DB_PASSWORD"`
	DBName               string `mapstructure:"DB_NAME"`
	DBMaxIdleConns       int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	MigratePath          string `mapstructure:"MIGRATE_PATH"`
	AuthExpirationAt     int    `mapstructure:"AUTH_EXPIRATION_AT"`
	AuthSecretKey        string `mapstructure:"AUTH_SECRET_KEY"`
	OtelTracesExporter   string `mapstructure:"OTEL_TRACES_EXPORTER"`
	OtelExporterEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg *Config

	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if err := viper.Unmarshal(&cfg); err != nil {
			log.Printf("unable to decode into struct, %v", err)
		}
	})

	return cfg, nil
}
