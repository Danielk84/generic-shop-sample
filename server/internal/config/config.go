package config

import (
	"fmt"
	"generic-shop-sample/internal"
	"log"
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
	FileStore    FileStorageConfig  `yaml:"file_upload"`

	RequestLoggerFilepath string `yaml:"request_logger_filepath" binding:"required,filepath"`

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

type FileStorageConfig struct {
	AllowedImgMimetype []string    `yaml:"allowed_image_mimetype" binding:"required,dive,required"`
	AwsS3              AwsS3Config `yaml:"aws_s3" binding:"required"`
}

type AwsS3Config struct {
	Key      string `yaml:"key" binding:"required"`
	Secret   string `yaml:"secret" binding:"required"`
	Region   string `yaml:"region" binding:"required"`
	Endpoint string `yaml:"endpoint" binding:"required,http_url"`
}

func (c *Config) ReadFile(fp string) error {
	if !filepath.IsAbs(fp) {
		return fmt.Errorf("%s: empty or not absolute file path", os.ErrInvalid)
	}
	yamlFile, err := os.ReadFile(fp)
	if err != nil {
		return err
	}
	replaced := os.ExpandEnv(string(yamlFile))
	if err = yaml.Unmarshal([]byte(replaced), c); err != nil {
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

func NewConfig(fp string) Config {
	c := &Config{}

	if err := c.ReadFile(fp); err != nil {
		panic(err)
	}
	if err := c.Validate(); err != nil {
		log.Printf(`config: "%+v"`, c)
		panic(err)
	}
	return *c
}
