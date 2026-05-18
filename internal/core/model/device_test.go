package model

import "testing"

func TestDevice_ZeroValue(t *testing.T) {
	var d Device

	if d.Name != "" {
		t.Errorf("Name: got %q, want empty", d.Name)
	}
	if d.ManagementIP != "" {
		t.Errorf("ManagementIP: got %q, want empty", d.ManagementIP)
	}
	if d.Vendor != "" {
		t.Errorf("Vendor: got %q, want empty", d.Vendor)
	}
	if d.Model != "" {
		t.Errorf("Model: got %q, want empty", d.Model)
	}
	if d.OSVersion != "" {
		t.Errorf("OSVersion: got %q, want empty", d.OSVersion)
	}
	if d.KernelUptime != nil {
		t.Errorf("KernelUptime: got non-nil, want nil")
	}
	if d.Location != "" {
		t.Errorf("Location: got %q, want empty", d.Location)
	}
	if d.Role != DeviceRoleUnknown {
		t.Errorf("Role: got %q, want DeviceRoleUnknown", d.Role)
	}
	if d.CPUKernel != nil {
		t.Errorf("CPUKernel: got non-nil, want nil")
	}
	if d.CPUUser != nil {
		t.Errorf("CPUUser: got non-nil, want nil")
	}
	if d.MemoryTotal != nil {
		t.Errorf("MemoryTotal: got non-nil, want nil")
	}
	if d.MemoryUsed != nil {
		t.Errorf("MemoryUsed: got non-nil, want nil")
	}
	if d.ResponseTime != nil {
		t.Errorf("ResponseTime: got non-nil, want nil")
	}
}

func TestDeviceRole_String(t *testing.T) {
	tests := []struct {
		role DeviceRole
		want string
	}{
		{DeviceRoleUnknown, "unknown"},
		{DeviceRoleSpine, "spine"},
		{DeviceRoleLeaf, "leaf"},
		{DeviceRoleBorder, "border"},
		{DeviceRoleEdge, "edge"},
	}

	for _, tc := range tests {
		got := tc.role.String()
		if got != tc.want {
			t.Errorf("DeviceRole(%q).String() = %q, want %q", string(tc.role), got, tc.want)
		}
	}
}
