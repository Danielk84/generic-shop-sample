package internal

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Mode, Addr              string
	TrustedProxies, Origins []string
	DatabaseURL             string
	JWTSecretKey            []byte
}

var (
	DefaultConfig *Config
	once          sync.Once
)

func NewConfig() *Config {
	once.Do(func() {
		args, err := godotenv.Read(getDotEnvFilePath())
		if err != nil {
			log.Panicln("error on reading .env file", err)
		}

		DefaultConfig = &Config{
			Mode:           args["MODE"],
			Addr:           args["ADDR"],
			TrustedProxies: getStrSliceFromStr(args["TRUSTED_PROXIES"]),
			Origins:        getStrSliceFromStr(args["ORIGINS"]),
			DatabaseURL:    args["DATABASE_URL"],
			JWTSecretKey:   []byte(args["JWT_SECRET_KEY"]),
		}
	})
	return DefaultConfig
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
