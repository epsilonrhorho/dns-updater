package dns

import (
	"context"
	"fmt"
	"time"

	"github.com/libdns/cloudflare"
	"github.com/libdns/libdns"
)

// CloudflareProvider wraps the libdns Cloudflare provider.
type CloudflareProvider struct {
	provider *cloudflare.Provider
}

// NewCloudflareProvider creates a new Cloudflare DNS provider.
func NewCloudflareProvider(config Config) (*CloudflareProvider, error) {
	if config.CFAPIToken == "" {
		return nil, fmt.Errorf("Cloudflare provider requires CF_API_TOKEN")
	}
	
	provider := &cloudflare.Provider{
		APIToken: config.CFAPIToken,
	}
	
	return &CloudflareProvider{
		provider: provider,
	}, nil
}

// UpdateARecord updates an A record using the Cloudflare provider.
func (c *CloudflareProvider) UpdateARecord(ctx context.Context, zone, name, ip string, ttl time.Duration) error {
	record, err := createRecord(name, zone, ip, ttl)
	if err != nil {
		return err
	}
	
	normalizedZone := normalizeZone(zone)
	
	// Use SetRecords to upsert the record
	_, err = c.provider.SetRecords(ctx, normalizedZone, []libdns.Record{record})
	return err
}