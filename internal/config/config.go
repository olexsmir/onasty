package config

import (
	"os"
)

type Config struct {
	AppEnv       string
	ServerPort   string
	PasswordSalt string

	LogLevel  string
	LogFormat string

	PostgresDSN string
}

func NewConfig() *Config {
	return &Config{
		AppEnv:      getenvOrDefault("APP_ENV", "debug"),
		ServerPort:  getenvOrDefault("SERVER_PORT", "3000"),
		PasswordSalt: getenvOrDefault("PASSWORD_SALT", ""),
		LogLevel:    getenvOrDefault("LOG_LEVEL", "debug"),
		LogFormat:   getenvOrDefault("LOG_FORMAT", "json"),
		PostgresDSN: getenvOrDefault("POSTGRESQL_DSN", ""),
	}
}

func (c *Config) IsDevMode() bool {
	return c.AppEnv == "debug" || c.AppEnv == "test"
}

func getenvOrDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
