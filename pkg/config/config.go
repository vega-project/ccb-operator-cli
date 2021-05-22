package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/user"
	"path/filepath"
)

var (
	DefaultConfigPath     = ".config/ccboc"
	DefaultConfigFileName = "config"
)

type Config struct {
	APIURL string `json:"api_url"`
	Token  string `json:"token"`
}

func (c *Config) Validate() error {
	if _, err := url.ParseRequestURI(c.APIURL); err != nil {
		return fmt.Errorf("non valid URL: %v", err)
	}
	return nil
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %s: %v", configPath, err)
	}

	var c *Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("coudln't unmarshal config %s: %v", configPath, err)
	}

	return c, nil
}

func GetDefaultConfigPath() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("couldn't get user's information: %v", err)
	}
	home := user.HomeDir

	return filepath.Join(home, DefaultConfigPath, DefaultConfigFileName), nil
}
