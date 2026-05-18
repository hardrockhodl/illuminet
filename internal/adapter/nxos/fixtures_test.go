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
