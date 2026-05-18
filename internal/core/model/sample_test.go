package model

import (
	"testing"
	"time"
)

func TestSample_ZeroValue(t *testing.T) {
	var s Sample

	if s.Device.Name != "" {
		t.Errorf("Device.Name: got %q, want empty", s.Device.Name)
	}
	if s.Interfaces != nil {
		t.Errorf("Interfaces: got non-nil, want nil")
	}
	if s.Buffers != nil {
		t.Errorf("Buffers: got non-nil, want nil")
	}
	if s.Bursts != nil {
		t.Errorf("Bursts: got non-nil, want nil")
	}
	if !s.CollectedAt.IsZero() {
		t.Errorf("CollectedAt: got %v, want zero value", s.CollectedAt)
	}
	if s.Latency != 0 {
		t.Errorf("Latency: got %v, want 0", s.Latency)
	}
}

func TestSample_MinimalRoundtrip(t *testing.T) {
	rxBytes := uint64(1024)
	txBytes := uint64(2048)
	speed := uint64(100_000_000_000)

	collectedAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	observedAt := collectedAt.Add(50 * time.Millisecond)

	s := Sample{
		Device: Device{
			Name:         "leaf-01",
			ManagementIP: "10.0.0.1",
			Vendor:       "cisco",
			Model:        "N9K-C9332D-GX2B",
			OSVersion:    "10.5(1)",
			Role:         DeviceRoleLeaf,
			Location:     "rack-A12",
		},
		Interfaces: []Interface{
			{
				Name:       "Ethernet1/1",
				AdminState: AdminStateUp,
				OperState:  OperStateUp,
				OperMode:   OperModeRouted,
				OperSpeed:  &speed,
				Counters: &InterfaceCounters{
					RxBytes: &rxBytes,
					TxBytes: &txBytes,
				},
				ObservedAt: observedAt,
			},
		},
		CollectedAt: collectedAt,
		Latency:     75 * time.Millisecond,
	}

	if s.Device.Name != "leaf-01" {
		t.Errorf("Device.Name: got %q, want %q", s.Device.Name, "leaf-01")
	}
	if s.Device.Role != DeviceRoleLeaf {
		t.Errorf("Device.Role: got %q, want DeviceRoleLeaf", s.Device.Role)
	}
	if got := s.Device.Role.String(); got != "leaf" {
		t.Errorf("Device.Role.String(): got %q, want %q", got, "leaf")
	}

	if len(s.Interfaces) != 1 {
		t.Fatalf("Interfaces: got %d entries, want 1", len(s.Interfaces))
	}
	iface := s.Interfaces[0]
	if iface.Name != "Ethernet1/1" {
		t.Errorf("Interfaces[0].Name: got %q, want %q", iface.Name, "Ethernet1/1")
	}
	if iface.AdminState != AdminStateUp {
		t.Errorf("Interfaces[0].AdminState: got %q, want AdminStateUp", iface.AdminState)
	}
	if iface.OperSpeed == nil || *iface.OperSpeed != speed {
		t.Errorf("Interfaces[0].OperSpeed: got %v, want %d", iface.OperSpeed, speed)
	}
	if iface.Counters == nil {
		t.Fatalf("Interfaces[0].Counters: got nil, want populated")
	}
	if iface.Counters.RxBytes == nil || *iface.Counters.RxBytes != rxBytes {
		t.Errorf("Interfaces[0].Counters.RxBytes: got %v, want %d", iface.Counters.RxBytes, rxBytes)
	}
	if iface.Counters.TxBytes == nil || *iface.Counters.TxBytes != txBytes {
		t.Errorf("Interfaces[0].Counters.TxBytes: got %v, want %d", iface.Counters.TxBytes, txBytes)
	}
	if iface.Counters.RxCRC != nil {
		t.Errorf("Interfaces[0].Counters.RxCRC: got non-nil, want nil (unreported)")
	}
	if !iface.ObservedAt.Equal(observedAt) {
		t.Errorf("Interfaces[0].ObservedAt: got %v, want %v", iface.ObservedAt, observedAt)
	}

	if !s.CollectedAt.Equal(collectedAt) {
		t.Errorf("CollectedAt: got %v, want %v", s.CollectedAt, collectedAt)
	}
	if s.Latency != 75*time.Millisecond {
		t.Errorf("Latency: got %v, want %v", s.Latency, 75*time.Millisecond)
	}
}
