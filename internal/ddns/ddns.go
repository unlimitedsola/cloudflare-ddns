package ddns

import (
	"cloudflare-ddns/internal/util"
	"encoding/json"
	"errors"
	"github.com/cloudflare/cloudflare-go"
	"io/ioutil"
	"net"
	"path/filepath"
)

type Config struct {
	APIToken    string `json:"api_token"`
	ZoneId      string `json:"zone_id"`
	DNSRecordId string `json:"dns_record_id"`
}

type Client struct {
	api    *cloudflare.API
	config *Config
}

func New() (*Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	client, err := createAPIClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		api:    client,
		config: config,
	}, nil
}

func (c *Client) Update() (hasChanged bool, oldIP string, newIP string, err error) {
	ip, err := getPublicIPv6Addr()
	if err != nil {
		return false, "", "", err
	}
	config := c.config
	api := c.api
	record, err := api.DNSRecord(config.ZoneId, config.DNSRecordId)
	if err != nil {
		return false, "", "", err
	}
	if record.Content != ip.String() {
		old := record.Content
		record.Content = ip.String()
		err := api.UpdateDNSRecord(config.ZoneId, config.DNSRecordId, record)
		if err != nil {
			return false, "", "", err
		}
		return true, old, ip.String(), nil
	}
	return false, record.Content, ip.String(), nil
}

func createAPIClient(config *Config) (*cloudflare.API, error) {
	return cloudflare.NewWithAPIToken(config.APIToken)
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

func getPublicIPv6Addr() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := addr.(*net.IPNet).IP
			if isPublicIPv6Addr(ip) {
				return ip, nil
			}
		}
	}
	return nil, errors.New("no public IPv6 address can be found")
}

func isPublicIPv6Addr(ip net.IP) bool {
	if ip.To4() != nil {
		return false
	}
	if ip[0] == 0xfe && ip[1] == 0x80 {
		return false
	}
	return true
}
