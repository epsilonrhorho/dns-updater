package service

import (
	"context"
	"testing"
	"time"
)

func TestService_Update(t *testing.T) {
	tests := []struct {
		name                string
		currentIP           string
		currentIPError      error
		lastIP              string
		lastIPError         error
		dnsError            error
		storageWriteError   error
		expectError         bool
		expectDNSCalled     bool
		expectStorageCalled bool
	}{
		{
			name:                "successful update with IP change",
			currentIP:           "192.168.1.2",
			lastIP:              "192.168.1.1",
			expectError:         false,
			expectDNSCalled:     true,
			expectStorageCalled: true,
		},
		{
			name:                "no update when IP unchanged",
			currentIP:           "192.168.1.1",
			lastIP:              "192.168.1.1",
			expectError:         false,
			expectDNSCalled:     false,
			expectStorageCalled: false,
		},
		{
			name:                "update when no previous IP stored",
			currentIP:           "192.168.1.1",
			lastIP:              "",
			expectError:         false,
			expectDNSCalled:     true,
			expectStorageCalled: true,
		},
		{
			name:           "error getting current IP",
			currentIPError: errIP,
			expectError:    true,
		},
		{
			name:        "error reading last IP",
			currentIP:   "192.168.1.1",
			lastIPError: errStorage,
			expectError: true,
		},
		{
			name:            "error updating DNS",
			currentIP:       "192.168.1.2",
			lastIP:          "192.168.1.1",
			dnsError:        errDNS,
			expectError:     true,
			expectDNSCalled: true,
		},
		{
			name:                "error writing to storage",
			currentIP:           "192.168.1.2",
			lastIP:              "192.168.1.1",
			storageWriteError:   errStorage,
			expectError:         true,
			expectDNSCalled:     true,
			expectStorageCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dnsProviderCalled, storageCalled bool

			mockDNS := &mockDNSProvider{
				updateARecordFunc: func(ctx context.Context, zone, name, ip string, ttl time.Duration) error {
					dnsProviderCalled = true
					return tt.dnsError
				},
			}

			mockIP := &mockIPClient{
				getIPFunc: func(ctx context.Context) (string, error) {
					return tt.currentIP, tt.currentIPError
				},
			}

			mockStore := &mockStorage{
				readLastIPFunc: func() (string, error) {
					return tt.lastIP, tt.lastIPError
				},
				writeIPFunc: func(ip string) error {
					storageCalled = true
					return tt.storageWriteError
				},
			}

			config := Config{
				Zone:       "example.com",
				RecordName: "home",
				TTL:        60 * time.Second,
			}

			service := New(mockDNS, mockIP, mockStore, config, time.Minute)
			err := service.Update(context.Background())

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if dnsProviderCalled != tt.expectDNSCalled {
				t.Errorf("expected DNS called: %v, got: %v", tt.expectDNSCalled, dnsProviderCalled)
			}
			if storageCalled != tt.expectStorageCalled {
				t.Errorf("expected storage called: %v, got: %v", tt.expectStorageCalled, storageCalled)
			}
		})
	}
}

func TestService_getCurrentIP(t *testing.T) {
	tests := []struct {
		name        string
		mockIP      string
		mockError   error
		expectedIP  string
		expectError bool
	}{
		{
			name:        "successful IP fetch",
			mockIP:      "203.0.113.1",
			expectedIP:  "203.0.113.1",
			expectError: false,
		},
		{
			name:        "IP fetch error",
			mockError:   errIP,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIP := &mockIPClient{
				getIPFunc: func(ctx context.Context) (string, error) {
					return tt.mockIP, tt.mockError
				},
			}

			service := &Service{ipClient: mockIP}
			ip, err := service.getCurrentIP(context.Background())

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ip != tt.expectedIP {
				t.Errorf("expected IP %q, got %q", tt.expectedIP, ip)
			}
		})
	}
}

