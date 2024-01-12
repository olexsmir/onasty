package config

import "os"

type Config struct {
	AppEnv     string
	ServerPort string
	CorsOrigin string

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
