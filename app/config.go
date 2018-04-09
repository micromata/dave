package app

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

// Config represents the configuration of the server application.
type Config struct {
	Address string
	Port    string
	TLS     *TLS
}

// TLS allows specification of a certificate and private key file
type TLS struct {
	CertFile string
	KeyFile  string
}

// ParseConfig parses the application configuration an sets defaults.
func ParseConfig() Config {
	var cfg Config

	setDefaults()
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.swd")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal error parsing config file: %s", err))
	}

	if cfg.TLS != nil {
		if _, err := os.Stat(cfg.TLS.KeyFile); err != nil {
			log.Fatal(fmt.Errorf("TLS keyFile doesn't exist: %s", err))
		}
		if _, err := os.Stat(cfg.TLS.CertFile); err != nil {
			log.Fatal(fmt.Errorf("TLS certFile doesn't exist: %s", err))
		}
	}

	return cfg
}

// setDefaults adds some default values for the configuration
func setDefaults() {
	viper.SetDefault("Address", "127.0.0.1")
	viper.SetDefault("Port", "8000")
	viper.SetDefault("TLS", nil)
}
