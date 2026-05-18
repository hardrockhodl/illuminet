package model

import "testing"

func TestBufferCounters_ZeroValue(t *testing.T) {
	var c BufferCounters

	if c.PeakCellDropPG != nil {
		t.Errorf("PeakCellDropPG: got non-nil, want nil")
	}
	if c.PeakCellNoDrop != nil {
		t.Errorf("PeakCellNoDrop: got non-nil, want nil")
	}
	if c.CurrentCellDropPG != nil {
		t.Errorf("CurrentCellDropPG: got non-nil, want nil")
	}
	if c.CurrentCellNoDrop != nil {
		t.Errorf("CurrentCellNoDrop: got non-nil, want nil")
	}
	if c.LastClear != nil {
		t.Errorf("LastClear: got non-nil, want nil")
	}
}

func TestBufferInstance_ZeroValue(t *testing.T) {
	var b BufferInstance

	if b.ID != 0 {
		t.Errorf("ID: got %d, want 0", b.ID)
	}
	if b.Name != "" {
		t.Errorf("Name: got %q, want empty", b.Name)
	}
	if b.Counters.PeakCellDropPG != nil ||
		b.Counters.PeakCellNoDrop != nil ||
		b.Counters.CurrentCellDropPG != nil ||
		b.Counters.CurrentCellNoDrop != nil ||
		b.Counters.LastClear != nil {
		t.Errorf("Counters: expected all pointer fields nil at zero value, got %+v", b.Counters)
	}
}
