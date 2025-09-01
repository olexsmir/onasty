package main

import (
	"os"
	"strconv"
	"sync"
)

var (
	configInstance *Config
	once           sync.Once
)

type Config struct {
	AppURL      string
	FrontendURL string

	NatsURL       string
	MailgunFrom   string
	MailgunDomain string
	MailgunAPIKey string

	LogLevel    string
	LogFormat   string
	LogShowLine bool

	MetricsEnabled bool
	MetricsPort    int
}

func NewConfig() *Config {
	once.Do(func() {
		configInstance = &Config{
			AppURL:         getenvOrDefault("APP_URL", ""),
			FrontendURL:    getenvOrDefault("FRONTEND_URL", ""),
			NatsURL:        getenvOrDefault("NATS_URL", ""),
			MailgunFrom:    getenvOrDefault("MAILGUN_FROM", ""),
			MailgunDomain:  getenvOrDefault("MAILGUN_DOMAIN", ""),
			MailgunAPIKey:  getenvOrDefault("MAILGUN_API_KEY", ""),
			LogLevel:       getenvOrDefault("LOG_LEVEL", "debug"),
			LogFormat:      getenvOrDefault("LOG_FORMAT", "json"),
			LogShowLine:    getenvOrDefault("LOG_SHOW_LINE", "true") == "true",
			MetricsPort:    mustGetenvOrDefaultInt("METRICS_PORT", 8001),
			MetricsEnabled: getenvOrDefault("METRICS_ENABLED", "true") == "true",
		}
	})
	return configInstance
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
