package app

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Mode, Addr              string
	TrustedProxies, Origins []string
}

func NewAppConfig() *AppConfig {
	if err := godotenv.Load(getDotEnvFilePath()); err != nil {
		log.Panicln("error on loading .env file", err)
	}

	return &AppConfig{
		Mode:           os.Getenv("MODE"),
		Addr:           os.Getenv("ADDR"),
		TrustedProxies: getStrSliceFromStr(os.Getenv("TRUSTED_PROXIES")),
		Origins:        getStrSliceFromStr(os.Getenv("ORIGINS")),
	}
}

func getDotEnvFilePath() string {
	if env := os.Getenv("DOTENV_PATH"); env != "" {
		return env
	}
	return ".env"
}

func getStrSliceFromStr(s string) []string {
	return strings.Split(strings.ReplaceAll(s, " ", ""), ",")
}
