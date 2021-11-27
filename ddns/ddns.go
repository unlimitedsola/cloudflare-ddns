package ddns

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
)

type DDNS struct {
	api    *cloudflare.API
	config *Config
}

func New() (ddns *DDNS, err error) {
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

func createAPIClient(config *Config) (api *cloudflare.API, err error) {
	return cloudflare.NewWithAPIToken(config.APIToken)
}

type UpdateResult struct {
	Name     string
	Updated  bool
	Previous string
	Current  string
}

type Handler func(res UpdateResult, err error)

func (d *DDNS) Run(ctx context.Context, handler Handler) error {
	if d.config.IPv4 {
		ipv4, err := getSelfIPv4Addr(ctx)
		if err != nil {
			return err
		}
		for _, zone := range d.config.Zones {
			d.updateZone(ctx, zone, "A", ipv4, handler)
		}
	}
	if d.config.IPv6 {
		ipv6, err := getSelfIPv6Addr(ctx)
		if err != nil {
			return err
		}
		for _, zone := range d.config.Zones {
			d.updateZone(ctx, zone, "AAAA", ipv6, handler)
		}
	}
	return nil
}

func (d *DDNS) updateZone(ctx context.Context, zone Zone, recordType string, ip string, handler Handler) {
	filter := cloudflare.DNSRecord{Type: recordType}
	records, err := d.api.DNSRecords(ctx, zone.ZoneId, filter)
	if err != nil {
		handler(UpdateResult{}, err)
		return
	}
main:
	for _, name := range zone.Records {
		for _, record := range records {
			if record.Name != name {
				continue
			}
			if record.Content == ip {
				handler(UpdateResult{name, false, record.Content, ip}, nil)
				continue main
			}
			old := record.Content
			record.Content = ip
			err := d.api.UpdateDNSRecord(ctx, zone.ZoneId, record.ID, record)
			if err != nil {
				handler(UpdateResult{}, err)
				continue main
			}
			handler(UpdateResult{name, true, old, record.Content}, nil)
			continue main
		}
		newRecord := cloudflare.DNSRecord{
			Type:    recordType,
			Name:    name,
			Content: ip,
			TTL:     60,
			Proxied: falsePtr(),
		}
		_, err := d.api.CreateDNSRecord(ctx, zone.ZoneId, newRecord)
		if err != nil {
			handler(UpdateResult{}, err)
			continue
		}
		handler(UpdateResult{name, true, "", ip}, nil)
	}
}

func falsePtr() *bool {
	false := false
	return &false
}
