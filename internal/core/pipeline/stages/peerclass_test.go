package stages

import (
	"context"
	"testing"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

func TestPeerClassification_Rules(t *testing.T) {
	cases := []struct {
		name         string
		capabilities []string
		sysDesc      string
		want         model.PeerType
	}{
		{
			name:         "router + Cisco sysDesc -> Switch",
			capabilities: []string{"router"},
			sysDesc:      "Cisco Nexus 9000 N9K-C9332D-GX2B NX-OS 10.5(1)",
			want:         model.PeerTypeSwitch,
		},
		{
			name:         "bridge + Linux sysDesc -> Host (Linux quirk)",
			capabilities: []string{"bridge"},
			sysDesc:      "Linux Ubuntu 22.04 with Docker",
			want:         model.PeerTypeHost,
		},
		{
			name:         "router + Linux sysDesc -> Host (Linux quirk)",
			capabilities: []string{"router"},
			sysDesc:      "Ubuntu Linux server, kernel 6.1",
			want:         model.PeerTypeHost,
		},
		{
			name:         "station -> Host",
			capabilities: []string{"station"},
			sysDesc:      "ESXi 8.0",
			want:         model.PeerTypeHost,
		},
		{
			name:         "empty capabilities -> Other",
			capabilities: nil,
			sysDesc:      "Some firewall",
			want:         model.PeerTypeOther,
		},
		{
			name:         "router + wlan-access-point -> Switch (router wins)",
			capabilities: []string{"router", "wlan-access-point"},
			sysDesc:      "Some AP controller",
			want:         model.PeerTypeSwitch,
		},
		{
			name:         "wlan-access-point only -> Host",
			capabilities: []string{"wlan-access-point"},
			sysDesc:      "Ruckus AP",
			want:         model.PeerTypeHost,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewPeerClassification(nil)
			sample := &model.Sample{
				Interfaces: []model.Interface{{
					Name: "Ethernet1/1",
					Peer: &model.Peer{
						Capabilities:      tc.capabilities,
						SystemDescription: tc.sysDesc,
					},
				}},
			}
			if err := s.Process(context.Background(), sample); err != nil {
				t.Fatalf("Process: %v", err)
			}
			got := sample.Interfaces[0].Peer.Type
			if got != tc.want {
				t.Errorf("Type: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPeerClassification_NilPeerLeavesTypeUnknown(t *testing.T) {
	s := NewPeerClassification(nil)
	sample := &model.Sample{
		Interfaces: []model.Interface{{Name: "Ethernet1/1", Peer: nil}},
	}
	if err := s.Process(context.Background(), sample); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if sample.Interfaces[0].Peer != nil {
		t.Errorf("Peer: got non-nil, want nil")
	}
}

func TestPeerClassification_NilLoggerFallsBackToDefault(t *testing.T) {
	s := NewPeerClassification(nil)
	if s.logger == nil {
		t.Fatal("logger: got nil, want slog.Default fallback")
	}
}

func TestPeerClassification_Name(t *testing.T) {
	if got := NewPeerClassification(nil).Name(); got != "peer-classification" {
		t.Errorf("Name: got %q", got)
	}
}
