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
	YouTubeAPIKey      string `mapstructure:"YOUTUBE_API_KEY"`
	SoundCloudClientID string `mapstructure:"SOUNDCLOUD_CLIENT_ID"`
	MaxUploadBytes     int64  `mapstructure:"MAX_UPLOAD_BYTES"`
	MaxRequestBytes    int64  `mapstructure:"MAX_REQUEST_BYTES"`
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
	viper.SetDefault("YOUTUBE_API_KEY", "")
	viper.SetDefault("SOUNDCLOUD_CLIENT_ID", "")
	viper.SetDefault("MAX_UPLOAD_BYTES", 5<<20)   // 5 MiB per audio file
	viper.SetDefault("MAX_REQUEST_BYTES", 50<<20) // 50 MiB per request (bulk uploads); proxied upload buffers in memory

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
