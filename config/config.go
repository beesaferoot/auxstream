package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBUrl              string `mapstructure:"DATABASE_URL"`
	SessionString      string `mapstructure:"SESSION_STRING"`
	GinMode            string `mapstructure:"GIN_MODE"`
	Addr               string `mapstructure:"ADDR"`
	Port               string `mapstructure:"PORT"`
	RedisAddr          string `mapstructure:"REDIS_ADDRESS"`
	FileStore          string `mapstructure:"FILE_STORE"`
	S3bucket           string `mapstructure:"S3_BUCKET_ID"`
	CloudinaryURL      string `mapstructure:"CLOUDINARY_URL"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.SetDefault("RedisAddr", "127.0.0.1:6379")
	viper.SetDefault("S3_BUCKET_ID", "")
	viper.SetDefault("CLOUDINARY_URL", "")
	viper.SetDefault("JWT_SECRET", "your-secret-key")
	viper.SetDefault("GOOGLE_CLIENT_ID", "")
	viper.SetDefault("GOOGLE_CLIENT_SECRET", "")
	viper.SetDefault("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
