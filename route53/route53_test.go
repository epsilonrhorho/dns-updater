package route53

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)


type mockRoute53Client struct {
	changeResourceRecordSetsFunc func(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
}

func (m *mockRoute53Client) ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
	if m.changeResourceRecordSetsFunc != nil {
		return m.changeResourceRecordSetsFunc(ctx, params, optFns...)
	}
	return &route53.ChangeResourceRecordSetsOutput{}, nil
}

func TestClient_UpdateARecord(t *testing.T) {
	tests := []struct {
		name           string
		hostedZoneID   string
		recordName     string
		ip             string
		ttl            int64
		mockFunc       func(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
		expectError    bool
		validateParams bool
	}{
		{
			name:         "successful update",
			hostedZoneID: "Z123456789",
			recordName:   "example.com",
			ip:           "192.168.1.1",
			ttl:          60,
			mockFunc: func(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
				return &route53.ChangeResourceRecordSetsOutput{
					ChangeInfo: &types.ChangeInfo{
						Id:     aws.String("C123456789"),
						Status: types.ChangeStatusInsync,
					},
				}, nil
			},
			expectError:    false,
			validateParams: true,
		},
		{
			name:         "API error",
			hostedZoneID: "Z123456789",
			recordName:   "example.com",
			ip:           "192.168.1.1",
			ttl:          60,
			mockFunc: func(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
				return nil, errors.New("Route53 API error")
			},
			expectError:    true,
			validateParams: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockRoute53Client{
				changeResourceRecordSetsFunc: tt.mockFunc,
			}

			var capturedParams *route53.ChangeResourceRecordSetsInput
			if tt.validateParams {
				mockClient.changeResourceRecordSetsFunc = func(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
					capturedParams = params
					return tt.mockFunc(ctx, params, optFns...)
				}
			}

			client := NewClient(mockClient)
			err := client.UpdateARecord(context.Background(), tt.hostedZoneID, tt.recordName, tt.ip, tt.ttl)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.validateParams && capturedParams != nil {
				if *capturedParams.HostedZoneId != tt.hostedZoneID {
					t.Errorf("expected hosted zone ID %q, got %q", tt.hostedZoneID, *capturedParams.HostedZoneId)
				}

				if len(capturedParams.ChangeBatch.Changes) != 1 {
					t.Fatalf("expected 1 change, got %d", len(capturedParams.ChangeBatch.Changes))
				}

				change := capturedParams.ChangeBatch.Changes[0]
				if change.Action != types.ChangeActionUpsert {
					t.Errorf("expected action %v, got %v", types.ChangeActionUpsert, change.Action)
				}

				recordSet := change.ResourceRecordSet
				if *recordSet.Name != tt.recordName {
					t.Errorf("expected record name %q, got %q", tt.recordName, *recordSet.Name)
				}
				if recordSet.Type != types.RRTypeA {
					t.Errorf("expected record type %v, got %v", types.RRTypeA, recordSet.Type)
				}
				if *recordSet.TTL != tt.ttl {
					t.Errorf("expected TTL %d, got %d", tt.ttl, *recordSet.TTL)
				}
				if len(recordSet.ResourceRecords) != 1 {
					t.Fatalf("expected 1 resource record, got %d", len(recordSet.ResourceRecords))
				}
				if *recordSet.ResourceRecords[0].Value != tt.ip {
					t.Errorf("expected IP %q, got %q", tt.ip, *recordSet.ResourceRecords[0].Value)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	mockClient := &mockRoute53Client{}
	client := NewClient(mockClient)

	if client == nil {
		t.Error("expected non-nil client")
	}
}