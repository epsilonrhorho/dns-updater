package service

import (
	"context"
	"errors"
)

type mockRoute53Client struct {
	updateARecordFunc func(ctx context.Context, hostedZoneID, recordName, ip string, ttl int64) error
}

func (m *mockRoute53Client) UpdateARecord(ctx context.Context, hostedZoneID, recordName, ip string, ttl int64) error {
	if m.updateARecordFunc != nil {
		return m.updateARecordFunc(ctx, hostedZoneID, recordName, ip, ttl)
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
	errRoute53   = errors.New("route53 error")
	errIP        = errors.New("ip client error")
	errStorage   = errors.New("storage error")
)