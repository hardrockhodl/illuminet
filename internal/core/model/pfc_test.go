package model

import "testing"

func TestPFCCounter_ZeroValue(t *testing.T) {
	var c PFCCounter

	if c.RxPause != nil {
		t.Errorf("RxPause: got non-nil, want nil")
	}
	if c.TxPause != nil {
		t.Errorf("TxPause: got non-nil, want nil")
	}
	if c.WatchdogEvents != nil {
		t.Errorf("WatchdogEvents: got non-nil, want nil")
	}
}
