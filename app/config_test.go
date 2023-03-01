package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestParseConfig(t *testing.T) {
	viper.Reset()

	tmpDir := filepath.Join(os.TempDir(), "dave__"+strconv.FormatInt(time.Now().UnixNano(), 10))
	os.Mkdir(tmpDir, 0700)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name string
		want *Config
	}{
		{"default", cfg(t, tmpDir)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := ParseConfig(""); !reflect.DeepEqual(got, tt.want) {
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				t.Errorf("ParseConfig(\"\") = %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}

func cfg(t *testing.T, tmpDir string) *Config {
	viper.SetConfigType("yaml")
	var yamlCfg = []byte(`
address: 1.2.3.4
port: 42
prefix: /oh-de-lally
tls:
  keyFile: ` + tmpDir + `/robin.pem
  certFile: ` + tmpDir + `/tuck.pem
dir: /sherwood/forest
realm: uk
users:
  lj:
    password: 123
    subdir: /littlejohn
  srf:
    password: 234
    subdir: /sheriff
log:
  error: true
`)

	err := ioutil.WriteFile(filepath.Join(tmpDir, "config.yaml"), yamlCfg, 0600)
	if err != nil {
		t.Errorf("error writing test config. error = %v", err)
	}

	err = viper.ReadConfig(bytes.NewBuffer(yamlCfg))
	if err != nil {
		t.Errorf("error reading test config. error = %v", err)
	}
	var cfg = &Config{}
	viper.Unmarshal(&cfg)

	// let viper read from the tmp directory fist
	viper.AddConfigPath(tmpDir)

	// add dummy cert and key file
	_, err = os.OpenFile(filepath.Join(tmpDir, "robin.pem"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Errorf("error creating key file. error = %v", err)
		return nil
	}

	_, err = os.OpenFile(filepath.Join(tmpDir, "tuck.pem"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Errorf("error creating cert file. error = %v", err)
		return nil
	}
	viper.AddConfigPath(tmpDir)

	return cfg
}

func TestSetDefaults(t *testing.T) {
	viper.Reset()

	tests := []struct {
		name  string
		value interface{}
	}{
		{"Address", "127.0.0.1"},
		{"Port", "8000"},
		{"Prefix", ""},
		{"Dir", "/tmp"},
		{"TLS", nil},
		{"Realm", "dave"},
		{"Log.Error", true},
		{"Log.Create", false},
		{"Log.Read", false},
		{"Log.Update", false},
		{"Log.Delete", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			setDefaults()

			if viper.Get(tt.name) != tt.value {
				t.Errorf("Default Keys doesn't fit. name = %v, want = %v", tt.name, tt.value)
			}
		})
	}
}
