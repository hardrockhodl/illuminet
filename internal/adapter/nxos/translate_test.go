package nxos

import (
	"testing"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

func TestTranslate_SingleInterfaceCounters(t *testing.T) {
	ts := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	devCtx := DeviceContext{Name: "leaf-01", ManagementIP: "10.0.0.1"}

	notifs := []*gnmi.Notification{
		sampleInterfaceCounters("Ethernet1/1", 1024, 2048, ts),
	}

	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if len(s.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(s.Interfaces))
	}
	iface := s.Interfaces[0]
	if iface.Name != "Ethernet1/1" {
		t.Errorf("Name: got %q, want Ethernet1/1", iface.Name)
	}
	if iface.Counters == nil {
		t.Fatal("Counters: got nil")
	}
	if iface.Counters.RxBytes == nil || *iface.Counters.RxBytes != 1024 {
		t.Errorf("RxBytes: got %v, want 1024", iface.Counters.RxBytes)
	}
	if iface.Counters.TxBytes == nil || *iface.Counters.TxBytes != 2048 {
		t.Errorf("TxBytes: got %v, want 2048", iface.Counters.TxBytes)
	}
	if iface.Counters.RxCRC != nil {
		t.Errorf("RxCRC: got non-nil, want nil (not reported)")
	}
	if !s.CollectedAt.Equal(ts) {
		t.Errorf("CollectedAt: got %v, want %v", s.CollectedAt, ts)
	}
}

func TestTranslate_MultipleInterfaces(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	notifs := []*gnmi.Notification{
		sampleInterfaceCounters("Ethernet1/1", 100, 200, ts),
		sampleInterfaceCounters("Ethernet1/2", 300, 400, ts),
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if len(s.Interfaces) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(s.Interfaces))
	}
	// Sorted by name: Ethernet1/1, Ethernet1/2.
	if s.Interfaces[0].Name != "Ethernet1/1" || s.Interfaces[1].Name != "Ethernet1/2" {
		t.Errorf("unexpected interface order: %q %q", s.Interfaces[0].Name, s.Interfaces[1].Name)
	}
}

func TestTranslate_SystemAndInterfacesCombined(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "ip-10-0-0-1", Location: "lab-rack-A"}

	notifs := []*gnmi.Notification{
		sampleSystemState("leaf-01", "10.5(1)", ts),
		sampleInterfaceCounters("Ethernet1/1", 10, 20, ts),
		sampleInterfaceState("Ethernet1/1", "UP", "UP", "uplink to spine-1", ts),
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}

	// Hostname from /system/state should override the placeholder set
	// from DeviceContext.
	if s.Device.Name != "leaf-01" {
		t.Errorf("Device.Name: got %q, want leaf-01 (from system/state)", s.Device.Name)
	}
	if s.Device.OSVersion != "10.5(1)" {
		t.Errorf("Device.OSVersion: got %q, want 10.5(1)", s.Device.OSVersion)
	}
	if s.Device.Vendor != "cisco" {
		t.Errorf("Device.Vendor: got %q, want cisco", s.Device.Vendor)
	}
	if s.Device.Location != "lab-rack-A" {
		t.Errorf("Device.Location: got %q, want lab-rack-A", s.Device.Location)
	}

	if len(s.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(s.Interfaces))
	}
	iface := s.Interfaces[0]
	if iface.AdminState != model.AdminStateUp {
		t.Errorf("AdminState: got %q, want up", iface.AdminState)
	}
	if iface.OperState != model.OperStateUp {
		t.Errorf("OperState: got %q, want up", iface.OperState)
	}
	if iface.Description != "uplink to spine-1" {
		t.Errorf("Description: got %q", iface.Description)
	}
}

func TestTranslate_UnknownPathSkipped(t *testing.T) {
	devCtx := DeviceContext{Name: "leaf-01"}
	notifs := []*gnmi.Notification{
		{
			Timestamp: time.Now().UnixNano(),
			Update: []*gnmi.Update{
				{
					Path: &gnmi.Path{Elem: []*gnmi.PathElem{
						{Name: "unknown-top-level"},
						{Name: "child"},
					}},
					Val: stringTV("ignored"),
				},
			},
		},
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("unknown path should not error, got %v", err)
	}
	if len(s.Interfaces) != 0 {
		t.Errorf("unknown path produced interfaces: %d", len(s.Interfaces))
	}
}

