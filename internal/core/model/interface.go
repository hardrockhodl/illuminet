package model

import "time"

// AdminState is the administratively configured state of an Interface.
// The zero value is AdminStateUnknown.
type AdminState string

// AdminState enumeration values.
const (
	// AdminStateUnknown indicates the admin state could not be
	// determined.
	AdminStateUnknown AdminState = ""
	// AdminStateUp indicates the interface is administratively
	// enabled.
	AdminStateUp AdminState = "up"
	// AdminStateDown indicates the interface is administratively
	// disabled (shutdown).
	AdminStateDown AdminState = "down"
)

// String returns the human-readable admin state. The zero value
// renders as "unknown".
func (s AdminState) String() string {
	if s == AdminStateUnknown {
		return "unknown"
	}
	return string(s)
}

// OperState is the observed operational state of an Interface. The
// zero value is OperStateUnknown.
type OperState string

// OperState enumeration values.
const (
	// OperStateUnknown indicates the operational state could not be
	// determined.
	OperStateUnknown OperState = ""
	// OperStateUp indicates the link is up and forwarding.
	OperStateUp OperState = "up"
	// OperStateDown indicates the link is down.
	OperStateDown OperState = "down"
	// OperStateTesting indicates the link is in a test or diagnostic
	// state and is not carrying production traffic.
	OperStateTesting OperState = "testing"
)

// String returns the human-readable operational state. The zero value
// renders as "unknown".
func (s OperState) String() string {
	if s == OperStateUnknown {
		return "unknown"
	}
	return string(s)
}

// OperMode is the forwarding mode of an Interface. The zero value is
// OperModeUnknown.
type OperMode string

// OperMode enumeration values.
const (
	// OperModeUnknown indicates the forwarding mode could not be
	// determined.
	OperModeUnknown OperMode = ""
	// OperModeRouted indicates a Layer 3 routed port.
	OperModeRouted OperMode = "routed"
	// OperModeAccess indicates a Layer 2 access port (single
	// untagged VLAN).
	OperModeAccess OperMode = "access"
	// OperModeTrunk indicates a Layer 2 trunk port (802.1Q tagged).
	OperModeTrunk OperMode = "trunk"
	// OperModeFEXFabric indicates a fabric-extender uplink port
	// (Cisco FEX). Reserved for platforms where this distinction
	// matters.
	OperModeFEXFabric OperMode = "fex-fabric"
	// OperModeOther indicates a forwarding mode that does not fit
	// the other values (e.g. private-VLAN promiscuous).
	OperModeOther OperMode = "other"
)

// String returns the human-readable operational mode. The zero value
// renders as "unknown".
func (m OperMode) String() string {
	if m == OperModeUnknown {
		return "unknown"
	}
	return string(m)
}

// PortClassification is an operator-intent label describing the role
// of an Interface within the fabric. It is set by enrichment, not by
// the adapter, and is independent of any vendor metadata.
type PortClassification string

// PortClassification enumeration values.
const (
	// PortClassificationUnknown indicates the port has not been
	// classified.
	PortClassificationUnknown PortClassification = ""
	// PortClassificationCore indicates a fabric-facing port that
	// carries inter-switch traffic.
	PortClassificationCore PortClassification = "core"
	// PortClassificationEdge indicates a host-facing or
	// service-facing port.
	PortClassificationEdge PortClassification = "edge"
	// PortClassificationManagement indicates a port used for
	// out-of-band management rather than data plane.
	PortClassificationManagement PortClassification = "management"
	// PortClassificationOther covers ports whose peer is reachable
	// but does not advertise any LLDP capability (no router, bridge,
	// or station), e.g. some firewalls, load balancers, or
	// passive optical equipment.
	PortClassificationOther PortClassification = "other"
)

// String returns the human-readable port classification. The zero
// value renders as "unknown".
func (c PortClassification) String() string {
	if c == PortClassificationUnknown {
		return "unknown"
	}
	return string(c)
}

// Interface is a physical or logical port on a Device. Vendor
// adapters populate AdminState, OperState, counters and peer data.
// Enrichment populates Classification. The interface name is
// preserved in its vendor-native form (e.g. "Ethernet1/1",
// "ge-0/0/0").
type Interface struct {
	// Name is the platform-native interface name. Not normalized.
	Name string

	// Description is the operator-supplied interface description, as
	// configured on the device. Empty when none is configured.
	Description string

	// AdminState is the administratively configured state.
	AdminState AdminState

	// OperState is the observed operational state.
	OperState OperState

	// OperMode is the forwarding mode at sample time.
	OperMode OperMode

	// OperSpeed is the negotiated link speed in bits per second. Nil
	// when the platform did not report it, including the common case
	// of an OperState=Down link with no negotiated rate.
	OperSpeed *uint64

	// DownReason is the vendor-supplied reason string for an
	// OperState=Down link (e.g. "Link not connected", "SFP absent").
	// Empty when the link is up or the platform did not report a
	// reason.
	DownReason string

	// Classification is the operator-intent role of this port. Set by
	// enrichment, not by the adapter.
	Classification PortClassification

	// Peer holds the LLDP-discovered neighbor on this interface. Nil
	// when no neighbor has been observed.
	Peer *Peer

	// Counters holds the link-layer counter snapshot. Nil when the
	// platform did not report any counters for the interface.
	Counters *InterfaceCounters

	// Queues holds the per-queue snapshots for this interface. May be
	// empty when no queue data is available.
	Queues []Queue

	// Buffer holds the per-interface buffer occupancy counters. Nil
	// when the platform does not expose per-interface buffer state;
	// such platforms expose buffer telemetry only at the BufferInstance
	// (ASIC slice) level on the parent Sample.
	Buffer *BufferCounters

	// ObservedAt is the timestamp at which this interface sample was
	// taken. The adapter prefers the device's own clock when the
	// platform exposes a sample timestamp; otherwise the collector's
	// wall clock at the moment the response was received. A zero
	// time.Time means the timestamp is not available.
	ObservedAt time.Time
}
