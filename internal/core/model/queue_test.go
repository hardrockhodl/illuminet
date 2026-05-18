package model

import "testing"

func TestQueue_ZeroValue(t *testing.T) {
	var q Queue

	if q.ID != 0 {
		t.Errorf("ID: got %d, want 0", q.ID)
	}
	if q.Name != "" {
		t.Errorf("Name: got %q, want empty", q.Name)
	}
	if q.PFC != nil {
		t.Errorf("PFC: got non-nil, want nil")
	}
	if q.ECN != nil {
		t.Errorf("ECN: got non-nil, want nil")
	}

	// Counters is stored by value; verify its own pointer fields are
	// nil at the zero value.
	if q.Counters.TxBytes != nil ||
		q.Counters.TxPkts != nil ||
		q.Counters.DropBytes != nil ||
		q.Counters.DropPkts != nil ||
		q.Counters.RandomDropBytes != nil ||
		q.Counters.RandomDropPkts != nil ||
		q.Counters.CurrentDepth != nil ||
		q.Counters.PeakDepth != nil {
		t.Errorf("Counters: expected all pointer fields nil at zero value, got %+v", q.Counters)
	}
}

func TestQueueCounters_ZeroValue(t *testing.T) {
	var c QueueCounters

	if c.TxBytes != nil {
		t.Errorf("TxBytes: got non-nil, want nil")
	}
	if c.TxPkts != nil {
		t.Errorf("TxPkts: got non-nil, want nil")
	}
	if c.DropBytes != nil {
		t.Errorf("DropBytes: got non-nil, want nil")
	}
	if c.DropPkts != nil {
		t.Errorf("DropPkts: got non-nil, want nil")
	}
	if c.RandomDropBytes != nil {
		t.Errorf("RandomDropBytes: got non-nil, want nil")
	}
	if c.RandomDropPkts != nil {
		t.Errorf("RandomDropPkts: got non-nil, want nil")
	}
	if c.CurrentDepth != nil {
		t.Errorf("CurrentDepth: got non-nil, want nil")
	}
	if c.PeakDepth != nil {
		t.Errorf("PeakDepth: got non-nil, want nil")
	}
}
