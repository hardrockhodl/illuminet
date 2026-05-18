package stages

import (
	"context"
	"log/slog"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// PortClassificationStage examines each Interface's Peer.Type and
// sets Interface.Classification accordingly.
//
// Mapping:
//
//	PeerType=Switch  -> Classification=Core
//	PeerType=Host    -> Classification=Edge
//	PeerType=Other   -> Classification=Other
//	PeerType=Unknown -> Classification=Unknown
//	No peer at all   -> Classification=Unknown
//
// Order matters: run this stage AFTER [PeerClassificationStage] so
// Peer.Type has been derived from LLDP data.
type PortClassificationStage struct {
	logger *slog.Logger
}

// NewPortClassification constructs the stage. A nil logger is
// replaced by slog.Default.
func NewPortClassification(logger *slog.Logger) *PortClassificationStage {
	if logger == nil {
		logger = slog.Default()
	}
	return &PortClassificationStage{logger: logger}
}

// Name returns "port-classification".
func (s *PortClassificationStage) Name() string { return "port-classification" }

// Process sets Interface.Classification for every Interface in the
// Sample. It never returns an error during normal operation.
func (s *PortClassificationStage) Process(_ context.Context, sample *model.Sample) error {
	for i := range sample.Interfaces {
		sample.Interfaces[i].Classification = classifyPort(sample.Interfaces[i].Peer)
	}
	return nil
}

func classifyPort(peer *model.Peer) model.PortClassification {
	if peer == nil {
		return model.PortClassificationUnknown
	}
	switch peer.Type {
	case model.PeerTypeSwitch:
		return model.PortClassificationCore
	case model.PeerTypeHost:
		return model.PortClassificationEdge
	case model.PeerTypeOther:
		return model.PortClassificationOther
	default:
		return model.PortClassificationUnknown
	}
}
