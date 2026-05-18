package stages

import (
	"context"
	"testing"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

func TestPortClassification_Mapping(t *testing.T) {
	cases := []struct {
		name     string
		peerType model.PeerType
		want     model.PortClassification
	}{
		{"Switch -> Core", model.PeerTypeSwitch, model.PortClassificationCore},
		{"Host -> Edge", model.PeerTypeHost, model.PortClassificationEdge},
		{"Other -> Other", model.PeerTypeOther, model.PortClassificationOther},
		{"Unknown -> Unknown", model.PeerTypeUnknown, model.PortClassificationUnknown},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewPortClassification(nil)
			sample := &model.Sample{
				Interfaces: []model.Interface{{
					Name: "Ethernet1/1",
					Peer: &model.Peer{Type: tc.peerType},
				}},
			}
			if err := s.Process(context.Background(), sample); err != nil {
				t.Fatalf("Process: %v", err)
			}
			if got := sample.Interfaces[0].Classification; got != tc.want {
				t.Errorf("Classification: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPortClassification_NilPeerYieldsUnknown(t *testing.T) {
	s := NewPortClassification(nil)
	sample := &model.Sample{
		Interfaces: []model.Interface{{Name: "Ethernet1/1", Peer: nil}},
	}
	if err := s.Process(context.Background(), sample); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if got := sample.Interfaces[0].Classification; got != model.PortClassificationUnknown {
		t.Errorf("Classification: got %q, want unknown", got)
	}
}

func TestPortClassification_Name(t *testing.T) {
	if got := NewPortClassification(nil).Name(); got != "port-classification" {
		t.Errorf("Name: got %q", got)
	}
}
