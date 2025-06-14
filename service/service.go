package service

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/epsilonrhorho/dns-updater/dns"
	"github.com/epsilonrhorho/dns-updater/ipify"
	"github.com/epsilonrhorho/dns-updater/storage"
)

// Config holds the configuration for the DNS updater service.
type Config struct {
	Zone       string
	RecordName string
	TTL        time.Duration
}

// Service handles DNS updates with dependency injection.
type Service struct {
	dnsProvider dns.Provider
	ipClient    ipify.ClientInterface
	storage     storage.Interface
	config      Config
	interval    time.Duration
}

// New creates a new Service instance.
func New(
	dnsProvider dns.Provider,
	ipClient ipify.ClientInterface,
	storage storage.Interface,
	config Config,
	interval time.Duration,
) *Service {
	return &Service{
		dnsProvider: dnsProvider,
		ipClient:    ipClient,
		storage:     storage,
		config:      config,
		interval:    interval,
	}
}

// getCurrentIP fetches the current public IP address.
func (s *Service) getCurrentIP(ctx context.Context) (string, error) {
	return s.ipClient.GetIP(ctx)
}

// hasIPChanged checks if the current IP differs from the stored IP.
func (s *Service) hasIPChanged(currentIP string) (bool, error) {
	lastIP, err := s.storage.ReadLastIP()
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return lastIP != currentIP, nil
}

// updateDNSRecord updates the DNS A record with the new IP.
func (s *Service) updateDNSRecord(ctx context.Context, ip string) error {
	return s.dnsProvider.UpdateARecord(ctx, s.config.Zone, s.config.RecordName, ip, s.config.TTL)
}

// storeIP persists the IP address to storage.
func (s *Service) storeIP(ip string) error {
	return s.storage.WriteIP(ip)
}

// Update performs a single DNS update check and update if necessary.
func (s *Service) Update(ctx context.Context) error {
	ip, err := s.getCurrentIP(ctx)
	if err != nil {
		return err
	}
	log.Println("Public IP:", ip)

	changed, err := s.hasIPChanged(ip)
	if err != nil {
		return err
	}

	if !changed {
		log.Println("IP unchanged; skipping DNS update")
		return nil
	}

	if err := s.updateDNSRecord(ctx, ip); err != nil {
		return err
	}

	if err := s.storeIP(ip); err != nil {
		return err
	}

	log.Println("DNS A record updated")
	return nil
}

// Run starts the continuous DNS update service.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		if err := s.Update(ctx); err != nil {
			log.Printf("update failed: %v", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			log.Println("shutting down")
			return
		}
	}
}