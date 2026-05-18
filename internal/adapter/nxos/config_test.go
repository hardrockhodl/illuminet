package nxos

import (
	"strings"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "empty Address",
			cfg:     Config{Username: "u", Password: "p"},
			wantErr: "Address is required",
		},
		{
			name:    "Address without port",
			cfg:     Config{Address: "router.example.com", Username: "u", Password: "p"},
			wantErr: "must include a port",
		},
		{
			name:    "empty Username",
			cfg:     Config{Address: "h:50051", Password: "p"},
			wantErr: "Username is required",
		},
		{
			name:    "empty Password without Insecure",
			cfg:     Config{Address: "h:50051", Username: "u"},
			wantErr: "Password is required",
		},
		{
			name:    "empty Password with Insecure",
			cfg:     Config{Address: "h:50051", Username: "u", Insecure: true},
			wantErr: "",
		},
		{
			name:    "negative Interval",
			cfg:     Config{Address: "h:50051", Username: "u", Password: "p", Interval: -1},
			wantErr: "Interval must be non-negative",
		},
		{
			name:    "negative Timeout",
			cfg:     Config{Address: "h:50051", Username: "u", Password: "p", Timeout: -1},
			wantErr: "Timeout must be non-negative",
		},
		{
			name:    "valid",
			cfg:     Config{Address: "h:50051", Username: "u", Password: "p"},
			wantErr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			switch {
			case tc.wantErr == "" && err != nil:
				t.Errorf("expected no error, got %v", err)
			case tc.wantErr != "" && err == nil:
				t.Errorf("expected error containing %q, got nil", tc.wantErr)
			case tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr):
				t.Errorf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestConfig_applyDefaults(t *testing.T) {
	cfg := Config{Address: "h:50051", Username: "u", Password: "p"}
	cfg.applyDefaults()
	if cfg.Interval != 10*time.Second {
		t.Errorf("Interval default: got %v, want 10s", cfg.Interval)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout default: got %v, want 10s", cfg.Timeout)
	}
}

func TestConfig_applyDefaults_preservesNonZero(t *testing.T) {
	cfg := Config{
		Address:  "h:50051",
		Username: "u",
		Password: "p",
		Interval: 3 * time.Second,
		Timeout:  7 * time.Second,
	}
	cfg.applyDefaults()
	if cfg.Interval != 3*time.Second {
		t.Errorf("Interval: got %v, want preserved 3s", cfg.Interval)
	}
	if cfg.Timeout != 7*time.Second {
		t.Errorf("Timeout: got %v, want preserved 7s", cfg.Timeout)
	}
}
