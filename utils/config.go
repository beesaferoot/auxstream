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
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
