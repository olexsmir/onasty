package config

import (
	"os"
	"time"
)

type Config struct {
	AppEnv       string
	ServerPort   string
	PasswordSalt string

	JwtSigningKey      string
	JwtAccessTokenTTL  time.Duration
	JwtRefreshTokenTTL time.Duration

	MailgunFrom         string
	MailgunDomain       string
	MailgunAPIKey       string
	VerficationTokenTTL time.Duration

	LogLevel  string
	LogFormat string

	PostgresDSN string
}

func NewConfig() *Config {
	return &Config{
		AppEnv:              getenvOrDefault("APP_ENV", "debug"),
		ServerPort:          getenvOrDefault("SERVER_PORT", "3000"),
		PasswordSalt:        getenvOrDefault("PASSWORD_SALT", ""),
		JwtSigningKey:       getenvOrDefault("JWT_SIGNING_KEY", ""),
		JwtAccessTokenTTL:   mustParseDuration(getenvOrDefault("JWT_ACCESS_TOKEN_TTL", "15m")),
		JwtRefreshTokenTTL:  mustParseDuration(getenvOrDefault("JWT_REFRESH_TOKEN_TTL", "15d")),
		MailgunFrom:         getenvOrDefault("MAILGUN_FROM", ""),
		MailgunDomain:       getenvOrDefault("MAILGUN_DOMAIN", ""),
		MailgunAPIKey:       getenvOrDefault("MAILGUN_API_KEY", ""),
		VerficationTokenTTL: mustParseDuration(getenvOrDefault("VERIFICATION_TOKEN_TTL", "24h")),
		LogLevel:            getenvOrDefault("LOG_LEVEL", "debug"),
		LogFormat:           getenvOrDefault("LOG_FORMAT", "json"),
		PostgresDSN:         getenvOrDefault("POSTGRESQL_DSN", ""),
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

func mustParseDuration(dur string) time.Duration {
	d, _ := time.ParseDuration(dur)
	return d
}
