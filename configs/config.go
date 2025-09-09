package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type (
	Config struct {
		DBConfig   DBConfig   `mapstructure:",squash"`
		HTTPConfig HTTPConfig `mapstructure:",squash"`
		O11yConfig O11yConfig `mapstructure:",squash"`
		AuthConfig AuthConfig `mapstructure:",squash"`
	}

	DBConfig struct {
		Driver         string `mapstructure:"DB_DRIVER"`
		Host           string `mapstructure:"DB_HOST"`
		Port           string `mapstructure:"DB_PORT"`
		User           string `mapstructure:"DB_USER"`
		Password       string `mapstructure:"DB_PASSWORD"`
		Name           string `mapstructure:"DB_NAME"`
		DBMaxIdleConns int    `mapstructure:"DB_MAX_IDLE_CONNS"`
		MigratePath    string `mapstructure:"MIGRATE_PATH"`
	}

	HTTPConfig struct {
		Port string `mapstructure:"HTTP_PORT"`
	}

	O11yConfig struct {
		ServiceName          string `mapstructure:"OTEL_SERVICE_NAME"`
		ServiceVersion       string `mapstructure:"OTEL_SERVICE_VERSION"`
		ExporterEndpoint     string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
		ExporterEndpointHTTP string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT_HTTP"`
	}

	AuthConfig struct {
		AuthSecretKey     string `mapstructure:"AUTH_SECRET_KEY"`
		AuthTokenDuration int    `mapstructure:"AUTH_TOKEN_DURATION"`
	}
)

func LoadConfig(path string) (*Config, error) {
	var config *Config

	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config file, %v", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	return config, nil
}
