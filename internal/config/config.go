package config

import (
	"errors"
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"os"
)

const (
	defaultConfigPath = "./internal/config/config.yaml"
)

type Config struct {
	Env             string `yaml:"env"`
	EnvPath         string `yaml:"env_path"`
	StoragePath     string `yaml:"storage_path"`
	BirthdayGroupID int64  `yaml:"birthday_group_id"`
	GroupOwnerID    int64  `yaml:"group_owner_id"`
}

func LoadConfig() (*Config, error) {

	configPath := fetchConfigPath(defaultConfigPath)
	if _, fcpErr := os.Stat(configPath); errors.Is(fcpErr, os.ErrExist) {
		return nil, fcpErr
	}

	var cfg Config

	if rcErr := cleanenv.ReadConfig(configPath, &cfg); rcErr != nil {
		return nil, rcErr
	}

	envLoadErr := godotenv.Load(cfg.EnvPath)
	if envLoadErr != nil {
		return nil, envLoadErr
	}

	return &cfg, nil
}

// fetchConfigPath return config file path with priority: flag > env > default
func fetchConfigPath(defaultConfigPath string) string {
	var configPath string

	flag.StringVar(&configPath, "c", "", "path to config file")

	if configPath == "" {
		var exists bool

		configPath, exists = os.LookupEnv("CONFIG_PATH")
		if !exists {
			configPath = defaultConfigPath
		}
		return configPath
	}
	return configPath
}
