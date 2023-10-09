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
	defaultHTTPPort         = "8080"
	defaultHTTPReadTimeout  = 5 * time.Second
	defaultHTTPWriteTimeout = 5 * time.Second
	defaultLogLevel         = "error"
	defaultCookieExpiration = 24 * time.Hour
)

type Config struct {
	DatabaseURL      string        `env:"DATABASE_URL" conf:"DATABASE_URL"`
	HTTPPort         string        `env:"HTTP_PORT" conf:"HTTP_PORT"`
	HTTPReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" conf:"HTTP_READ_TIMEOUT"`
	HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" conf:"HTTP_WRITE_TIMEOUT"`
	LogLevel         string        `env:"LOG_LEVEL" conf:"LOG_LEVEL"`
	LogRequests      string        `env:"LOG_REQUESTS" conf:"LOG_REQUESTS"`
	EncryptionKey    string        `env:"ENCRYPTION_KEY" conf:"ENCRYPTION_KEY"`
	CookieExpiration time.Duration `env:"COOKIE_EXPIRATION" conf:"COOKIE_EXPIRATION"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config new_config unable to parse env variables: %e", err)
	}

	fileConfig, _ := LoadConfigFile("config.conf")

	if cfg.HTTPPort == "" {
		if fileConfig != nil && fileConfig.HTTPPort != "" {
			cfg.HTTPPort = fileConfig.HTTPPort
		} else {
			cfg.HTTPPort = defaultHTTPPort
		}
	}

	if cfg.HTTPReadTimeout == 0 {
		if fileConfig != nil && fileConfig.HTTPReadTimeout != 0 {
			cfg.HTTPReadTimeout = fileConfig.HTTPReadTimeout
		} else {
			cfg.HTTPReadTimeout = defaultHTTPReadTimeout
		}
	}

	if cfg.HTTPWriteTimeout == 0 {
		if fileConfig != nil && fileConfig.HTTPWriteTimeout != 0 {
			cfg.HTTPWriteTimeout = fileConfig.HTTPWriteTimeout
		} else {
			cfg.HTTPWriteTimeout = defaultHTTPWriteTimeout
		}
	}

	if cfg.LogLevel == "" {
		if fileConfig != nil && fileConfig.LogLevel != "" {
			cfg.LogLevel = fileConfig.LogLevel
		} else {
			cfg.LogLevel = defaultLogLevel
		}
	}

	if cfg.LogRequests == "" {
		if fileConfig != nil && fileConfig.LogRequests != "" {
			cfg.LogRequests = fileConfig.LogRequests
		} else {
			cfg.LogRequests = "false"
		}
	}

	if cfg.EncryptionKey == "" {
		if fileConfig != nil && fileConfig.EncryptionKey != "" {
			cfg.EncryptionKey = fileConfig.EncryptionKey
		} else {
			return nil, fmt.Errorf("config is broken, encrypt key is empty")
		}
	}

	if cfg.CookieExpiration == 0 {
		if fileConfig != nil && fileConfig.CookieExpiration != 0 {
			cfg.CookieExpiration = fileConfig.CookieExpiration
		} else {
			cfg.CookieExpiration = defaultCookieExpiration
		}
	}

	return &cfg, nil
}

func LoadConfigFile(filepath string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("config file not found %s", err)
	}

	configBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s", err)
	}

	cfg := Config{}

	configString := string(configBytes)
	configs := strings.Split(configString, "\n")
	typeElem := reflect.TypeOf(&cfg).Elem()
	elem := reflect.ValueOf(&cfg).Elem()

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
	return &cfg, nil
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
