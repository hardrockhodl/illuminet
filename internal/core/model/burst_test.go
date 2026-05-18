package model

import "testing"

func TestBurstEvent_ZeroValue(t *testing.T) {
	var b BurstEvent

	if b.Interface != "" {
		t.Errorf("Interface: got %q, want empty", b.Interface)
	}
	if b.QueueID != 0 {
		t.Errorf("QueueID: got %d, want 0", b.QueueID)
	}
	if b.StartDepth != nil {
		t.Errorf("StartDepth: got non-nil, want nil")
	}
	if b.EndDepth != nil {
		t.Errorf("EndDepth: got non-nil, want nil")
	}
	if b.PeakDepth != nil {
		t.Errorf("PeakDepth: got non-nil, want nil")
	}
	if b.Duration != 0 {
		t.Errorf("Duration: got %v, want 0", b.Duration)
	}
	if !b.PeakTime.IsZero() {
		t.Errorf("PeakTime: got %v, want zero value", b.PeakTime)
	}
}
