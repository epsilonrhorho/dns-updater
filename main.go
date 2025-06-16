package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
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
	Zone             string        `yaml:"zone,omitempty"`
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
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	config, err := loadConfig(configPath)
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
				Zone:       rConfig.Zone,
				RecordName: name,
				TTL:        rConfig.TTL,
			}

			dnsService := service.New(dnsProvider, ipClient, storageClient, serviceConfig, config.UpdateInterval)
			dnsService.Run(ctx)
		}(recordName, recordConfig)
	}

	wg.Wait()
}
