package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
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

func GetPathToCalculation(defaultPath string, fileName string) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("couldn't get user's information: %v", err)
	}
	home := user.HomeDir

	if defaultPath == "" {
		return filepath.Join(home, DefaultConfigPath, fileName), nil
	}

	_, err = os.Stat(defaultPath)
	if err != nil {
		return "", fmt.Errorf("couldn't stat path %s: %v", defaultPath, err)
	}
	path := filepath.Join(defaultPath, filepath.Base(fileName))

	return path, nil
}

func CreateAndWriteFile(body []byte, defaultPath string) error {
	file, err := os.Create(defaultPath)
	if err != nil {
		return fmt.Errorf("couldn't create the calculation result file")
	}

	defer file.Close()

	_, err = file.Write(body)
	if err != nil {
		return fmt.Errorf("couldn't write the data into the tar file")
	}
	return nil
}
