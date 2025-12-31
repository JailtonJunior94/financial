package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Environment string     `mapstructure:"ENVIRONMENT"`
		DBConfig    DBConfig   `mapstructure:",squash"`
		HTTPConfig  HTTPConfig `mapstructure:",squash"`
		O11yConfig  O11yConfig `mapstructure:",squash"`
		AuthConfig  AuthConfig `mapstructure:",squash"`
	}

	DBConfig struct {
		Driver                   string `mapstructure:"DB_DRIVER"`
		Host                     string `mapstructure:"DB_HOST"`
		Port                     string `mapstructure:"DB_PORT"`
		User                     string `mapstructure:"DB_USER"`
		Password                 string `mapstructure:"DB_PASSWORD"`
		Name                     string `mapstructure:"DB_NAME"`
		DBMaxIdleConns           int    `mapstructure:"DB_MAX_IDLE_CONNS"`
		DBMaxOpenConns           int    `mapstructure:"DB_MAX_OPEN_CONNS"`
		DBConnMaxLifeTimeMinutes int    `mapstructure:"DB_CONN_MAX_LIFE_TIME_MINUTES"`
		DBConnMaxIdleTimeMinutes int    `mapstructure:"DB_CONN_MAX_IDLE_TIME_MINUTES"`
		MigratePath              string `mapstructure:"MIGRATE_PATH"`
	}

	HTTPConfig struct {
		Port string `mapstructure:"HTTP_PORT"`
	}

	O11yConfig struct {
		ServiceName      string  `mapstructure:"OTEL_SERVICE_NAME"`
		ServiceVersion   string  `mapstructure:"OTEL_SERVICE_VERSION"`
		ExporterEndpoint string  `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
		ExporterProtocol string  `mapstructure:"OTEL_EXPORTER_OTLP_PROTOCOL"`
		ExporterInsecure bool    `mapstructure:"OTEL_EXPORTER_OTLP_INSECURE"`
		TraceSampleRate  float64 `mapstructure:"OTEL_TRACE_SAMPLE_RATE"`
		LogLevel         string  `mapstructure:"LOG_LEVEL"`
		LogFormat        string  `mapstructure:"LOG_FORMAT"`
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

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: error reading config file, %v", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("config: unable to decode into struct, %v", err)
	}

	return config, nil
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
}
