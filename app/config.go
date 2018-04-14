package app

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	Log     Logging
	Users   map[string]*UserInfo
}

// Logging allows definition for logging each CRUD method.
type Logging struct {
	Create bool
	Read   bool
	Update bool
	Delete bool
}

// TLS allows specification of a certificate and private key file.
type TLS struct {
	CertFile string
	KeyFile  string
}

// UserInfo allows storing of a password and user directory.
type UserInfo struct {
	Password string
	Subdir   *string
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
	viper.SetDefault("Log.Create", false)
	viper.SetDefault("Log.Read", false)
	viper.SetDefault("Log.Update", false)
	viper.SetDefault("Log.Delete", false)
}

func (cfg *Config) updateConfig(e fsnotify.Event) {
	log.WithField("path", e.Name).Info("Config file changed")

	file, err := os.Open(e.Name)
	if err != nil {
		log.WithField("path", e.Name).Warn("Error reloading config")
	}

	var updatedCfg = &Config{}
	viper.ReadConfig(file)
	viper.Unmarshal(&updatedCfg)

	for username := range cfg.Users {
		if updatedCfg.Users[username] == nil {
			log.WithField("user", username).Info("Removed User from configuration")
			cfg.Users[username] = nil
		}
	}

	for username, v := range updatedCfg.Users {
		if cfg.Users[username] == nil {
			log.WithField("user", username).Info("Added User to configuration")
			cfg.Users[username] = v
		} else {
			if cfg.Users[username].Password != v.Password {
				log.WithField("user", username).Info("Updated password of user")
				cfg.Users[username].Password = v.Password
			}
		}
	}

	cfg.ensureUserDirs()

	if cfg.Log.Create != updatedCfg.Log.Create {
		cfg.Log.Create = updatedCfg.Log.Create
		log.WithField("enabled", cfg.Log.Create).Info("Set logging for create operations")
	}

	if cfg.Log.Read != updatedCfg.Log.Read {
		cfg.Log.Read = updatedCfg.Log.Read
		log.WithField("enabled", cfg.Log.Read).Info("Set logging for read operations")
	}

	if cfg.Log.Update != updatedCfg.Log.Update {
		cfg.Log.Update = updatedCfg.Log.Update
		log.WithField("enabled", cfg.Log.Update).Info("Set logging for update operations")
	}

	if cfg.Log.Delete != updatedCfg.Log.Delete {
		cfg.Log.Delete = updatedCfg.Log.Delete
		log.WithField("enabled", cfg.Log.Delete).Info("Set logging for delete operations")
	}
}

func (cfg *Config) ensureUserDirs() {
	if _, err := os.Stat(cfg.Dir); os.IsNotExist(err) {
		os.Mkdir(cfg.Dir, os.ModePerm)
		log.WithField("path", cfg.Dir).Info("Created base dir")
	}

	for _, user := range cfg.Users {
		if user.Subdir != nil {
			path := filepath.Join(cfg.Dir, *user.Subdir)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				os.Mkdir(path, os.ModePerm)
				log.WithField("path", path).Info("Created user dir")
			}
		}
	}
}
