package app

import (
	"github.com/spf13/viper"
	"fmt"
)

// Config represents the configuration of the server application.
type Config struct {
	Address string
	Port string
}

// ParseConfig parses the application configuration an sets defaults.
func ParseConfig() Config {
	var cfg Config

	setDefaults();

	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.swd")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(fmt.Errorf("Fatal error parsing config file: %s", err))
	}

	return cfg
}

// setDefaults adds some default values for the configuration
func setDefaults() {
	viper.SetDefault("Address", "127.0.0.1")
	viper.SetDefault("Port", "8000")
}
