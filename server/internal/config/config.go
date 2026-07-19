package config

import (
	"fmt"
	"generic-shop-sample/internal"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	App          AppConfig          `yaml:"app"`
	Auth         AuthConfig         `yaml:"auth"`
	EmailBroker  EmailBrokerConfig  `yaml:"email_broker"`
	ProductImage ProductImageConfig `yaml:"product_image"`
	Payment      PaymentConfig      `yaml:"payment"`
	FileUpload   FileUploadConfig   `yaml:"file_upload"`

	RequestLoggerFilepath string   `yaml:"request_logger_filepath" binding:"required,filepath"`
	Origins               []string `yaml:"origins" binding:"required,dive,required,origin"`

	DatabaseURL string `yaml:"database_url" binding:"required,url"`
	CacheURL    string `yaml:"cache_url" binding:"required,url"`

	Pagination int `yaml:"pagination" binding:"required,min=0"`

	JWTSecretKey string `yaml:"jwt_secret_key" binding:"required,ascii"`
}

type AppConfig struct {
	Mode               string   `yaml:"mode" binding:"required,oneof=debug release"`
	MaxMultipartMemory int64    `yaml:"max_multipart_memory" binding:"required,min=0"`
	TrustedProxies     []string `yaml:"trusted_proxies" binding:"required,dive,required,ip"`
	AppLoggerFilepath  string   `yaml:"app_logger_filepath" binding:"required,filepath"`
	Addr               string   `yaml:"addr" binding:"required"`
}

type AuthConfig struct {
	AuthExpiration int `yaml:"auth_expiration" binding:"required,min=0"`
}

type EmailBrokerConfig struct {
	FromEmail    string `yaml:"from_email" binding:"required,email"`
	SMTPHost     string `yaml:"smtp_host" binding:"required,hostname"`
	SMTPPort     int    `yaml:"smtp_port" binding:"required,min=1,max=65535"`
	SMTPPassword string `yaml:"smtp_password" binding:"omitempty"`
}

type ProductImageConfig struct {
	MaxProductImagesAmount int `yaml:"max_product_images_amount" binding:"required,min=5"`
}

type PaymentConfig struct {
	ZPMerchantID       string `yaml:"zp_merchant_id" binding:"required"`
	PaymentCallbackURL string `yaml:"payment_callback_url" binding:"required,http_url"`
}

type FileUploadConfig struct {
	UploadPath         string   `yaml:"upload_path" binding:"required,dirpath"`
	AllowedImgMimetype []string `yaml:"allowed_image_mimetype" binding:"required,dive,required"`
}

func (c *Config) ReadFile(fp string) error {
	if !filepath.IsAbs(fp) {
		return fmt.Errorf("%s: empty or not absolute file path", os.ErrInvalid)
	}
	yamlFile, err := os.ReadFile(fp)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(yamlFile, c); err != nil {
		return err
	}

	return nil
}

func (c *Config) Validate() error {
	if err := internal.GetValidator().Struct(c); err != nil {
		return err
	}
	return nil
}

var defualtConfig *Config

func NewConfig(fp string) Config {
	if defualtConfig == nil {
		defualtConfig = &Config{}
	}
	if err := defualtConfig.ReadFile(fp); err != nil {
		panic(err)
	}
	if err := defualtConfig.Validate(); err != nil {
		panic(err)
	}
	return *defualtConfig
}

func GetConfig() *Config {
	if defualtConfig == nil {
		panic("nil default config")
	}
	return defualtConfig
}
