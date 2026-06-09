package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Mode                   string        `yaml:"mode" binding:"required,oneof=debug release"`
	Addr                   string        `yaml:"addr" binding:"required"`
	TrustedProxies         []string      `yaml:"trusted_proxies" binding:",required,ip"`
	Origins                []string      `yaml:"origins" binding:"required,origin"`
	DatabaseURL            string        `yaml:"basebase_url" binding:"required,url"`
	CacheURL               string        `yaml:"cache_url" binding:"required,url"`
	JWTSecretKey           []byte        `yaml:"jwt_secret_key" binding:"required,ascii"`
	Pagination             int           `yaml:"pagination" binding:"required,min=0"`
	MaxMultipartMemory     int64         `yaml:"max_multipart_memory" binding:"required,min=0"`
	AllowedImgMimetype     []string      `yaml:"allowed_img_mimetype" binding:"required,mimetype=image/*"`
	UploadPath             string        `yaml:"upload_path" binding:"required,filepath"`
	MaxProductImagesAmount int           `yaml:"max_product_image_amount" binding:"required,min=5"`
	AuthExpiration         time.Duration `yaml:"auth_expiration" binding:"required,min=15m"`
	RequestLoggerFilepath  string        `yaml:"request_logger_filepath" binding:"required,filepth"`
	AppLoggerFilepath      string        `yaml:"app_logger_filepath" binding:"required,filepath"`
	ZPMerchantID           string        `yaml:"zp_merchant_id" binding:"required"`
	PaymentCallbackURL     string        `yaml:"payment_callback_url" binding:"required,http_url"`
	FromEmail              string        `yaml:"from_email" binding:"required,email"`
	SMTPHost               string        `yaml:"smtp_host" binding:"required,hostname"`
	SMTPPort               int           `yaml:"smtp_port" binding:"required,port"`
	SMTPPassword           string        `yaml:"smtp_password" binding:"omitempty"`
}

type ConfigFile struct {
	Conf Config `yaml:"conf"`
}

func (c *ConfigFile) ReadFile(fp string) error {
	if filepath.IsAbs(fp) {
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

func (c *ConfigFile) Validate() error {
	if err := GetValidator().Struct(c.Conf); err != nil {
		return err
	}
	return nil
}

var defualtConfig *ConfigFile

func NewConfig(fp string) Config {
	if defualtConfig == nil {
		defualtConfig = &ConfigFile{}
	}
	if err := defualtConfig.ReadFile(fp); err != nil {
		panic(err)
	}
	if err := defualtConfig.Validate(); err != nil {
		panic(err)
	}
	return defualtConfig.Conf
}

func GetConfig() Config {
	if defualtConfig == nil {
		panic("nil default config")
	}
	return defualtConfig.Conf
}
