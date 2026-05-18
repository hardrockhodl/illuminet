package model

import "testing"

func TestInterface_ZeroValue(t *testing.T) {
	var i Interface

	if i.Name != "" {
		t.Errorf("Name: got %q, want empty", i.Name)
	}
	if i.Description != "" {
		t.Errorf("Description: got %q, want empty", i.Description)
	}
	if i.AdminState != AdminStateUnknown {
		t.Errorf("AdminState: got %q, want AdminStateUnknown", i.AdminState)
	}
	if i.OperState != OperStateUnknown {
		t.Errorf("OperState: got %q, want OperStateUnknown", i.OperState)
	}
	if i.OperMode != OperModeUnknown {
		t.Errorf("OperMode: got %q, want OperModeUnknown", i.OperMode)
	}
	if i.OperSpeed != nil {
		t.Errorf("OperSpeed: got non-nil, want nil")
	}
	if i.DownReason != "" {
		t.Errorf("DownReason: got %q, want empty", i.DownReason)
	}
	if i.Classification != PortClassificationUnknown {
		t.Errorf("Classification: got %q, want PortClassificationUnknown", i.Classification)
	}
	if i.Peer != nil {
		t.Errorf("Peer: got non-nil, want nil")
	}
	if i.Counters != nil {
		t.Errorf("Counters: got non-nil, want nil")
	}
	if i.Queues != nil {
		t.Errorf("Queues: got non-nil, want nil")
	}
	if i.Buffer != nil {
		t.Errorf("Buffer: got non-nil, want nil")
	}
	if !i.ObservedAt.IsZero() {
		t.Errorf("ObservedAt: got %v, want zero value", i.ObservedAt)
	}
}

func TestAdminState_String(t *testing.T) {
	tests := []struct {
		state AdminState
		want  string
	}{
		{AdminStateUnknown, "unknown"},
		{AdminStateUp, "up"},
		{AdminStateDown, "down"},
	}

	for _, tc := range tests {
		got := tc.state.String()
		if got != tc.want {
			t.Errorf("AdminState(%q).String() = %q, want %q", string(tc.state), got, tc.want)
		}
	}
}

func TestOperState_String(t *testing.T) {
	tests := []struct {
		state OperState
		want  string
	}{
		{OperStateUnknown, "unknown"},
		{OperStateUp, "up"},
		{OperStateDown, "down"},
		{OperStateTesting, "testing"},
	}

	for _, tc := range tests {
		got := tc.state.String()
		if got != tc.want {
			t.Errorf("OperState(%q).String() = %q, want %q", string(tc.state), got, tc.want)
		}
	}
}

func TestOperMode_String(t *testing.T) {
	tests := []struct {
		mode OperMode
		want string
	}{
		{OperModeUnknown, "unknown"},
		{OperModeRouted, "routed"},
		{OperModeAccess, "access"},
		{OperModeTrunk, "trunk"},
		{OperModeFEXFabric, "fex-fabric"},
		{OperModeOther, "other"},
	}

	for _, tc := range tests {
		got := tc.mode.String()
		if got != tc.want {
			t.Errorf("OperMode(%q).String() = %q, want %q", string(tc.mode), got, tc.want)
		}
	}
}

func TestPortClassification_String(t *testing.T) {
	tests := []struct {
		class PortClassification
		want  string
	}{
		{PortClassificationUnknown, "unknown"},
		{PortClassificationCore, "core"},
		{PortClassificationEdge, "edge"},
		{PortClassificationManagement, "management"},
		{PortClassificationOther, "other"},
	}

	for _, tc := range tests {
		got := tc.class.String()
		if got != tc.want {
			t.Errorf("PortClassification(%q).String() = %q, want %q", string(tc.class), got, tc.want)
		}
	}
}
