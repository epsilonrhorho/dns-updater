package service

import (
	"context"
	"errors"
	"time"
)

type mockDNSProvider struct {
	updateARecordFunc func(ctx context.Context, zone, name, ip string, ttl time.Duration) error
}

func (m *mockDNSProvider) UpdateARecord(ctx context.Context, zone, name, ip string, ttl time.Duration) error {
	if m.updateARecordFunc != nil {
		return m.updateARecordFunc(ctx, zone, name, ip, ttl)
	}
	return nil
}

type mockIPClient struct {
	getIPFunc func(ctx context.Context) (string, error)
}

func (m *mockIPClient) GetIP(ctx context.Context) (string, error) {
	if m.getIPFunc != nil {
		return m.getIPFunc(ctx)
	}
	return "192.168.1.1", nil
}

type mockStorage struct {
	readLastIPFunc func() (string, error)
	writeIPFunc    func(ip string) error
}

func (m *mockStorage) ReadLastIP() (string, error) {
	if m.readLastIPFunc != nil {
		return m.readLastIPFunc()
	}
	return "", nil
}

func (m *mockStorage) WriteIP(ip string) error {
	if m.writeIPFunc != nil {
		return m.writeIPFunc(ip)
	}
	return nil
}

var (
	errDNS     = errors.New("dns provider error")
	errIP      = errors.New("ip client error")
	errStorage = errors.New("storage error")
)