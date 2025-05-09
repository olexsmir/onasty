package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv  string
	AppURL  string
	NatsURL string

	HTTPPort            int
	HTTPWriteTimeout    time.Duration
	HTTPReadTimeout     time.Duration
	HTTPHeaderMaxSizeMb int

	PostgresDSN      string
	PasswordSalt     string
	NotePasswordSalt string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	CacheUsersTTL time.Duration
	CacheNoteTTL  time.Duration

	JwtSigningKey      string
	JwtAccessTokenTTL  time.Duration
	JwtRefreshTokenTTL time.Duration

	GoogleClientID    string
	GoogleSecret      string
	GoogleRedirectURL string

	GitHubClientID    string
	GitHubSecret      string
	GitHubRedirectURL string

	VerificationTokenTTL time.Duration
	ResetPasswrdTokenTTL time.Duration

	MetricsEnabled bool
	MetricsPort    int

	LogLevel    string
	LogFormat   string
	LogShowLine bool

	RateLimiterRPS   int
	RateLimiterBurst int
	RateLimiterTTL   time.Duration
}

func NewConfig() *Config {
	return &Config{
		AppEnv:  getenvOrDefault("APP_ENV", "debug"),
		AppURL:  getenvOrDefault("APP_URL", ""),
		NatsURL: getenvOrDefault("NATS_URL", ""),

		HTTPPort:            mustGetenvOrDefaultInt("HTTP_PORT", 3000),
		HTTPWriteTimeout:    mustParseDuration(getenvOrDefault("HTTP_WRITE_TIMEOUT", "10s")),
		HTTPReadTimeout:     mustParseDuration(getenvOrDefault("HTTP_READ_TIMEOUT", "10s")),
		HTTPHeaderMaxSizeMb: mustGetenvOrDefaultInt("HTTP_HEADER_MAX_SIZE_MB", 1),

		PostgresDSN:      getenvOrDefault("POSTGRESQL_DSN", ""),
		PasswordSalt:     getenvOrDefault("PASSWORD_SALT", ""),
		NotePasswordSalt: getenvOrDefault("NOTE_PASSWORD_SALT", ""),

		RedisAddr:     getenvOrDefault("REDIS_ADDR", ""),
		RedisPassword: getenvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:       mustGetenvOrDefaultInt(getenvOrDefault("REDIS_DB", "0"), 0),

		CacheUsersTTL: mustParseDuration(getenvOrDefault("CACHE_USERS_TTL", "1h")),
		CacheNoteTTL:  mustParseDuration(getenvOrDefault("CACHE_NOTE_TTL", "1h")),

		JwtSigningKey: getenvOrDefault("JWT_SIGNING_KEY", ""),
		JwtAccessTokenTTL: mustParseDuration(
			getenvOrDefault("JWT_ACCESS_TOKEN_TTL", "15m"),
		),
		JwtRefreshTokenTTL: mustParseDuration(
			getenvOrDefault("JWT_REFRESH_TOKEN_TTL", "24h"),
		),

		GoogleClientID:    getenvOrDefault("GOOGLE_CLIENTID", ""),
		GoogleSecret:      getenvOrDefault("GOOGLE_SECRET", ""),
		GoogleRedirectURL: getenvOrDefault("GOOGLE_REDIRECTURL", ""),

		GitHubClientID:    getenvOrDefault("GITHUB_CLIENTID", ""),
		GitHubSecret:      getenvOrDefault("GITHUB_SECRET", ""),
		GitHubRedirectURL: getenvOrDefault("GITHUB_REDIRECTURL", ""),

		VerificationTokenTTL: mustParseDuration(getenvOrDefault("VERIFICATION_TOKEN_TTL", "24h")),
		ResetPasswrdTokenTTL: mustParseDuration(getenvOrDefault("RESET_PASSWORD_TOKEN_TTL", "1h")),

		MetricsPort:    mustGetenvOrDefaultInt("METRICS_PORT", 3001),
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
