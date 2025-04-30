package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("couldn't get $HOME directory: %w", err)
	}

	path := filepath.Join(home, configFileName)
	return path, nil
}

func Write(cfg *Config) error {
	data, err := cfg.FormatPrettyJSON()
	// data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("couldn't marshal JSON configuration file: %w", err)
	}

	path, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("couldn't write configuration file: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("couldn't write configuration file: %w", err)
	}

	return nil
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("couldn't load '%v' from disk: %w", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't load '%v' from disk: %w", path, err)
	}

	var gatorConfig Config
	if err := json.Unmarshal(data, &gatorConfig); err != nil {
		return Config{}, fmt.Errorf("couldn't unmarshal the contents of '%v': %w", path, err)
	}

	return gatorConfig, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	err := Write(c)
	if err != nil {
		return fmt.Errorf("couldn't write configuration file to disk after setting user: %w", err)
	}
	return nil
}

func (c *Config) FormatPrettyJSON() ([]byte, error) {
	jsonData, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal JSON configuration file while prettyfying: %w", err)
	}
	return jsonData, nil
}
