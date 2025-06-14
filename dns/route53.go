package dns

import (
	"context"
	"time"

	"github.com/libdns/libdns"
	"github.com/libdns/route53"
)

// Route53Provider wraps the libdns Route53 provider.
type Route53Provider struct {
	provider *route53.Provider
}

// NewRoute53Provider creates a new Route53 DNS provider.
func NewRoute53Provider(config Config) (*Route53Provider, error) {
	provider := &route53.Provider{
		AccessKeyId:     config.AWSAccessKeyID,
		SecretAccessKey: config.AWSSecretAccessKey,
		Region:          config.AWSRegion,
	}
	
	return &Route53Provider{
		provider: provider,
	}, nil
}

// UpdateARecord updates an A record using the Route53 provider.
func (r *Route53Provider) UpdateARecord(ctx context.Context, zone, name, ip string, ttl time.Duration) error {
	record, err := createRecord(name, zone, ip, ttl)
	if err != nil {
		return err
	}
	
	normalizedZone := normalizeZone(zone)
	
	// Use SetRecords to upsert the record
	_, err = r.provider.SetRecords(ctx, normalizedZone, []libdns.Record{record})
	return err
}