package config

import (
	"os"
	"time"
)

type Config struct {
	AppEnv     string
	ServerPort string
	CorsOrigin string

	PasswordHashSalt string

	JWTSigningKey      string
	JWTAccessTokenTTL  time.Duration
	JWTRefreshTokenTTL time.Duration

	LogLevel  string
	LogFormat string

	PostgresUsername string
	PostgresPassword string
	PostgresHost     string
	PostgresPort     string
	PostgresDatabase string
}

func New() (*Config, error) {
	return &Config{
		AppEnv:     GetenvOrDefault("APP_ENV", "debug"),
		ServerPort: GetenvOrDefault("SERVER_PORT", "3000"),
		CorsOrigin: GetenvOrDefault("CORS_ORIGIN", "*"),

		PasswordHashSalt: GetenvOrDefault("PASSWORD_HASH_SALT", ""),

		JWTSigningKey:      GetenvOrDefault("JWT_SIGNING_KEY", "IT-HAS-TO-BE-SECRET"),
		JWTAccessTokenTTL:  MustParseDuration(GetenvOrDefault("JWT_ACCESS_TOKEN_TTL", "60m")),
		JWTRefreshTokenTTL: MustParseDuration(GetenvOrDefault("JWT_REFRESH_TOKEN_TTL", "15d")),

		LogLevel:  GetenvOrDefault("LOG_LEVEL", "debug"),
		LogFormat: GetenvOrDefault("LOG_FORMAT", "json"),

		PostgresUsername: GetenvOrDefault("POSTGRES_USERNAME", ""),
		PostgresPassword: GetenvOrDefault("POSTGRES_PASSWORD", ""),
		PostgresHost:     GetenvOrDefault("POSTGRES_HOST", ""),
		PostgresPort:     GetenvOrDefault("POSTGRES_PORT", ""),
		PostgresDatabase: GetenvOrDefault("POSTGRES_DATABASE", ""),
	}, nil
}

func GetenvOrDefault(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func MustParseDuration(dur string) time.Duration {
	d, _ := time.ParseDuration(dur)
	return d
}
