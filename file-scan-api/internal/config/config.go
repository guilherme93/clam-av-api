package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	App struct {
		Name string `json:"name" mapstructure:"NAME"`
		Port int    `json:"port" mapstructure:"PORT"`
	} `json:"app" mapstructure:"APP"`

	ClamAV ClamAVConfig `json:"clam_av" mapstructure:"CLAM_AV"`

	MaxFileSize int `json:"max_file_size" mapstructure:"MAX_FILE_SIZE"`
}

type ClamAVConfig struct {
	Address        string `json:"address" mapstructure:"ADDRESS"`
	TimeoutSeconds int    `json:"timeout" mapstructure:"TIMEOUT"`
}

func New() (*Config, error) {
	viper.SetConfigType("json")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config

	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err = viper.ReadConfig(bytes.NewReader(b)); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err = viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if err = validate[Config](cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	if cfg.ClamAV.TimeoutSeconds == 0 {
		cfg.ClamAV.TimeoutSeconds = 30
	}

	return &cfg, nil
}

// validate validates the config against the struct tags.
func validate[T any](target T) error {
	return validator.New().Struct(target)
}
