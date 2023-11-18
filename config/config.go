package config

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const defaultConfigFile = "config.toml"

type API struct {
	ListenAddress      string   `mapstructure:"listen_address"`
	Port               int      `mapstructure:"port"`
	CorsAllowedOrigins []string `mapstructure:"cors_allowed_origins"`
}

type Auth struct {
	SessionTtl int    `mapstructure:"session_ttl"`
	SigningKey string `mapstructure:"signing_key"`
}

type Database struct {
	Location string `mapstructure:"location"`
}

type Observability struct {
	WriteToFile    bool `mapstructure:"write_to_file"`
	WriteToConsole bool `mapstructure:"write_to_console"`

	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type Config struct {
	API           API           `mapstructure:"api"`
	Auth          Auth          `mapstructure:"auth"`
	Database      Database      `mapstructure:"database"`
	Observability Observability `mapstructure:"observability"`
}

func LoadDefault() (*Config, error) {
	return Load(defaultConfigFile)
}

func Load(configFile string) (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()

	v.SetConfigType("toml")
	v.SetConfigFile(configFile)

	v.SetEnvPrefix("sparovec")
	v.AllowEmptyEnv(true)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := &Config{}
	err := v.Unmarshal(conf, func(dc *mapstructure.DecoderConfig) {
		dc.ErrorUnset = true
		dc.ErrorUnused = true
	})

	return conf, err
}
