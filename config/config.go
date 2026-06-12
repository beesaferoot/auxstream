package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBUrl         string `mapstructure:"DATABASE_URL"`   // postgres DSN, e.g. postgres://user:pass@host:5432/db?sslmode=disable
	SessionString string `mapstructure:"SESSION_STRING"` // secret key backing the gin cookie session store
	GinMode       string `mapstructure:"GIN_MODE"`       // "debug", "release", or "test"; "debug"/"development" select dev logging
	Addr          string `mapstructure:"ADDR"`           // host/interface to bind, blank means all interfaces
	Port          string `mapstructure:"PORT"`           // TCP port to listen on, digits only (no leading colon)
	RedisAddr     string `mapstructure:"REDIS_ADDRESS"`  // host:port of Redis, defaults to 127.0.0.1:6379
	FileStore     string `mapstructure:"FILE_STORE"`     // selects the upload backend (e.g. s3, cloudinary, local)
	S3bucket      string `mapstructure:"S3_BUCKET_ID"`   // S3 bucket name, required only when FileStore is s3
	CloudinaryURL string `mapstructure:"CLOUDINARY_URL"` // cloudinary://key:secret@cloud credential URL, required only for cloudinary
	JWTSecret     string `mapstructure:"JWT_SECRET"`     // HMAC signing key for issued JWTs; override the insecure default in prod
	// External catalog API credentials; blank disables the corresponding search source.
	YouTubeAPIKey      string `mapstructure:"YOUTUBE_API_KEY"`
	SoundCloudClientID string `mapstructure:"SOUNDCLOUD_CLIENT_ID"`
	MaxUploadBytes     int64  `mapstructure:"MAX_UPLOAD_BYTES"`  // per-file upload cap in bytes
	MaxRequestBytes    int64  `mapstructure:"MAX_REQUEST_BYTES"` // whole-request body cap in bytes, bounds bulk uploads
}

// LoadConfig reads an app.env file under path, falling back to matching
// environment variables, and applies the defaults below for optional settings.
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
