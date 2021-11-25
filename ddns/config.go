package ddns

import (
	"cloudflare-ddns/internal/util"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	APIToken    string `json:"api_token"`
	ZoneId      string `json:"zone_id"`
	DNSRecordId string `json:"dns_record_id"`
}

func getConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getConfigPath() (string, error) {
	execPath, err := util.GetExecutableAbsolutePath()
	if err != nil {
		return "", err
	}
	configDir := filepath.Dir(execPath)
	configPath := filepath.Join(configDir, "config.json")
	return configPath, nil
}
