package main

import (
	"testing"
)

func TestExtractZoneFromRecordName(t *testing.T) {
	tests := []struct {
		name        string
		recordName  string
		expectedZone string
		expectError bool
	}{
		{
			name:        "valid 3-label record",
			recordName:  "foo.example.com",
			expectedZone: "example.com",
			expectError: false,
		},
		{
			name:        "valid 4-label record",
			recordName:  "foo.bar.example.com",
			expectedZone: "bar.example.com",
			expectError: false,
		},
		{
			name:        "valid 5-label record",
			recordName:  "api.v1.prod.example.com",
			expectedZone: "v1.prod.example.com",
			expectError: false,
		},
		{
			name:        "2-label record should error",
			recordName:  "example.com",
			expectError: true,
		},
		{
			name:        "1-label record should error",
			recordName:  "localhost",
			expectError: true,
		},
		{
			name:        "empty string should error",
			recordName:  "",
			expectError: true,
		},
		{
			name:        "subdomain case",
			recordName:  "web.staging.example.org",
			expectedZone: "staging.example.org",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zone, err := extractZoneFromRecordName(tt.recordName)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for record name %q, but got none", tt.recordName)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for record name %q: %v", tt.recordName, err)
				return
			}

			if zone != tt.expectedZone {
				t.Errorf("for record name %q, expected zone %q, got %q", tt.recordName, tt.expectedZone, zone)
			}
		})
	}
}