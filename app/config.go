package app

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

// Config represents the configuration of the server application.
type Config struct {
	Address string
	Port    string
	Prefix  string
	Dir     string
	TLS     *TLS
	Users   map[string]*UserInfo
}

// TLS allows specification of a certificate and private key file.
type TLS struct {
	CertFile string
	KeyFile  string
}

// UserInfo allows storing of a password and user directory.
type UserInfo struct {
	Password string
}

// ParseConfig parses the application configuration an sets defaults.
func ParseConfig() *Config {
	var cfg = &Config{}

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

	viper.WatchConfig()
	viper.OnConfigChange(cfg.updateConfig)

	cfg.ensureUserDirs()

	return cfg
}

// setDefaults adds some default values for the configuration
func setDefaults() {
	viper.SetDefault("Address", "127.0.0.1")
	viper.SetDefault("Port", "8000")
	viper.SetDefault("Prefix", "")
	viper.SetDefault("Dir", "/tmp")
	viper.SetDefault("TLS", nil)
}

func (cfg *Config) updateConfig(e fsnotify.Event) {
	fmt.Println("Config file changed:", e.Name)

	file, err := os.Open(e.Name)
	if err != nil {
		fmt.Println("Error reloading config", e.Name)
	}

	var updatedCfg = &Config{}
	viper.ReadConfig(file)
	viper.Unmarshal(&updatedCfg)

	for username := range cfg.Users {
		if updatedCfg.Users[username] == nil {
			fmt.Printf("Removed User from configuration: %s\n", username)
			cfg.Users[username] = nil
		}
	}

	for username, v := range updatedCfg.Users {
		if cfg.Users[username] == nil {
			fmt.Printf("Added User to configuration: %s\n", username)
			cfg.Users[username] = v
		} else {
			if cfg.Users[username].Password != v.Password {
				fmt.Printf("Updated password of user: %s\n", username)
				cfg.Users[username].Password = v.Password
			}
		}
	}

	cfg.ensureUserDirs()
}

func (cfg *Config) ensureUserDirs() {
	if _, err := os.Stat(cfg.Dir); os.IsNotExist(err) {
		os.Mkdir(cfg.Dir, os.ModePerm)
		fmt.Printf("Created base dir: %s\n", cfg.Dir)
	}

	for username := range cfg.Users {
		path := filepath.Join(cfg.Dir, username)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, os.ModePerm)
			fmt.Printf("Created user dir: %s\n", path)
		}
	}
}
