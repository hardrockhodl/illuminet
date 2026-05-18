package nxos

import (
	"strings"
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
	if s.Device.Vendor != model.VendorCisco {
		t.Errorf("Device.Vendor: got %q, want VendorCisco", s.Device.Vendor)
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
	clearedAt := "2026-05-18T11:00:00Z"

	notifs := []*gnmi.Notification{
		{
			Timestamp: ts.UnixNano(),
			Update: []*gnmi.Update{
				{Path: ifacePath("Ethernet1/1", "counters", "in-octets"), Val: uintTV(1)},
				{Path: ifacePath("Ethernet1/1", "counters", "last-clear"), Val: stringTV(clearedAt)},
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
	want, _ := time.Parse(time.RFC3339, clearedAt)
	if !got.Equal(want) {
		t.Errorf("LastClear: got %v, want %v", got, want)
	}
}

func TestTranslate_SingleQueue(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	notifs := []*gnmi.Notification{
		sampleQueueCounters("Ethernet1/1", "QOS-GROUP-0", 10, 200, 3, 64, ts),
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
		t.Errorf("Name: got %q", iface.Name)
	}
	if len(iface.Queues) != 1 {
		t.Fatalf("expected 1 queue, got %d", len(iface.Queues))
	}
	q := iface.Queues[0]
	if q.Name != "QOS-GROUP-0" {
		t.Errorf("Queue.Name: got %q", q.Name)
	}
	if q.Counters.TxPkts == nil || *q.Counters.TxPkts != 10 {
		t.Errorf("TxPkts: got %v, want 10", q.Counters.TxPkts)
	}
	if q.Counters.TxBytes == nil || *q.Counters.TxBytes != 200 {
		t.Errorf("TxBytes: got %v, want 200", q.Counters.TxBytes)
	}
	if q.Counters.DropPkts == nil || *q.Counters.DropPkts != 3 {
		t.Errorf("DropPkts: got %v, want 3", q.Counters.DropPkts)
	}
	if q.Counters.DropBytes == nil || *q.Counters.DropBytes != 64 {
		t.Errorf("DropBytes: got %v, want 64", q.Counters.DropBytes)
	}
	if q.ECN != nil {
		t.Errorf("ECN: got non-nil, want nil (no ECN data)")
	}
	if q.PFC != nil {
		t.Errorf("PFC: got non-nil, want nil (not in this iteration)")
	}
}

func TestTranslate_MultipleQueuesPerInterface(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	notifs := []*gnmi.Notification{
		sampleQueueCounters("Ethernet1/1", "Q0", 100, 1000, 0, 0, ts),
		sampleQueueCounters("Ethernet1/1", "Q3", 200, 2000, 1, 64, ts),
		sampleQueueCounters("Ethernet1/1", "Q7", 50, 500, 0, 0, ts),
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if len(s.Interfaces) != 1 {
		t.Fatalf("expected 1 interface (not duplicated per queue), got %d", len(s.Interfaces))
	}
	queues := s.Interfaces[0].Queues
	if len(queues) != 3 {
		t.Fatalf("expected 3 queues, got %d", len(queues))
	}
	byName := map[string]model.Queue{}
	for _, q := range queues {
		byName[q.Name] = q
	}
	if q := byName["Q0"]; q.Counters.TxPkts == nil || *q.Counters.TxPkts != 100 {
		t.Errorf("Q0.TxPkts: got %v, want 100", q.Counters.TxPkts)
	}
	if q := byName["Q3"]; q.Counters.DropPkts == nil || *q.Counters.DropPkts != 1 {
		t.Errorf("Q3.DropPkts: got %v, want 1", q.Counters.DropPkts)
	}
	if q := byName["Q7"]; q.Counters.TxBytes == nil || *q.Counters.TxBytes != 500 {
		t.Errorf("Q7.TxBytes: got %v, want 500", q.Counters.TxBytes)
	}
}

func TestTranslate_QueueAndInterfaceCountersMerged(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	// /interfaces/interface[name=Ethernet1/1]/... AND
	// /qos/interfaces/interface[interface-id=Ethernet1/1]/...
	// must land on the same model.Interface.
	notifs := []*gnmi.Notification{
		sampleInterfaceCounters("Ethernet1/1", 1024, 2048, ts),
		sampleQueueCounters("Ethernet1/1", "Q0", 10, 200, 0, 0, ts),
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if len(s.Interfaces) != 1 {
		t.Fatalf("expected 1 interface (interface-id should unify with name), got %d", len(s.Interfaces))
	}
	iface := s.Interfaces[0]
	if iface.Counters == nil {
		t.Fatal("Counters: nil (interface-counter notification was dropped)")
	}
	if iface.Counters.RxBytes == nil || *iface.Counters.RxBytes != 1024 {
		t.Errorf("RxBytes: got %v, want 1024", iface.Counters.RxBytes)
	}
	if len(iface.Queues) != 1 {
		t.Fatalf("expected 1 queue, got %d", len(iface.Queues))
	}
	if iface.Queues[0].Counters.TxPkts == nil || *iface.Queues[0].Counters.TxPkts != 10 {
		t.Errorf("Queue[0].TxPkts: got %v, want 10", iface.Queues[0].Counters.TxPkts)
	}
}

func TestTranslate_ECNMarkedSeparateFromCounters(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	t.Run("ECN only", func(t *testing.T) {
		notifs := []*gnmi.Notification{
			sampleQueueECN("Ethernet1/1", "Q0", 42, ts),
		}
		s, err := Translate(devCtx, notifs)
		if err != nil {
			t.Fatalf("Translate: %v", err)
		}
		q := s.Interfaces[0].Queues[0]
		if q.ECN == nil {
			t.Fatal("ECN: nil")
		}
		if q.ECN.MarkedPkts == nil || *q.ECN.MarkedPkts != 42 {
			t.Errorf("ECN.MarkedPkts: got %v, want 42", q.ECN.MarkedPkts)
		}
		if q.Counters.TxPkts != nil ||
			q.Counters.TxBytes != nil ||
			q.Counters.DropPkts != nil ||
			q.Counters.DropBytes != nil {
			t.Errorf("Counters: expected all nil when only ECN reported, got %+v", q.Counters)
		}
	})

	t.Run("counters only", func(t *testing.T) {
		notifs := []*gnmi.Notification{
			sampleQueueCounters("Ethernet1/1", "Q0", 10, 200, 0, 0, ts),
		}
		s, err := Translate(devCtx, notifs)
		if err != nil {
			t.Fatalf("Translate: %v", err)
		}
		q := s.Interfaces[0].Queues[0]
		if q.ECN != nil {
			t.Errorf("ECN: got non-nil, want nil when no ECN reported")
		}
		if q.Counters.TxPkts == nil || *q.Counters.TxPkts != 10 {
			t.Errorf("TxPkts: got %v, want 10", q.Counters.TxPkts)
		}
	})
}

func TestTranslate_QueueOnUnknownInterface(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	// Queue notification arrives without a prior interface-counter
	// notification; the interface must still be created so the queue
	// has somewhere to live.
	notifs := []*gnmi.Notification{
		sampleQueueCounters("Ethernet1/2", "Q0", 5, 100, 0, 0, ts),
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if len(s.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(s.Interfaces))
	}
	iface := s.Interfaces[0]
	if iface.Name != "Ethernet1/2" {
		t.Errorf("Name: got %q", iface.Name)
	}
	if iface.Counters != nil {
		t.Errorf("Counters: got non-nil, want nil (no interface counters reported)")
	}
	if len(iface.Queues) != 1 {
		t.Fatalf("expected 1 queue, got %d", len(iface.Queues))
	}
}

func TestTranslate_LLDPNeighbor(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	notifs := []*gnmi.Notification{
		sampleLLDPNeighbor("Ethernet1/1", "0",
			"spine-1", "Cisco Nexus 9000 N9K-C9332D-GX2B NX-OS 10.5(1)",
			[]string{"router", "bridge"}, ts),
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if len(s.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(s.Interfaces))
	}
	iface := s.Interfaces[0]
	if iface.Peer == nil {
		t.Fatal("Peer: got nil")
	}
	if iface.Peer.Name != "spine-1" {
		t.Errorf("Peer.Name: got %q, want spine-1", iface.Peer.Name)
	}
	if !strings.Contains(iface.Peer.SystemDescription, "NX-OS") {
		t.Errorf("Peer.SystemDescription: got %q", iface.Peer.SystemDescription)
	}
	if iface.Peer.LearnedVia != "lldp" {
		t.Errorf("Peer.LearnedVia: got %q, want lldp", iface.Peer.LearnedVia)
	}
	if iface.Peer.Type != model.PeerTypeUnknown {
		t.Errorf("Peer.Type: got %q, want Unknown (classification is a pipeline stage)", iface.Peer.Type)
	}
	wantCaps := []string{"router", "bridge"}
	if len(iface.Peer.Capabilities) != 2 ||
		iface.Peer.Capabilities[0] != wantCaps[0] ||
		iface.Peer.Capabilities[1] != wantCaps[1] {
		t.Errorf("Peer.Capabilities: got %v, want %v", iface.Peer.Capabilities, wantCaps)
	}
}

func TestTranslate_LLDPMergedWithInterfaceCounters(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	notifs := []*gnmi.Notification{
		sampleInterfaceCounters("Ethernet1/1", 1024, 2048, ts),
		sampleLLDPNeighbor("Ethernet1/1", "0", "spine-1", "Cisco N9K",
			[]string{"router"}, ts),
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if len(s.Interfaces) != 1 {
		t.Fatalf("expected 1 interface (lldp + counters should merge), got %d", len(s.Interfaces))
	}
	iface := s.Interfaces[0]
	if iface.Counters == nil || iface.Counters.RxBytes == nil || *iface.Counters.RxBytes != 1024 {
		t.Errorf("Counters.RxBytes: got %+v, want 1024", iface.Counters)
	}
	if iface.Peer == nil || iface.Peer.Name != "spine-1" {
		t.Errorf("Peer: got %+v, want Name=spine-1", iface.Peer)
	}
}

func TestTranslate_LLDPPortIDOverridesChassisID(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	notifs := []*gnmi.Notification{
		{
			Timestamp: ts.UnixNano(),
			Update: []*gnmi.Update{
				{Path: lldpNeighborPath("Ethernet1/1", "0", "system-name"), Val: stringTV("spine-1")},
				{Path: lldpNeighborPath("Ethernet1/1", "0", "chassis-id"), Val: stringTV("aa:bb:cc:dd:ee:ff")},
				{Path: lldpNeighborPath("Ethernet1/1", "0", "port-id"), Val: stringTV("Ethernet1/2")},
			},
		},
	}
	s, err := Translate(devCtx, notifs)
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if got := s.Interfaces[0].Peer.Interface; got != "Ethernet1/2" {
		t.Errorf("Peer.Interface: got %q, want Ethernet1/2 (port-id should win over chassis-id)", got)
	}
}

func TestTranslate_AllInterfaceCounters(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	devCtx := DeviceContext{Name: "leaf-01"}

	leaves := map[string]uint64{
		"in-octets":          100,
		"out-octets":         200,
		"in-unicast-pkts":    300,
		"out-unicast-pkts":   400,
		"in-multicast-pkts":  500,
		"out-multicast-pkts": 600,
		"in-broadcast-pkts":  700,
		"out-broadcast-pkts": 800,
		"in-crc-errors":      9,
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
