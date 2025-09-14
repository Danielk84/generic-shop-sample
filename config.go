package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Mode, Addr     string
	TrustedProxies []string
}

func NewAppConfig() *AppConfig {
	if err := godotenv.Load(".env"); err != nil {
		log.Panicln("error on loading .env file", err)
	}

	trustedProxies := strings.ReplaceAll(os.Getenv("TRUSTED_PROXIES"), " ", "")

	return &AppConfig{
		Mode:           os.Getenv("MODE"),
		Addr:           os.Getenv("ADDR"),
		TrustedProxies: strings.Split(trustedProxies, ","),
	}
}
