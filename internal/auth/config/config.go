package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/viper"
)

// AuthConfig holds static settings for the auth-service plus Vault connection info.
type AuthConfig struct {
	Server  ServerConfig    `mapstructure:"server"`
	Logging LoggingConfig   `mapstructure:"logging"`
	Vault   VaultAuthConfig `mapstructure:"vault"`
}

// ServerConfig configures HTTP server port.
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// LoggingConfig sets log level.
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// VaultAuthConfig describes how auth-service talks to Vault.
type VaultAuthConfig struct {
	Address     string        `mapstructure:"address"`      // e.g. http://127.0.0.1:8200
	AuthMethod  string        `mapstructure:"auth_method"`  // "token" or "k8s"
	TokenEnv    string        `mapstructure:"token_env"`    // env-var holding Vault token (if token mode)
	JWTPath     string        `mapstructure:"jwt_path"`     // path to K8s ServiceAccount JWT (if k8s mode)
	Role        string        `mapstructure:"role"`         // Vault Kubernetes auth role
	KVPath      string        `mapstructure:"kv_path"`      // e.g. "secret/data/"
	TransitPath string        `mapstructure:"transit_path"` // e.g. "transit/"
	TokenTTL    time.Duration `mapstructure:"token_ttl"`    // TTL for issued tokens (e.g. "1h")
}

// LoadAuthConfig reads configs from file/env and returns an AuthConfig.
func LoadAuthConfig(path string) (*AuthConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvPrefix("AUTH")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg AuthConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal auth config: %w", err)
	}

	return &cfg, nil
}

// NewVaultClient initializes and authenticates a Vault client per VaultAuthConfig.
func NewVaultClient(vc VaultAuthConfig) (*api.Client, error) {
	vaultCfg := api.DefaultConfig()
	vaultCfg.Address = vc.Address
	client, err := api.NewClient(vaultCfg)
	if err != nil {
		return nil, err
	}

	switch vc.AuthMethod {
	case "token":
		token := os.Getenv(vc.TokenEnv)
		if token == "" {
			return nil, fmt.Errorf("env %s is empty", vc.TokenEnv)
		}
		client.SetToken(token)

	case "k8s":
		jwt, err := os.ReadFile(vc.JWTPath)
		if err != nil {
			return nil, fmt.Errorf("read jwt (%s): %w", vc.JWTPath, err)
		}
		resp, err := client.Logical().Write("auth/kubernetes/login", map[string]interface{}{
			"role": vc.Role,
			"jwt":  string(jwt),
		})
		if err != nil {
			return nil, fmt.Errorf("vault k8s login: %w", err)
		}
		client.SetToken(resp.Auth.ClientToken)

	default:
		return nil, fmt.Errorf("unsupported vault auth_method: %s", vc.AuthMethod)
	}

	return client, nil
}
