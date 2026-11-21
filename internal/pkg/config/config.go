package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

const SourceKey = "CONFIG_SOURCE"
const SourceEnv = "ENVIRONMENT"

type ChatApiConfig struct {
	AppName string `mapstructure:"APP_NAME"`
	AppEnv  string `mapstructure:"APP_ENV"`
	AppPort int    `mapstructure:"APP_PORT"`

	DatabaseHost        string `mapstructure:"DB_HOST"`
	DatabasePort        string `mapstructure:"DB_PORT"`
	DatabaseUsername    string `mapstructure:"DB_USERNAME"`
	DatabasePassword    string `mapstructure:"DB_PASSWORD"`
	DatabaseName        string `mapstructure:"DB_DATABASE_NAME"`
	DatabaseSSLMode     string `mapstructure:"DB_SSL_MODE"`
	DatabaseSSLRootCert string `mapstructure:"DB_SSL_ROOT_CERT"`

	SwaggerHostUrl    string `mapstructure:"SWAGGER_HOST_URL"`
	SwaggerHostScheme string `mapstructure:"SWAGGER_HOST_SCHEME"`
	SwaggerUsername   string `mapstructure:"SWAGGER_USERNAME"`
	SwaggerPassword   string `mapstructure:"SWAGGER_PASSWORD"`

	AuthSecret       string `mapstructure:"AUTH_SECRET"`
	AuthExpiryPeriod int    `mapstructure:"AUTH_EXPIRY_PERIOD"`
	AllowedOrigins   string `mapstructure:"ALLOWED_ORIGINS"`

	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	StorageProvider string `mapstructure:"STORAGE_PROVIDER"`
	MinIOEndpoint   string `mapstructure:"MINIO_ENDPOINT"`
	MinIOAccessKey  string `mapstructure:"MINIO_ACCESS_KEY"`
	MinIOSecretKey  string `mapstructure:"MINIO_SECRET_KEY"`
	MinIOBucket     string `mapstructure:"MINIO_BUCKET"`
	MinIOUseSSL     bool   `mapstructure:"MINIO_USE_SSL"`
	MinIORegion     string `mapstructure:"MINIO_REGION"`
	MinIOPublicURL  string `mapstructure:"MINIO_PUBLIC_URL"`
}

type Options struct {
	ConfigFile       string
	ConfigFileSource string
}

func NewConfig(opt Options) (ChatApiConfig, error) {
	return NewFromEnvironmentVariable(opt)
}

func NewFromEnvironmentVariable(opt Options) (cfg ChatApiConfig, err error) {
	v := viper.New()
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	for _, key := range []string{
		"APP_NAME", "APP_ENV", "APP_PORT", "DB_HOST", "DB_PORT", "DB_USERNAME",
		"DB_PASSWORD", "DB_DATABASE_NAME", "DB_SSL_MODE", "DB_SSL_ROOT_CERT", "SWAGGER_HOST_URL",
		"SWAGGER_HOST_SCHEME", "SWAGGER_USERNAME", "SWAGGER_PASSWORD", "AUTH_SECRET",
		"AUTH_EXPIRY_PERIOD", "ALLOWED_ORIGINS", "REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB",
		"STORAGE_PROVIDER", "MINIO_ENDPOINT", "MINIO_ACCESS_KEY", "MINIO_SECRET_KEY",
		"MINIO_BUCKET", "MINIO_USE_SSL", "MINIO_REGION", "MINIO_PUBLIC_URL",
	} {
		if err := v.BindEnv(key); err != nil {
			return cfg, fmt.Errorf("failed to bind configuration key %s: %w", key, err)
		}
	}
	if opt.ConfigFile != "" {
		v.SetConfigFile(opt.ConfigFile)
		if readErr := v.ReadInConfig(); readErr != nil {
			if _, ok := readErr.(viper.ConfigFileNotFoundError); !ok {
				return cfg, fmt.Errorf("failed to read config file: %w", readErr)
			}
		}
	}
	err = v.Unmarshal(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to load configuration: %v", err)
	}
	if cfg.AppPort == 0 {
		cfg.AppPort = 8080
	}
	cfg.AppEnv = strings.ToLower(strings.TrimSpace(cfg.AppEnv))
	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}
	if cfg.AppEnv != "development" && cfg.AppEnv != "test" && cfg.AppEnv != "production" {
		return cfg, fmt.Errorf("APP_ENV must be development, test, or production")
	}
	if cfg.DatabaseSSLMode == "" {
		cfg.DatabaseSSLMode = "disable"
	}
	if cfg.RedisAddr == "" {
		cfg.RedisAddr = "localhost:6379"
	}
	if cfg.AuthExpiryPeriod == 0 && cfg.AppEnv != "production" {
		cfg.AuthExpiryPeriod = 24
	}
	if cfg.AuthExpiryPeriod <= 0 || cfg.AuthExpiryPeriod > 8760 {
		return cfg, fmt.Errorf("AUTH_EXPIRY_PERIOD must be between 1 and 8760 hours")
	}
	if cfg.RedisDB < 0 {
		return cfg, fmt.Errorf("REDIS_DB must be zero or greater")
	}
	if cfg.AllowedOrigins == "" {
		if strings.EqualFold(cfg.AppEnv, "production") {
			return cfg, fmt.Errorf("ALLOWED_ORIGINS is required in production")
		}
		cfg.AllowedOrigins = "http://localhost:8081,http://127.0.0.1:8081,http://localhost:19006"
	}
	if strings.EqualFold(cfg.AppEnv, "production") && strings.Contains(cfg.AllowedOrigins, "*") {
		return cfg, fmt.Errorf("wildcard ALLOWED_ORIGINS is not allowed in production")
	}
	if cfg.AppEnv == "production" {
		for _, origin := range strings.Split(cfg.AllowedOrigins, ",") {
			parsed, parseErr := url.Parse(strings.TrimSpace(origin))
			if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
				return cfg, fmt.Errorf("ALLOWED_ORIGINS must contain only HTTPS origins in production")
			}
		}
	}
	if len(cfg.AuthSecret) < 32 {
		return cfg, fmt.Errorf("AUTH_SECRET must contain at least 32 characters")
	}
	if cfg.DatabaseHost == "" || cfg.DatabasePort == "" || cfg.DatabaseUsername == "" || cfg.DatabasePassword == "" || cfg.DatabaseName == "" {
		return cfg, fmt.Errorf("DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, and DB_DATABASE_NAME are required")
	}
	if cfg.AppPort < 1 || cfg.AppPort > 65535 {
		return cfg, fmt.Errorf("APP_PORT must be between 1 and 65535")
	}
	if cfg.AppEnv == "production" {
		if cfg.AuthSecret == "replace-with-a-random-secret-of-at-least-32-characters" {
			return cfg, fmt.Errorf("AUTH_SECRET must be replaced with a production secret")
		}
		if cfg.DatabasePassword == "local-db-password-change-me" || cfg.RedisPassword == "local-redis-password-change-me" || cfg.RedisPassword == "" {
			return cfg, fmt.Errorf("production database and Redis credentials must be explicitly configured")
		}
		if cfg.DatabaseSSLMode != "verify-ca" && cfg.DatabaseSSLMode != "verify-full" {
			return cfg, fmt.Errorf("DB_SSL_MODE must be verify-ca or verify-full in production")
		}
	}

	return cfg, nil
}
