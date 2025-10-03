package internal

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Mode, Addr              string
	TrustedProxies, Origins []string
	DatabaseURL             string
	JWTSecretKey            []byte
	Pagination              int
	MaxMultipartMemory      int64
	AllowedImgMimetype      []string
	UploadPath              string
	MaxProductImagesAmount  int
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
			Mode:                   args["MODE"],
			Addr:                   args["ADDR"],
			TrustedProxies:         getStrSliceFromStr(args["TRUSTED_PROXIES"]),
			Origins:                getStrSliceFromStr(args["ORIGINS"]),
			DatabaseURL:            args["DATABASE_URL"],
			JWTSecretKey:           []byte(args["JWT_SECRET_KEY"]),
			Pagination:             getNumberFromStr(args["PAGINATION"]),
			MaxMultipartMemory:     int64(getNumberFromStr(args["MAX_MULTIPART_MEMORY"])),
			AllowedImgMimetype:     getStrSliceFromStr(args["ALLOWED_IMAGE_MIMETYPE"]),
			UploadPath:             strings.TrimSuffix(args["UPLOAD_PATH"], "/"),
			MaxProductImagesAmount: getNumberFromStr(args["MAX_PRODUCT_IMAGES_AMOUNT"]),
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

func getNumberFromStr(arg string) int {
	num, err := strconv.Atoi(arg)
	if err != nil {
		log.Panicf(`invalid arg number "%s", %s\n`, arg, err)
	}
	return num
}
