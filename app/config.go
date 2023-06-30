package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config represents the configuration of the server application.
type Config struct {
	Address string
	Port    string
	Prefix  string
	Dir     string
	Deny    Deny
	TLS     *TLS
	Log     Logging
	Realm   string
	Users   map[string]*UserInfo
	Cors    Cors
}

type Deny struct {
	Create Create
}

type Create struct {
	File      []string
	Directory []string
}

// Logging allows definition for logging each CRUD method.
type Logging struct {
	Error  bool
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

// Cors contains settings related to Cross-Origin Resource Sharing (CORS)
type Cors struct {
	Origin      string
	Credentials bool
}

// ParseConfig parses the application configuration an sets defaults.
func ParseConfig(path string) *Config {
	var cfg = &Config{}

	setDefaults()
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.swd")
		viper.AddConfigPath("$HOME/.dave")
		viper.AddConfigPath(".")
	}

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
	viper.OnConfigChange(cfg.handleConfigUpdate)

	cfg.ensureUserDirs()

	return cfg
}

// setDefaults adds some default values for the configuration
func setDefaults() {
	viper.SetDefault("Address", "127.0.0.1")
	viper.SetDefault("Port", "8000")
	viper.SetDefault("Prefix", "")
	viper.SetDefault("Dir", "/tmp")
	viper.SetDefault("Deny.Create.File", nil)
	viper.SetDefault("Deny.Create.Directory", nil)
	viper.SetDefault("Users", nil)
	viper.SetDefault("TLS", nil)
	viper.SetDefault("Realm", "dave")
	viper.SetDefault("Log.Error", true)
	viper.SetDefault("Log.Create", false)
	viper.SetDefault("Log.Read", false)
	viper.SetDefault("Log.Update", false)
	viper.SetDefault("Log.Delete", false)
	viper.SetDefault("Cors.Credentials", false)
}

// AuthenticationNeeded returns whether users are defined and authentication is required
func (cfg *Config) AuthenticationNeeded() bool {
	return cfg.Users != nil && len(cfg.Users) != 0
}

func (cfg *Config) handleConfigUpdate(e fsnotify.Event) {
	var err error
	defer func() {
		r := recover()
		switch t := r.(type) {
		case string:
			log.WithError(errors.New(t)).Error("Error updating configuration. Please restart the server...")
		case error:
			log.WithError(t).Error("Error updating configuration. Please restart the server...")
		}
	}()

	log.WithField("path", e.Name).Info("Config file changed")

	file, err := os.Open(e.Name)
	if err != nil {
		log.WithField("path", e.Name).Warn("Error reloading config")
	}

	var updatedCfg = &Config{}
	viper.ReadConfig(file)
	viper.Unmarshal(&updatedCfg)

	updateConfig(cfg, updatedCfg)
}

func updateConfig(cfg *Config, updatedCfg *Config) {
	for username := range cfg.Users {
		if updatedCfg.Users[username] == nil {
			log.WithField("user", username).Info("Removed User from configuration")
			delete(cfg.Users, username)
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
			if cfg.Users[username].Subdir != v.Subdir {
				log.WithField("user", username).Info("Updated subdir of user")
				cfg.Users[username].Subdir = v.Subdir
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
	if !stringSlicesEqual(cfg.Deny.Create.File, updatedCfg.Deny.Create.File) {
		cfg.Deny.Create.File = updatedCfg.Deny.Create.File
		log.WithField("updated", strings.Join(cfg.Deny.Create.File, "; ")).Info("Updated denied file create entries")
	}
	if !stringSlicesEqual(cfg.Deny.Create.Directory, updatedCfg.Deny.Create.Directory) {
		cfg.Deny.Create.Directory = updatedCfg.Deny.Create.Directory
		log.WithField("updated", strings.Join(cfg.Deny.Create.Directory, "; ")).Info("Updated denied directory create entries")
	}
}

func (cfg *Config) ensureUserDirs() {
	if _, err := os.Stat(cfg.Dir); os.IsNotExist(err) {
		mkdirErr := os.Mkdir(cfg.Dir, os.ModePerm)
		if mkdirErr != nil {
			log.WithField("path", cfg.Dir).WithField("error", err).Warn("Can't create base dir")
			return
		}
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

func stringSlicesEqual(f, j []string) bool {
	if len(f) != len(j) {
		return false
	}
	for i, v := range f {
		if v != j[i] {
			return false
		}
	}
	return true
}