func TestService_hasIPChanged(t *testing.T) {
	tests := []struct {
		name        string
		currentIP   string
		lastIP      string
		storageErr  error
		expectError bool
		expected    bool
	}{
		{
			name:      "IP changed",
			currentIP: "192.168.1.2",
			lastIP:    "192.168.1.1",
			expected:  true,
		},
		{
			name:      "IP unchanged",
			currentIP: "192.168.1.1",
			lastIP:    "192.168.1.1",
			expected:  false,
		},
		{
			name:      "no previous IP",
			currentIP: "192.168.1.1",
			lastIP:    "",
			expected:  true,
		},
		{
			name:        "storage error",
			currentIP:   "192.168.1.1",
			storageErr:  errStorage,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &mockStorage{
				readLastIPFunc: func() (string, error) {
					return tt.lastIP, tt.storageErr
				},
			}

			service := &Service{storage: mockStore}
			changed, err := service.hasIPChanged(tt.currentIP)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if changed != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, changed)
			}
		})
	}
}

func TestService_updateDNSRecord(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		dnsError    error
		expectError bool
	}{
		{
			name:        "successful DNS update",
			ip:          "192.168.1.1",
			expectError: false,
		},
		{
			name:        "DNS update error",
			ip:          "192.168.1.1",
			dnsError:    errDNS,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedParams []interface{}
			mockDNS := &mockDNSProvider{
				updateARecordFunc: func(ctx context.Context, zone, name, ip string, ttl time.Duration) error {
					capturedParams = []interface{}{zone, name, ip, ttl}
					return tt.dnsError
				},
			}

			config := Config{
				Zone:       "example.com",
				RecordName: "home",
				TTL:        300 * time.Second,
			}

			service := &Service{
				dnsProvider: mockDNS,
				config:      config,
			}

			err := service.updateDNSRecord(context.Background(), tt.ip)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && len(capturedParams) == 4 {
				if capturedParams[0] != config.Zone {
					t.Errorf("expected zone %q, got %q", config.Zone, capturedParams[0])
				}
				if capturedParams[1] != config.RecordName {
					t.Errorf("expected record name %q, got %q", config.RecordName, capturedParams[1])
				}
				if capturedParams[2] != tt.ip {
					t.Errorf("expected IP %q, got %q", tt.ip, capturedParams[2])
				}
				if capturedParams[3] != config.TTL {
					t.Errorf("expected TTL %v, got %v", config.TTL, capturedParams[3])
				}
			}
		})
	}
}

func TestService_storeIP(t *testing.T) {
	tests := []struct {
		name         string
		ip           string
		storageError error
		expectError  bool
	}{
		{
			name:        "successful IP storage",
			ip:          "192.168.1.1",
			expectError: false,
		},
		{
			name:         "storage error",
			ip:           "192.168.1.1",
			storageError: errStorage,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var storedIP string
			mockStore := &mockStorage{
				writeIPFunc: func(ip string) error {
					storedIP = ip
					return tt.storageError
				},
			}

			service := &Service{storage: mockStore}
			err := service.storeIP(tt.ip)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && storedIP != tt.ip {
				t.Errorf("expected stored IP %q, got %q", tt.ip, storedIP)
			}
		})
	}
}

func TestNew(t *testing.T) {
	mockDNS := &mockDNSProvider{}
	mockIP := &mockIPClient{}
	mockStore := &mockStorage{}
	config := Config{Zone: "example.com", RecordName: "test", TTL: 60 * time.Second}
	interval := time.Minute

	service := New(mockDNS, mockIP, mockStore, config, interval)

	if service == nil {
		t.Fatal("expected non-nil service")
	}
	if service.dnsProvider != mockDNS {
		t.Error("DNS provider not properly set")
	}
	if service.ipClient != mockIP {
		t.Error("IP client not properly set")
	}
	if service.storage != mockStore {
		t.Error("storage not properly set")
	}
	if service.config != config {
		t.Error("config not properly set")
	}
	if service.interval != interval {
		t.Error("interval not properly set")
	}
}