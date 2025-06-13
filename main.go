package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/epsilonrhorho/dns-updater/dns"
	"github.com/epsilonrhorho/dns-updater/ipify"
	"github.com/epsilonrhorho/dns-updater/service"
	"github.com/epsilonrhorho/dns-updater/storage"
)

type Config struct {
	// General settings
	DNSProvider    string        `envconfig:"DNS_PROVIDER" required:"true"`
	Zone           string        `envconfig:"ZONE" required:"true"`
	RecordName     string        `envconfig:"RECORD_NAME" required:"true"`
	StoragePath    string        `envconfig:"STORAGE_PATH" required:"true"`
	UpdateInterval time.Duration `envconfig:"UPDATE_INTERVAL" default:"2m"`
	TTL            time.Duration `envconfig:"TTL" default:"60s"`
	
	// AWS Route53 settings (prefixed with AWS_)
	AWSAccessKeyID     string `envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY"`
	AWSRegion          string `envconfig:"AWS_REGION"`
	
	// Cloudflare settings (prefixed with CF_)
	CFAPIToken string `envconfig:"CF_API_TOKEN"`
	CFEmail    string `envconfig:"CF_EMAIL"`
	CFAPIKey   string `envconfig:"CF_API_KEY"`
}


func main() {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatalf("failed to parse environment variables: %v", err)
	}

	// Create DNS provider configuration
	dnsConfig := dns.Config{
		Provider:           c.DNSProvider,
		AWSAccessKeyID:     c.AWSAccessKeyID,
		AWSSecretAccessKey: c.AWSSecretAccessKey,
		AWSRegion:          c.AWSRegion,
		CFAPIToken:         c.CFAPIToken,
		CFEmail:            c.CFEmail,
		CFAPIKey:           c.CFAPIKey,
	}

	// Create DNS provider
	dnsProvider, err := dns.NewProvider(dnsConfig)
	if err != nil {
		log.Fatalf("failed to create DNS provider: %v", err)
	}

	// Create other clients
	ipClient := ipify.NewClient(nil)
	storageClient := storage.NewFileStorage(c.StoragePath)

	// Create service configuration
	serviceConfig := service.Config{
		Zone:       c.Zone,
		RecordName: c.RecordName,
		TTL:        c.TTL,
	}

	// Create and run the service
	dnsService := service.New(dnsProvider, ipClient, storageClient, serviceConfig, c.UpdateInterval)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dnsService.Run(ctx)
}
