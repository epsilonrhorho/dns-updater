package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/epsilonrhorho/dns-updater/dns"
	"github.com/epsilonrhorho/dns-updater/ipify"
	"github.com/epsilonrhorho/dns-updater/service"
	"github.com/epsilonrhorho/dns-updater/storage"
)

type RecordConfig struct {
	Provider         string        `yaml:"provider"`
	TTL              time.Duration `yaml:"ttl,omitempty"`
	AWSAccessKeyID   string        `yaml:"aws_access_key_id,omitempty"`
	AWSSecretKey     string        `yaml:"aws_secret_key,omitempty"`
	AWSRegion        string        `yaml:"aws_region,omitempty"`
	CFAPIToken       string        `yaml:"cf_api_token,omitempty"`
	CFEmail          string        `yaml:"cf_email,omitempty"`
	CFAPIKey         string        `yaml:"cf_api_key,omitempty"`
}

type Config struct {
	UpdateInterval time.Duration                `yaml:"update_interval"`
	StoragePath    string                       `yaml:"storage_path"`
	Records        map[string]RecordConfig      `yaml:"records"`
}

// extractZoneFromRecordName extracts the zone from a record name.
// Assumes the hostname part is exactly one DNS label.
// For example: "foo.example.com" -> "example.com"
// Returns an error if the record name has fewer than 3 labels.
func extractZoneFromRecordName(recordName string) (string, error) {
	parts := strings.Split(recordName, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("record name '%s' must have at least 3 DNS labels (e.g., host.domain.tld)", recordName)
	}
	return strings.Join(parts[1:], "."), nil
}


func loadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.UpdateInterval == 0 {
		config.UpdateInterval = 2 * time.Minute
	}
	if config.StoragePath == "" {
		config.StoragePath = "/tmp/dns-updater"
	}

	for recordName, recordConfig := range config.Records {
		if recordConfig.TTL == 0 {
			rc := recordConfig
			rc.TTL = 60 * time.Second
			config.Records[recordName] = rc
		}
	}

	return &config, nil
}

func main() {
	configPath := flag.String("c", "/usr/local/etc/dns-updater.yaml", "path to configuration file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	if len(config.Records) == 0 {
		log.Fatal("no DNS records configured")
	}

	ipClient := ipify.NewClient(nil)
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	
	for recordName, recordConfig := range config.Records {
		wg.Add(1)
		go func(name string, rConfig RecordConfig) {
			defer wg.Done()
			
			zone, err := extractZoneFromRecordName(name)
			if err != nil {
				log.Printf("invalid record name %s: %v", name, err)
				return
			}
			
			dnsConfig := dns.Config{
				Provider:           rConfig.Provider,
				AWSAccessKeyID:     rConfig.AWSAccessKeyID,
				AWSSecretAccessKey: rConfig.AWSSecretKey,
				AWSRegion:          rConfig.AWSRegion,
				CFAPIToken:         rConfig.CFAPIToken,
				CFEmail:            rConfig.CFEmail,
				CFAPIKey:           rConfig.CFAPIKey,
			}

			dnsProvider, err := dns.NewProvider(dnsConfig)
			if err != nil {
				log.Printf("failed to create DNS provider for %s: %v", name, err)
				return
			}

			storageClient := storage.NewFileStorage(config.StoragePath + "/" + name)

			serviceConfig := service.Config{
				Zone:       zone,
				RecordName: name,
				TTL:        rConfig.TTL,
			}

			dnsService := service.New(dnsProvider, ipClient, storageClient, serviceConfig, config.UpdateInterval)
			dnsService.Run(ctx)
		}(recordName, recordConfig)
	}

	wg.Wait()
}
