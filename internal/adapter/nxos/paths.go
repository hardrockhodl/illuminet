// Paths in this file follow the OpenConfig YANG models. NX-OS 10.4+
// supports openconfig-interfaces and openconfig-system at a minimum;
// the paths below are valid against those models.
//
// For Cisco-native paths (where OpenConfig coverage is incomplete --
// queue counters, PFC, buffer occupancy, BFD telemetry) a separate
// native_paths.go will be added in a follow-up iteration.

package nxos

const (
	// PathInterfaces returns the full operational state subtree for
	// every interface: counters, admin-status, oper-status,
	// description.
	PathInterfaces = "/interfaces/interface[name=*]/state"

	// PathInterfaceCounters returns just the counters subtree per
	// interface. Use this for high-frequency subscriptions where the
	// rest of the state subtree is redundant.
	PathInterfaceCounters = "/interfaces/interface[name=*]/state/counters"

	// PathSystem returns hostname, software version and related
	// device-level state.
	PathSystem = "/system/state"

	// PathComponents returns chassis and module information used to
	// populate Device.Model.
	PathComponents = "/components/component[name=*]/state"

	// PathLLDPNeighbors returns LLDP neighbor state per interface.
	// Each neighbor advertises system-name, system-description,
	// management-address, port-id, chassis-id and system-capabilities
	// at minimum. NX-OS supports openconfig-lldp from 10.4+.
	PathLLDPNeighbors = "/lldp/interfaces/interface[name=*]/neighbors/neighbor[id=*]/state"

	// PathQoSQueueState returns operational queue state per interface,
	// including transmit and drop counters and ECN marked packets.
	//
	// The /qos/interfaces/interface list uses interface-id as its key,
	// not name like the /interfaces/interface list does. The key value
	// is the same interface string on NX-OS; only the key name
	// differs between the two YANG modules.
	PathQoSQueueState = "/qos/interfaces/interface[interface-id=*]/output/queues/queue[name=*]/state"
)
