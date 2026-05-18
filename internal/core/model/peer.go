package model

// PeerType classifies a discovered neighbor. The zero value is
// PeerTypeUnknown.
type PeerType string

// PeerType enumeration values. The empty string is treated as
// Unknown so that a zero-value Peer naturally has an Unknown type.
const (
	// PeerTypeUnknown indicates the neighbor could not be classified.
	PeerTypeUnknown PeerType = ""
	// PeerTypeHost indicates the neighbor is a server, hypervisor or
	// other endpoint rather than a network device.
	PeerTypeHost PeerType = "host"
	// PeerTypeSwitch indicates the neighbor is another switch or
	// router participating in the fabric.
	PeerTypeSwitch PeerType = "switch"
	// PeerTypeOther covers neighbors that are neither hosts nor
	// switches in the fabric topology (firewalls, load balancers,
	// optical equipment, etc.).
	PeerTypeOther PeerType = "other"
)

// String returns the human-readable peer type. The zero value renders
// as "unknown".
func (t PeerType) String() string {
	if t == PeerTypeUnknown {
		return "unknown"
	}
	return string(t)
}

// Peer is a neighbor discovered on an Interface, typically via LLDP.
// A nil *Peer on an Interface means no neighbor has been observed.
type Peer struct {
	// Name is the neighbor's system name as advertised over LLDP.
	Name string

	// Interface is the neighbor's port identifier as reported by LLDP
	// (its local interface name on the far side of the link).
	Interface string

	// MgmtIP is the management address advertised by the neighbor.
	// Empty when not advertised.
	MgmtIP string

	// Type classifies the neighbor. Adapters typically leave this at
	// PeerTypeUnknown; downstream pipeline stages derive Type from
	// Capabilities and SystemDescription.
	Type PeerType

	// LearnedVia names the discovery protocol that produced this Peer
	// record. Currently always "lldp"; reserved for CDP or other
	// sources in the future.
	LearnedVia string

	// Capabilities lists the LLDP system capabilities advertised by
	// the peer ("router", "bridge", "station", "wlan-access-point",
	// etc.). Empty when the peer did not advertise capabilities or
	// was discovered through a protocol that does not carry them.
	// Used by classification stages downstream.
	Capabilities []string

	// SystemDescription is the LLDP system-description string. Used
	// by classification stages to disambiguate hosts from switches
	// when capabilities alone are insufficient (e.g. Linux servers
	// that advertise bridge/router capability because of container
	// networking).
	SystemDescription string
}
