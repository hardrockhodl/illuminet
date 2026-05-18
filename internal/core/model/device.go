package model

import "time"

// DeviceRole is the operational role a Device plays in the fabric. The
// zero value is DeviceRoleUnknown.
type DeviceRole string

// DeviceRole enumeration values. The empty string is treated as
// Unknown so that a zero-value Device naturally has an Unknown role.
const (
	// DeviceRoleUnknown indicates the role has not been classified.
	DeviceRoleUnknown DeviceRole = ""
	// DeviceRoleSpine indicates a spine switch in a Clos fabric.
	DeviceRoleSpine DeviceRole = "spine"
	// DeviceRoleLeaf indicates a leaf (top-of-rack) switch.
	DeviceRoleLeaf DeviceRole = "leaf"
	// DeviceRoleBorder indicates a border leaf bridging the fabric to
	// an external routing domain.
	DeviceRoleBorder DeviceRole = "border"
	// DeviceRoleEdge indicates an edge / aggregation device facing
	// hosts or the WAN rather than the fabric core.
	DeviceRoleEdge DeviceRole = "edge"
)

// String returns the human-readable role name. The zero value renders
// as "unknown".
func (r DeviceRole) String() string {
	if r == DeviceRoleUnknown {
		return "unknown"
	}
	return string(r)
}

// Device is a single switch or router that the collector samples.
// Fields populated by vendor adapters describe the platform; fields
// populated by enrichment describe the operator's intent (Role,
// Location).
type Device struct {
	// Name is the switch hostname as reported by the platform.
	Name string

	// ManagementIP is the address the collector uses to reach the
	// device. It is recorded for diagnostics; it is not used as a
	// stable identifier downstream.
	ManagementIP string

	// Vendor is the lowercase vendor identifier, e.g. "cisco",
	// "arista", "juniper". Adapters set this consistently.
	Vendor string

	// Model is the raw platform model string reported by the vendor,
	// e.g. "N9K-C9332D-GX2B". Not normalized.
	Model string

	// OSVersion is the raw OS version string reported by the vendor,
	// e.g. "10.5(1)" for NX-OS.
	OSVersion string

	// KernelUptime is the time since the device last booted. Nil when
	// the platform did not report uptime.
	KernelUptime *time.Duration

	// Location is an operator-supplied tag (rack, room, site). It
	// corresponds to the [Location] section in the legacy NTM input
	// file. Empty when not provided.
	Location string

	// Role classifies the device's place in the fabric. Set by
	// enrichment, not by the adapter.
	Role DeviceRole

	// CPUKernel is the kernel-space CPU utilization in percent (0-100).
	// Nil when the platform did not report it.
	CPUKernel *float64

	// CPUUser is the user-space CPU utilization in percent (0-100).
	// Nil when the platform did not report it.
	CPUUser *float64

	// MemoryTotal is the total system memory in bytes. Nil when the
	// platform did not report it.
	MemoryTotal *uint64

	// MemoryUsed is the currently used system memory in bytes. Nil
	// when the platform did not report it.
	MemoryUsed *uint64

	// ResponseTime is the collector's end-to-end wall-clock duration
	// spent sampling this device for the current tick. Nil when the
	// duration was not measured.
	ResponseTime *time.Duration
}
