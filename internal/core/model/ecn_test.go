package model

import "testing"

func TestECNCounter_ZeroValue(t *testing.T) {
	var c ECNCounter

	if c.MarkedPkts != nil {
		t.Errorf("MarkedPkts: got non-nil, want nil")
	}
}
