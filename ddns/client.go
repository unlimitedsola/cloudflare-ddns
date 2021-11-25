package ddns

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
)

type DDNS struct {
	api    *cloudflare.API
	config *Config
}

func New() (*DDNS, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	client, err := createAPIClient(config)
	if err != nil {
		return nil, err
	}
	return &DDNS{
		api:    client,
		config: config,
	}, nil
}

func createAPIClient(config *Config) (*cloudflare.API, error) {
	return cloudflare.NewWithAPIToken(config.APIToken)
}

func (d *DDNS) Update(ctx context.Context) (hasChanged bool, oldIP string, newIP string, err error) {
	ip, err := getSelfIpv6Addr(ctx)
	if err != nil {
		return false, "", "", err
	}
	config := d.config
	api := d.api
	record, err := api.DNSRecord(ctx, config.ZoneId, config.DNSRecordId)
	if err != nil {
		return false, "", "", err
	}
	if record.Content != ip {
		old := record.Content
		record.Content = ip
		err := api.UpdateDNSRecord(ctx, config.ZoneId, config.DNSRecordId, record)
		if err != nil {
			return false, "", "", err
		}
		return true, old, ip, nil
	}
	return false, record.Content, ip, nil
}
