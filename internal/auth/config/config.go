package config

import (
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Signing  SigningConfig  `mapstructure:"signing"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

type SigningConfig struct {
	// Base64-encoded PEM приватного ключа
	PrivateKeyB64 string `mapstructure:"private_key_b64"`
	// Base64-encoded PEM публичного ключа
	PublicKeyB64 string `mapstructure:"public_key_b64"`
}

func loadConfig(path string) Config {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvPrefix("PLUTO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	return cfg
}

var (
	instance   *Config
	once       sync.Once
	configPath = "configs/auth.yaml"
)

// GetConfig возвращает singleton Config
func GetConfig() *Config {
	once.Do(func() {
		cfg := loadConfig(configPath)
		instance = &cfg
	})
	return instance
}
