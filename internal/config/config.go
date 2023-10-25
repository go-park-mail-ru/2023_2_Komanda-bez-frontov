package config

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog"
)

const (
	defaultHTTPPort                    = "8080"
	defaultHTTPReadTimeout             = 5 * time.Second
	defaultHTTPWriteTimeout            = 5 * time.Second
	defaultLogLevel                    = "error"
	defaultCookieExpiration            = 24 * time.Hour
	defaultLogRequests                 = "true"
	defaultDatabaseMaxConnections      = 40
	defaultDatabaseMigrationsDir       = "./db/migrations"
	defaultDatabaseConnectMaxRetries   = 20
	defaultDatabaseConnectRetryTimeout = 1 * time.Second
	defaultAcquireTimeout              = 1 * time.Second
)

type Config struct {
	DatabaseURL                 string        `env:"DATABASE_URL" conf:"DATABASE_URL" json:"DATABASE_URL"`
	DatabaseMaxConnections      int           `env:"DATABASE_MAX_CONNECTIONS" conf:"DATABASE_MAX_CONNECTIONS" json:"DATABASE_MAX_CONNECTIONS"`
	DatabaseMigrationsDir       string        `env:"DATABASE_MIGRATIONS_DIR" conf:"DATABASE_MIGRATIONS_DIR" json:"DATABASE_MIGRATIONS_DIR"`
	DatabaseConnectMaxRetries   int           `env:"DATABASE_CONNECT_MAX_RETRIES" conf:"DATABASE_CONNECT_MAX_RETRIES" json:"DATABASE_CONNECT_MAX_RETRIES"`
	DatabaseConnectRetryTimeout time.Duration `env:"DATABASE_CONNECT_RETRY_TIMEOUT" conf:"DATABASE_CONNECT_RETRY_TIMEOUT" json:"DATABASE_CONNECT_RETRY_TIMEOUT"`
	DatabaseAcquireTimeout      time.Duration `env:"DATABASE_ACQUIRE_TIMEOUT" conf:"DATABASE_ACQUIRE_TIMEOUT" json:"DATABASE_ACQUIRE_TIMEOUT"`

	HTTPPort         string        `env:"HTTP_PORT" conf:"HTTP_PORT" json:"HTTP_PORT"`
	HTTPReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" conf:"HTTP_READ_TIMEOUT" json:"HTTP_READ_TIMEOUT"`
	HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" conf:"HTTP_WRITE_TIMEOUT" json:"HTTP_WRITE_TIMEOUT"`
	LogLevel         string        `env:"LOG_LEVEL" conf:"LOG_LEVEL" json:"LOG_LEVEL"`
	LogRequests      string        `env:"LOG_REQUESTS" conf:"LOG_REQUESTS" json:"LOG_REQUESTS"`
	EncryptionKey    string        `env:"ENCRYPTION_KEY" conf:"ENCRYPTION_KEY" json:"ENCRYPTION_KEY"`
	CookieExpiration time.Duration `env:"COOKIE_EXPIRATION" conf:"COOKIE_EXPIRATION" json:"COOKIE_EXPIRATION"`
}

func NewConfig() (*Config, error) {
	cfg := Config{
		HTTPPort:                    defaultHTTPPort,
		HTTPReadTimeout:             defaultHTTPReadTimeout,
		HTTPWriteTimeout:            defaultHTTPWriteTimeout,
		LogLevel:                    defaultLogLevel,
		LogRequests:                 defaultLogRequests,
		CookieExpiration:            defaultCookieExpiration,
		DatabaseMaxConnections:      defaultDatabaseMaxConnections,
		DatabaseMigrationsDir:       defaultDatabaseMigrationsDir,
		DatabaseConnectMaxRetries:   defaultDatabaseConnectMaxRetries,
		DatabaseConnectRetryTimeout: defaultDatabaseConnectRetryTimeout,
		DatabaseAcquireTimeout:      defaultAcquireTimeout,
	}

	_ = LoadConfigFile(&cfg, "config.conf")

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config new_config unable to parse env variables: %e", err)
	}

	if cfg.EncryptionKey == "" {
		return nil, fmt.Errorf("config is broken, encryption key is empty")
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config is broken, database url is empty")
	}

	return &cfg, nil
}

func LoadConfigFile(cfg *Config, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("config file not found %s", err)
	}

	configBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read config %s", err)
	}

	configString := string(configBytes)
	configs := strings.Split(configString, "\n")
	typeElem := reflect.TypeOf(cfg).Elem()
	elem := reflect.ValueOf(cfg).Elem()

	for _, v := range configs {
		s := strings.SplitN(v, "=", 2)
		if len(s) < 2 {
			continue
		}
		n := elem.NumField()
		for i := 0; i < n; i++ {
			field := typeElem.Field(i)
			tag := field.Tag.Get("conf")

			if tag != strings.TrimSpace(s[0]) {
				continue
			}

			fieldValue := elem.FieldByName(field.Name)

			fieldType := field.Type.String()

			val := strings.Join(strings.Fields(s[1]), " ")

			configValue, err := strconv.Unquote(val)
			if err != nil {
				configValue = val
			}

			configValue = strings.TrimSpace(configValue)

			switch fieldType {
			case "int":
				val, _ := strconv.ParseInt(configValue, 10, 0)
				fieldValue.Set(reflect.ValueOf(int(val)))
			case "string":
				fieldValue.Set(reflect.ValueOf(configValue))
			case "time.Duration":
				val, _ := time.ParseDuration(configValue)
				fieldValue.Set(reflect.ValueOf(val))
			}
		}
	}
	return nil
}

func ZeroLogLevel(l string) zerolog.Level {
	var logLevel zerolog.Level

	switch l {
	case "fatal":
		logLevel = zerolog.FatalLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "debug":
		logLevel = zerolog.DebugLevel

	default:
		logLevel = zerolog.ErrorLevel
	}

	return logLevel
}
