package stages

import (
	"context"
	"log/slog"
	"strings"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// PeerClassificationStage examines each Interface's Peer and sets
// Peer.Type based on LLDP capabilities and system description.
//
// Classification rules, in priority order:
//
//  1. No peer learned: Type remains Unknown.
//  2. Capabilities include "router" or "bridge" AND SystemDescription
//     does NOT contain "Linux": Switch.
//  3. Capabilities include "router" or "bridge" AND SystemDescription
//     contains "Linux": Host. This handles the Docker-on-Linux quirk
//     where Linux hosts advertise router or bridge capability because
//     of container networking.
//  4. Capabilities include any value other than router and bridge
//     (e.g. "station", "wlan-access-point"): Host.
//  5. No capabilities advertised: Other.
//
// The stage is safe for concurrent use and has no external
// dependencies.
type PeerClassificationStage struct {
	logger *slog.Logger
}

// NewPeerClassification constructs the stage. A nil logger is
// replaced by slog.Default.
func NewPeerClassification(logger *slog.Logger) *PeerClassificationStage {
	if logger == nil {
		logger = slog.Default()
	}
	return &PeerClassificationStage{logger: logger}
}

// Name returns "peer-classification".
func (s *PeerClassificationStage) Name() string { return "peer-classification" }

// Process classifies every Interface.Peer in the Sample. It never
// returns an error during normal operation.
func (s *PeerClassificationStage) Process(_ context.Context, sample *model.Sample) error {
	for i := range sample.Interfaces {
		peer := sample.Interfaces[i].Peer
		if peer == nil {
			continue
		}
		peer.Type = classifyPeer(peer.Capabilities, peer.SystemDescription)
	}
	return nil
}

func classifyPeer(capabilities []string, sysDesc string) model.PeerType {
	if len(capabilities) == 0 {
		return model.PeerTypeOther
	}
	if hasRouterOrBridge(capabilities) {
		if strings.Contains(sysDesc, "Linux") {
			return model.PeerTypeHost
		}
		return model.PeerTypeSwitch
	}
	return model.PeerTypeHost
}

func hasRouterOrBridge(caps []string) bool {
	for _, c := range caps {
		if c == "router" || c == "bridge" {
			return true
		}
	}
	return false
}
