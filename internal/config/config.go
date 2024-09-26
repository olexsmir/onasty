package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	AppEnv       string
	AppURL       string
	ServerPort   string
	PostgresDSN  string
	PasswordSalt string

	JwtSigningKey      string
	JwtAccessTokenTTL  time.Duration
	JwtRefreshTokenTTL time.Duration

	MailgunFrom          string
	MailgunDomain        string
	MailgunAPIKey        string
	VerificationTokenTTL time.Duration

	MetricsEnabled bool
	MetricsPort    string

	LogLevel    string
	LogFormat   string
	LogShowLine bool
}

func NewConfig() *Config {
	return &Config{
		AppEnv:       getenvOrDefault("APP_ENV", "debug"),
		AppURL:       getenvOrDefault("APP_URL", ""),
		ServerPort:   getenvOrDefault("SERVER_PORT", "3000"),
		PostgresDSN:  getenvOrDefault("POSTGRESQL_DSN", ""),
		PasswordSalt: getenvOrDefault("PASSWORD_SALT", ""),

		JwtSigningKey: getenvOrDefault("JWT_SIGNING_KEY", ""),
		JwtAccessTokenTTL: mustParseDurationOrPanic(
			getenvOrDefault("JWT_ACCESS_TOKEN_TTL", "15m"),
		),
		JwtRefreshTokenTTL: mustParseDurationOrPanic(
			getenvOrDefault("JWT_REFRESH_TOKEN_TTL", "24h"),
		),

		MailgunFrom:   getenvOrDefault("MAILGUN_FROM", ""),
		MailgunDomain: getenvOrDefault("MAILGUN_DOMAIN", ""),
		MailgunAPIKey: getenvOrDefault("MAILGUN_API_KEY", ""),
		VerificationTokenTTL: mustParseDurationOrPanic(
			getenvOrDefault("VERIFICATION_TOKEN_TTL", "24h"),
		),

		MetricsPort:    getenvOrDefault("METRICS_PORT", "3001"),
		MetricsEnabled: getenvOrDefault("METRICS_ENABLED", "true") == "true",

		LogLevel:    getenvOrDefault("LOG_LEVEL", "debug"),
		LogFormat:   getenvOrDefault("LOG_FORMAT", "json"),
		LogShowLine: getenvOrDefault("LOG_SHOW_LINE", "true") == "true",
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

func mustParseDurationOrPanic(dur string) time.Duration {
	d, err := time.ParseDuration(dur)
	if err != nil {
		panic(errors.Join(errors.New("cannot time.ParseDuration"), err)) //nolint:err113
	}

	return d
}
