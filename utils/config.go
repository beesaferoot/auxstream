package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBUrl         string `mapstructure:"DATABASE_URL"`
	SessionString string `mapstructure:"SESSION_STRING"`
	GinMode       string `mapstructure:"GIN_MODE"`
	Addr          string `mapstructure:"ADDR"`
	Port          string `mapstructure:"PORT"`
	RedisAddr     string `mapstructure:"REDIS_ADDRESS"`
	FileStore     string `mapstructure:"FILE_STORE"`
	S3bucket      string `mapstructure:"S3_BUCKET_ID"`
	CloudinaryURL string `mapstructure:"CLOUDINARY_URL"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.SetDefault("RedisAddr", "127.0.0.1:6379")
	viper.SetDefault("S3_BUCKET_ID", "")
	viper.SetDefault("CLOUDINARY_URL", "")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
