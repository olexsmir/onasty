package main

import "os"

type Config struct {
	AppURL        string
	NatsURL       string
	MailgunFrom   string
	MailgunDomain string
	MailgunAPIKey string

	LogLevel    string
	LogFormat   string
	LogShowLine bool

	MetricsEnabled bool
	MetricsPort    string
}

func NewConfig() *Config {
	return &Config{
		AppURL:         getenvOrDefault("APP_URL", ""),
		NatsURL:        getenvOrDefault("NATS_URL", ""),
		MailgunFrom:    getenvOrDefault("MAILGUN_FROM", ""),
		MailgunDomain:  getenvOrDefault("MAILGUN_DOMAIN", ""),
		MailgunAPIKey:  getenvOrDefault("MAILGUN_API_KEY", ""),
		LogLevel:       getenvOrDefault("LOG_LEVEL", "debug"),
		LogFormat:      getenvOrDefault("LOG_FORMAT", "json"),
		LogShowLine:    getenvOrDefault("LOG_SHOW_LINE", "true") == "true",
		MetricsPort:    getenvOrDefault("METRICS_PORT", ""),
		MetricsEnabled: getenvOrDefault("METRICS_ENABLED", "true") == "true",
	}
}

func getenvOrDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
