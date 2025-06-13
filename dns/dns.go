package dns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Provider defines the interface for DNS operations.
type Provider interface {
	UpdateARecord(ctx context.Context, zone, name, ip string, ttl time.Duration) error
}

// Config represents the configuration for DNS providers.
type Config struct {
	Provider string // "route53" or "cloudflare"
	
	// AWS Route53 settings (prefixed with AWS_)
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion         string
	
	// Cloudflare settings (prefixed with CF_)
	CFAPIToken string
	CFEmail    string
	CFAPIKey   string
}

// NewProvider creates a new DNS provider based on the configuration.
func NewProvider(config Config) (Provider, error) {
	switch strings.ToLower(config.Provider) {
	case "route53":
		return NewRoute53Provider(config)
	case "cloudflare":
		return NewCloudflareProvider(config)
	default:
		return nil, fmt.Errorf("unsupported DNS provider: %s", config.Provider)
	}
}

// normalizeZone ensures the zone name ends with a dot.
func normalizeZone(zone string) string {
	if !strings.HasSuffix(zone, ".") {
		return zone + "."
	}
	return zone
}

// normalizeName ensures the record name is properly formatted.
func normalizeName(name, zone string) string {
	if name == "" || name == "@" {
		return zone
	}
	
	// If name already contains the zone, use as-is
	if strings.HasSuffix(name, zone) {
		return name
	}
	
	// If name doesn't end with dot, add zone
	if !strings.HasSuffix(name, ".") {
		return name + "." + zone
	}
	
	return name
}

// validateIP checks if the provided string is a valid IPv4 address.
func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
}

// createRecord creates a libdns.Record for an A record.
func createRecord(name, zone, ip string, ttl time.Duration) (libdns.Record, error) {
	if err := validateIP(ip); err != nil {
		return libdns.Record{}, err
	}
	
	normalizedZone := normalizeZone(zone)
	recordName := name
	
	// For libdns, we need the record name relative to the zone
	if name == "" || name == "@" {
		recordName = "@"
	} else if strings.HasSuffix(name, "."+normalizedZone) {
		// Remove the zone suffix to make it relative
		recordName = strings.TrimSuffix(name, "."+normalizedZone)
	} else if !strings.Contains(name, ".") {
		// Simple name like "home" - use as-is
		recordName = name
	}
	
	return libdns.Record{
		Type:  "A",
		Name:  recordName,
		Value: ip,
		TTL:   ttl,
	}, nil
}