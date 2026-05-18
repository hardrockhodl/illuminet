package model

import "testing"

func TestPeer_ZeroValue(t *testing.T) {
	var p Peer

	if p.Name != "" {
		t.Errorf("Name: got %q, want empty", p.Name)
	}
	if p.Interface != "" {
		t.Errorf("Interface: got %q, want empty", p.Interface)
	}
	if p.MgmtIP != "" {
		t.Errorf("MgmtIP: got %q, want empty", p.MgmtIP)
	}
	if p.Type != PeerTypeUnknown {
		t.Errorf("Type: got %q, want PeerTypeUnknown", p.Type)
	}
	if p.LearnedVia != "" {
		t.Errorf("LearnedVia: got %q, want empty", p.LearnedVia)
	}
	if p.Capabilities != nil {
		t.Errorf("Capabilities: got %v, want nil", p.Capabilities)
	}
	if p.SystemDescription != "" {
		t.Errorf("SystemDescription: got %q, want empty", p.SystemDescription)
	}
}

func TestPeerType_String(t *testing.T) {
	tests := []struct {
		typ  PeerType
		want string
	}{
		{PeerTypeUnknown, "unknown"},
		{PeerTypeHost, "host"},
		{PeerTypeSwitch, "switch"},
		{PeerTypeOther, "other"},
	}

	for _, tc := range tests {
		got := tc.typ.String()
		if got != tc.want {
			t.Errorf("PeerType(%q).String() = %q, want %q", string(tc.typ), got, tc.want)
		}
	}
}
