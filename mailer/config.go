package main

import "os"

type Config struct {
	AppURL        string
	NatsURL       string
	MailgunFrom   string
	MailgunDomain string
	MailgunAPIKey string
}

func NewConfig() *Config {
	return &Config{
		AppURL:        getenvOrDefault("APP_URL", ""),
		NatsURL:       getenvOrDefault("NATS_URL", ""),
		MailgunFrom:   getenvOrDefault("MAILGUN_FROM", ""),
		MailgunDomain: getenvOrDefault("MAILGUN_DOMAIN", ""),
		MailgunAPIKey: getenvOrDefault("MAILGUN_API_KEY", ""),
	}
}

func getenvOrDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
