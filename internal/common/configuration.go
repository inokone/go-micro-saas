package common

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const ConfigFolder = "/etc/microsaas/"

// WebConfig is a configuration of the web application.
type WebConfig struct {
	Port int `mapstructure:"PORT"`
}

// RDBConfig is a configuration of the relational database.
type RDBConfig struct {
	Host            string        `mapstructure:"DB_HOST"`
	Port            int           `mapstructure:"DB_PORT"`
	DBName          string        `mapstructure:"DB_NAME"`
	Username        string        `mapstructure:"DB_USER"`
	Password        string        `mapstructure:"DB_PASS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONN"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONN"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_LIFETIME"`
	SSLMode         string        `mapstructure:"DB_SSL_MODE"`
	SSLCert         string        `mapstructure:"DB_SSL_CERT"`
}

func (c RDBConfig) String() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", c.Username, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// AuthConfig is a configuration of the authentication.
type AuthConfig struct {
	JWTSecret          string `mapstructure:"JWT_SIGN_SECRET"`
	JWTExp             int    `mapstructure:"JWT_EXPIRATION_HOURS"`
	JWTSecure          bool   `mapstructure:"JWT_COOKIE_SECURE"`
	TLSCert            string `mapstructure:"TLS_CERT_PATH"`
	TLSKey             string `mapstructure:"TLS_KEY_PATH"`
	FrontendRoot       string `mapstructure:"FRONTEND_ROOT"`
	BackendRoot        string `mapstructure:"BACKEND_ROOT"`
	RecaptchaAppCreds  string `mapstructure:"GOOGLE_APPLICATION_CREDENTIALS"`
	RecaptchaProjectID string `mapstructure:"GOOGLE_PROJECT_ID"`
	RecaptchaKey       string `mapstructure:"GOOGLE_RECAPTCHA_KEY"`
	GoogleKey          string `mapstructure:"GOOGLE_AUTH_KEY"`
	GoogleSecret       string `mapstructure:"GOOGLE_AUTH_SECRET"`
	FacebookKey        string `mapstructure:"FACEBOOK_AUTH_KEY"`
	FacebookSecret     string `mapstructure:"FACEBOOK_AUTH_SECRET"`
}

// MailConfig is a configuration of e-mail massaging.
type MailConfig struct {
	ApplicationName string `mapstructure:"APPLICATION_NAME"`
	NoReplyAddress  string `mapstructure:"MAIL_NO_REPLY_ADDRESS"`
	SMTPAddress     string `mapstructure:"MAIL_SMTP_ADDRESS"`
	SMTPUser        string `mapstructure:"MAIL_SMTP_USER"`
	SMTPPassword    string `mapstructure:"MAIL_SMTP_PASSWORD"`
	SMTPPort        int    `mapstructure:"MAIL_SMTP_PORT"`
}

// LogConfig is a configuration of the logging.
type LogConfig struct {
	LogLevel  string `mapstructure:"LOG_LEVEL"`
	PrettyLog bool   `mapstructure:"PRETTY_LOG"`
}

type AnalyticsConfig struct {
	StatsigServerKey string `mapstructure:"STATSIG_SERVER_SECRET_KEY"`
}

// AppConfig is the holder of all configurations for the application
type AppConfig struct {
	DB        *RDBConfig
	Auth      *AuthConfig
	Log       *LogConfig
	Mail      *MailConfig
	Web       *WebConfig
	Analytics *AnalyticsConfig
	Path      string
}

func configPaths(path string) [4]string {
	return [4]string{path, ".", ConfigFolder, "$HOME/.microsaas"}
}

func (c AppConfig) PathFor(file string) string {
	for _, confPath := range configPaths(c.Path) {
		target := filepath.Join(confPath, file)
		if _, err := os.Stat(target); err == nil {
			return target
		}
	}
	panic(fmt.Sprintf("Configuration file %s not found", file))
}

// loadConfig is a function loading the configuration from app.env file in the runtime directory or environment variables.
// As a fallback `$HOME/.microsaas` directory also can be used for the .env file.
func loadConfig(path string) (*AppConfig, error) {
	var db RDBConfig
	var au AuthConfig
	var lg LogConfig
	var ml MailConfig
	var wb WebConfig
	var an AnalyticsConfig

	for _, confPath := range configPaths(path) {
		viper.AddConfigPath(confPath)
	}
	viper.SetConfigType("env")
	viper.SetConfigName("app")
	viper.SetDefault("JWT_COOKIE_SECURE", true)
	viper.SetDefault("JWT_EXPIRATION_HOURS", 24)
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("IMG_STORE_USE_PRESIGNED", false)
	viper.SetDefault("IMG_STORE_PRESIGNED_TTL", 300)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	for _, config := range [6]any{&wb, &db, &au, &lg, &ml, &an} {
		if err = viper.Unmarshal(config); err != nil {
			return nil, err
		}
	}
	return &AppConfig{
		DB:        &db,
		Auth:      &au,
		Log:       &lg,
		Mail:      &ml,
		Web:       &wb,
		Analytics: &an,
		Path:      path,
	}, nil
}
