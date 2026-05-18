package fake

import (
	"context"
	"testing"
	"time"
)

func TestFake_CountersMonotonic(t *testing.T) {
	f := newFake(time.Second, nil)

	s1, err := f.fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch 1: %v", err)
	}
	s2, err := f.fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch 2: %v", err)
	}

	pairs := []struct {
		name string
		a, b uint64
	}{
		{"Eth1/1 RxBytes", *s1.Interfaces[0].Counters.RxBytes, *s2.Interfaces[0].Counters.RxBytes},
		{"Eth1/1 TxBytes", *s1.Interfaces[0].Counters.TxBytes, *s2.Interfaces[0].Counters.TxBytes},
		{"Eth1/49 RxBytes", *s1.Interfaces[1].Counters.RxBytes, *s2.Interfaces[1].Counters.RxBytes},
		{"Eth1/49 TxBytes", *s1.Interfaces[1].Counters.TxBytes, *s2.Interfaces[1].Counters.TxBytes},
	}
	for _, p := range pairs {
		if p.b <= p.a {
			t.Errorf("%s did not increase: %d -> %d", p.name, p.a, p.b)
		}
	}
}

func TestFake_Deterministic(t *testing.T) {
	f1 := newFake(time.Second, nil)
	f2 := newFake(time.Second, nil)

	for i := 0; i < 3; i++ {
		s1, err := f1.fetch(context.Background())
		if err != nil {
			t.Fatalf("f1 fetch %d: %v", i, err)
		}
		s2, err := f2.fetch(context.Background())
		if err != nil {
			t.Fatalf("f2 fetch %d: %v", i, err)
		}
		if *s1.Interfaces[0].Counters.RxBytes != *s2.Interfaces[0].Counters.RxBytes {
			t.Errorf("tick %d: Eth1/1 RxBytes diverged: %d vs %d",
				i, *s1.Interfaces[0].Counters.RxBytes, *s2.Interfaces[0].Counters.RxBytes)
		}
		if *s1.Interfaces[1].Counters.TxBytes != *s2.Interfaces[1].Counters.TxBytes {
			t.Errorf("tick %d: Eth1/49 TxBytes diverged: %d vs %d",
				i, *s1.Interfaces[1].Counters.TxBytes, *s2.Interfaces[1].Counters.TxBytes)
		}
	}
}

func TestFake_SampleStructure(t *testing.T) {
	f := newFake(time.Second, nil)
	s, err := f.fetch(context.Background())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	if s.Device.Name != "fake-switch" {
		t.Errorf("Device.Name: got %q, want fake-switch", s.Device.Name)
	}
	if s.Device.Vendor != "fake" {
		t.Errorf("Device.Vendor: got %q, want fake", s.Device.Vendor)
	}
	if s.Device.CPUKernel == nil {
		t.Error("Device.CPUKernel: got nil")
	}

	if len(s.Interfaces) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(s.Interfaces))
	}
	wantNames := []string{"Ethernet1/1", "Ethernet1/49"}
	for i, want := range wantNames {
		if s.Interfaces[i].Name != want {
			t.Errorf("Interfaces[%d].Name: got %q, want %q", i, s.Interfaces[i].Name, want)
		}
		if len(s.Interfaces[i].Queues) != 1 {
			t.Errorf("Interfaces[%d] queues: got %d, want 1", i, len(s.Interfaces[i].Queues))
			continue
		}
		if s.Interfaces[i].Queues[0].PFC == nil {
			t.Errorf("Interfaces[%d] queue 0 missing PFC", i)
		}
	}

	if len(s.Buffers) != 1 {
		t.Errorf("expected 1 buffer instance, got %d", len(s.Buffers))
	}
}

func TestFake_CPUOscillates(t *testing.T) {
	f := newFake(5*time.Second, nil)

	// Collect 12 fetches; with interval=5s the synthetic clock covers
	// 60s, exactly one full period of the CPU sinusoid.
	var minCPU, maxCPU float64 = 100, 0
	for i := 0; i < 12; i++ {
		s, err := f.fetch(context.Background())
		if err != nil {
			t.Fatalf("fetch %d: %v", i, err)
		}
		v := *s.Device.CPUKernel
		if v < minCPU {
			minCPU = v
		}
		if v > maxCPU {
			maxCPU = v
		}
	}
	// Sinusoid centered at 30 with amplitude 20 should hit at least
	// 45+ on the high side and 15- on the low side within a period.
	if maxCPU < 45 {
		t.Errorf("CPU max %.2f too low; expected sinusoid to peak above 45", maxCPU)
	}
	if minCPU > 15 {
		t.Errorf("CPU min %.2f too high; expected sinusoid to trough below 15", minCPU)
	}
}
