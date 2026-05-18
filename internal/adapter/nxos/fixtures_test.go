package nxos

import (
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
)

// Spec called for a separate testdata/ directory of fixture builders,
// but Go's tooling excludes directories named "testdata" from package
// resolution, so fixture builders that need to be importable live as
// test helpers in this file instead. The intent (reusable fixture
// constructors) is preserved.

func ifacePath(ifaceName string, leaves ...string) *gnmi.Path {
	elems := []*gnmi.PathElem{
		{Name: "interfaces"},
		{Name: "interface", Key: map[string]string{"name": ifaceName}},
		{Name: "state"},
	}
	for _, l := range leaves {
		elems = append(elems, &gnmi.PathElem{Name: l})
	}
	return &gnmi.Path{Elem: elems}
}

func systemPath(leaves ...string) *gnmi.Path {
	elems := []*gnmi.PathElem{{Name: "system"}, {Name: "state"}}
	for _, l := range leaves {
		elems = append(elems, &gnmi.PathElem{Name: l})
	}
	return &gnmi.Path{Elem: elems}
}

func uintTV(v uint64) *gnmi.TypedValue {
	return &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: v}}
}

func stringTV(v string) *gnmi.TypedValue {
	return &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: v}}
}

// sampleInterfaceCounters builds a Notification carrying the standard
// in/out counters for one interface.
func sampleInterfaceCounters(ifaceName string, in, out uint64, ts time.Time) *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: ts.UnixNano(),
		Update: []*gnmi.Update{
			{Path: ifacePath(ifaceName, "counters", "in-octets"), Val: uintTV(in)},
			{Path: ifacePath(ifaceName, "counters", "out-octets"), Val: uintTV(out)},
		},
	}
}

// sampleInterfaceState builds a Notification carrying admin-status,
// oper-status and description for one interface.
func sampleInterfaceState(ifaceName, admin, oper, desc string, ts time.Time) *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: ts.UnixNano(),
		Update: []*gnmi.Update{
			{Path: ifacePath(ifaceName, "admin-status"), Val: stringTV(admin)},
			{Path: ifacePath(ifaceName, "oper-status"), Val: stringTV(oper)},
			{Path: ifacePath(ifaceName, "description"), Val: stringTV(desc)},
		},
	}
}

// queuePath returns the absolute path to a leaf under one queue's
// state container, using the openconfig-qos list keys (interface-id
// on /qos/interfaces/interface, name on /queues/queue).
func queuePath(intfName, queueName string, leaves ...string) *gnmi.Path {
	elems := []*gnmi.PathElem{
		{Name: "qos"},
		{Name: "interfaces"},
		{Name: "interface", Key: map[string]string{"interface-id": intfName}},
		{Name: "output"},
		{Name: "queues"},
		{Name: "queue", Key: map[string]string{"name": queueName}},
		{Name: "state"},
	}
	for _, l := range leaves {
		elems = append(elems, &gnmi.PathElem{Name: l})
	}
	return &gnmi.Path{Elem: elems}
}

// sampleQueueCounters builds a Notification carrying the four standard
// queue counters (transmit-pkts, transmit-octets, dropped-pkts,
// dropped-octets) for one queue under one interface.
func sampleQueueCounters(intfName, queueName string,
	txPkts, txOctets, droppedPkts, droppedOctets uint64,
	ts time.Time,
) *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: ts.UnixNano(),
		Update: []*gnmi.Update{
			{Path: queuePath(intfName, queueName, "transmit-pkts"), Val: uintTV(txPkts)},
			{Path: queuePath(intfName, queueName, "transmit-octets"), Val: uintTV(txOctets)},
			{Path: queuePath(intfName, queueName, "dropped-pkts"), Val: uintTV(droppedPkts)},
			{Path: queuePath(intfName, queueName, "dropped-octets"), Val: uintTV(droppedOctets)},
		},
	}
}

// sampleQueueECN builds a Notification carrying ecn-marked-pkts for one
// queue under one interface.
func sampleQueueECN(intfName, queueName string, marked uint64, ts time.Time) *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: ts.UnixNano(),
		Update: []*gnmi.Update{
			{Path: queuePath(intfName, queueName, "ecn-marked-pkts"), Val: uintTV(marked)},
		},
	}
}

// sampleSystemState builds a Notification carrying hostname and
// software-version.
func sampleSystemState(hostname, osVersion string, ts time.Time) *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: ts.UnixNano(),
		Update: []*gnmi.Update{
			{Path: systemPath("hostname"), Val: stringTV(hostname)},
			{Path: systemPath("software-version"), Val: stringTV(osVersion)},
		},
	}
}
