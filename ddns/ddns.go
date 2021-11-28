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

type Handler interface {
	OnZoneError(zone string, err error)
	OnError(name string, err error)
	OnCreate(name string, recordType string, current string)
	OnUpdate(name string, recordType string, previous string, current string)
}

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
		handler.OnZoneError(zone.ZoneId, err)
		return
	}
main:
	for _, name := range zone.Records {
		for _, record := range records {
			if record.Name != name {
				continue
			}
			if record.Content == ip {
				continue main
			}
			old := record.Content
			record.Content = ip
			err := d.api.UpdateDNSRecord(ctx, zone.ZoneId, record.ID, record)
			if err != nil {
				handler.OnError(name, err)
				continue main
			}
			handler.OnUpdate(name, recordType, old, record.Content)
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
			handler.OnError(name, err)
			continue
		}
		handler.OnCreate(name, recordType, ip)
	}
}

func falsePtr() *bool {
	false := false
	return &false
}
