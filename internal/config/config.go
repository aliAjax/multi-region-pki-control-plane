package config

import "os"

type Config struct {
	ListenAddr     string
	DevelopmentHSM bool
	Region         string
	DatabaseURL    string
	AuditKey       string
}

func Load() Config {
	c := Config{ListenAddr: ":8080", Region: "local", DatabaseURL: "postgres://localhost/pki", AuditKey: "dev-audit-key"}
	if v := os.Getenv("PKI_LISTEN_ADDR"); v != "" {
		c.ListenAddr = v
	}
	if os.Getenv("PKI_DEV_HSM") == "true" {
		c.DevelopmentHSM = true
	}
	if v := os.Getenv("PKI_REGION"); v != "" {
		c.Region = v
	}
	return c
}