func TestTranslate_WrongTypedValueReturnsError(t *testing.T) {
	devCtx := DeviceContext{Name: "leaf-01"}
	notifs := []*gnmi.Notification{
		{
			Timestamp: time.Now().UnixNano(),
			Update: []*gnmi.Update{
				{
					Path: ifacePath("Ethernet1/1", "counters", "in-octets"),
					Val:  stringTV("not-a-number"),
				},
			},
		},
	}
	if _, err := Translate(devCtx, notifs); err == nil {
		t.Fatal("expected error from string-where-uint, got nil")
	}
}

func TestTranslate_MissingFieldsStayNil(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	// Notification carries RxBytes only; everything else must remain
	// nil/zero per the model's "nil means not reported" contract.
	notifs := []*gnmi.Notification{
		{
			Timestamp: ts.UnixNano(),
			Update: []*gnmi.Update{
				{Path: ifacePath("Ethernet1/1", "counters", "in-octets"), Val: uintTV(42)},
			},
		},
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	iface := s.Interfaces[0]
	if iface.Counters.RxBytes == nil || *iface.Counters.RxBytes != 42 {
		t.Errorf("RxBytes: got %v, want 42", iface.Counters.RxBytes)
	}
	if iface.Counters.TxBytes != nil {
		t.Error("TxBytes: got non-nil, want nil")
	}
	if iface.Counters.RxUcastPkts != nil {
		t.Error("RxUcastPkts: got non-nil, want nil")
	}
	if iface.AdminState != model.AdminStateUnknown {
		t.Errorf("AdminState: got %q, want unknown", iface.AdminState)
	}
}

func TestTranslate_LastClearParsesRFC3339(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}
	clear := "2026-05-18T11:00:00Z"

	notifs := []*gnmi.Notification{
		{
			Timestamp: ts.UnixNano(),
			Update: []*gnmi.Update{
				{Path: ifacePath("Ethernet1/1", "counters", "in-octets"), Val: uintTV(1)},
				{Path: ifacePath("Ethernet1/1", "counters", "last-clear"), Val: stringTV(clear)},
			},
		},
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	got := s.Interfaces[0].Counters.LastClear
	if got == nil {
		t.Fatal("LastClear: got nil")
	}
	want, _ := time.Parse(time.RFC3339, clear)
	if !got.Equal(want) {
		t.Errorf("LastClear: got %v, want %v", got, want)
	}
}

func TestTranslate_AllInterfaceCounters(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	leaves := map[string]uint64{
		"in-octets":           100,
		"out-octets":          200,
		"in-unicast-pkts":     300,
		"out-unicast-pkts":    400,
		"in-multicast-pkts":   500,
		"out-multicast-pkts":  600,
		"in-broadcast-pkts":   700,
		"out-broadcast-pkts":  800,
		"in-crc-errors":       9,
	}

	updates := make([]*gnmi.Update, 0, len(leaves))
	for leaf, v := range leaves {
		updates = append(updates, &gnmi.Update{
			Path: ifacePath("Ethernet1/1", "counters", leaf),
			Val:  uintTV(v),
		})
	}
	notifs := []*gnmi.Notification{{Timestamp: ts.UnixNano(), Update: updates}}

	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	c := s.Interfaces[0].Counters
	type check struct {
		name string
		ptr  *uint64
		want uint64
	}
	for _, ch := range []check{
		{"RxBytes", c.RxBytes, 100},
		{"TxBytes", c.TxBytes, 200},
		{"RxUcastPkts", c.RxUcastPkts, 300},
		{"TxUcastPkts", c.TxUcastPkts, 400},
		{"RxMcastPkts", c.RxMcastPkts, 500},
		{"TxMcastPkts", c.TxMcastPkts, 600},
		{"RxBcastPkts", c.RxBcastPkts, 700},
		{"TxBcastPkts", c.TxBcastPkts, 800},
		{"RxCRC", c.RxCRC, 9},
	} {
		if ch.ptr == nil || *ch.ptr != ch.want {
			t.Errorf("%s: got %v, want %d", ch.name, ch.ptr, ch.want)
		}
	}
}
