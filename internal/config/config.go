package config

import (
	"errors"
	"os"
	"strconv"
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

	RateLimiterRPS   int
	RateLimiterBurst int
	RateLimiterTTL   time.Duration
}

func NewConfig() *Config {
	return &Config{
		AppEnv:       getenvOrDefault("APP_ENV", "debug"),
		AppURL:       getenvOrDefault("APP_URL", ""),
		ServerPort:   getenvOrDefault("SERVER_PORT", "3000"),
		PostgresDSN:  getenvOrDefault("POSTGRESQL_DSN", ""),
		PasswordSalt: getenvOrDefault("PASSWORD_SALT", ""),

		JwtSigningKey: getenvOrDefault("JWT_SIGNING_KEY", ""),
		JwtAccessTokenTTL: mustParseDuration(
			getenvOrDefault("JWT_ACCESS_TOKEN_TTL", "15m"),
		),
		JwtRefreshTokenTTL: mustParseDuration(
			getenvOrDefault("JWT_REFRESH_TOKEN_TTL", "24h"),
		),

		MailgunFrom:   getenvOrDefault("MAILGUN_FROM", ""),
		MailgunDomain: getenvOrDefault("MAILGUN_DOMAIN", ""),
		MailgunAPIKey: getenvOrDefault("MAILGUN_API_KEY", ""),
		VerificationTokenTTL: mustParseDuration(
			getenvOrDefault("VERIFICATION_TOKEN_TTL", "24h"),
		),

		MetricsPort:    getenvOrDefault("METRICS_PORT", "3001"),
		MetricsEnabled: getenvOrDefault("METRICS_ENABLED", "true") == "true",

		LogLevel:    getenvOrDefault("LOG_LEVEL", "debug"),
		LogFormat:   getenvOrDefault("LOG_FORMAT", "json"),
		LogShowLine: getenvOrDefault("LOG_SHOW_LINE", "true") == "true",

		RateLimiterRPS:   mustGetenvOrDefaultInt("RATELIMITER_RPS", 100),
		RateLimiterBurst: mustGetenvOrDefaultInt("RATELIMITER_BURST", 10),
		RateLimiterTTL:   mustParseDuration(getenvOrDefault("RATELIMITER_TTL", "1m")),
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

func mustGetenvOrDefaultInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		r, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		return r
	}
	return def
}

func mustParseDuration(dur string) time.Duration {
	d, err := time.ParseDuration(dur)
	if err != nil {
		panic(errors.Join(errors.New("cannot time.ParseDuration"), err)) //nolint:err113
	}

	return d
}
