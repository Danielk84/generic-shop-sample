package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

type Option struct {
	Mode                   string   `yaml:"mode" binding:"required,oneof=debug release"`
	Addr                   string   `yaml:"addr" binding:"required"`
	TrustedProxies         []string `yaml:"trusted_proxies" binding:"required,dive,required,ip"`
	Origins                []string `yaml:"origins" binding:"required,dive,required,origin"`
	DatabaseURL            string   `yaml:"database_url" binding:"required,url"`
	CacheURL               string   `yaml:"cache_url" binding:"required,url"`
	JWTSecretKey           string   `yaml:"jwt_secret_key" binding:"required,ascii"`
	Pagination             int      `yaml:"pagination" binding:"required,min=0"`
	MaxMultipartMemory     int64    `yaml:"max_multipart_memory" binding:"required,min=0"`
	AllowedImgMimetype     []string `yaml:"allowed_img_mimetype" binding:"dive,required,mimetype=image/*"`
	UploadPath             string   `yaml:"upload_path" binding:"required,dirpath"`
	MaxProductImagesAmount int      `yaml:"max_product_images_amount" binding:"required,min=5"`
	AuthExpiration         int      `yaml:"auth_expiration" binding:"required,min=0"`
	RequestLoggerFilepath  string   `yaml:"request_logger_filepath" binding:"required,filepath"`
	AppLoggerFilepath      string   `yaml:"app_logger_filepath" binding:"required,filepath"`
	ZPMerchantID           string   `yaml:"zp_merchant_id" binding:"required"`
	PaymentCallbackURL     string   `yaml:"payment_callback_url" binding:"required,http_url"`
	FromEmail              string   `yaml:"from_email" binding:"required,email"`
	SMTPHost               string   `yaml:"smtp_host" binding:"required,hostname"`
	SMTPPort               int      `yaml:"smtp_port" binding:"required,min=1,max=65535"`
	SMTPPassword           string   `yaml:"smtp_password" binding:"omitempty"`
}

type Config struct {
	Opt Option `yaml:"opt"`
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
	if err := GetValidator().Struct(c.Opt); err != nil {
		return err
	}
	return nil
}

var defualtConfig *Config

func NewConfig(fp string) *Config {
	if defualtConfig == nil {
		defualtConfig = &Config{}
	}
	if err := defualtConfig.ReadFile(fp); err != nil {
		panic(err)
	}
	if err := defualtConfig.Validate(); err != nil {
		panic(err)
	}
	return defualtConfig
}

func GetConfig() *Config {
	if defualtConfig == nil {
		panic("nil default config")
	}
	return defualtConfig
}
