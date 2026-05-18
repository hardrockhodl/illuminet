// Package nxos is the Cisco NX-OS adapter.
//
// It connects to a single NX-OS switch over gNMI and translates
// streaming telemetry into [model.Sample] values. The adapter targets
// Cisco NX-OS 10.4+ where the gNMI server supports the OpenConfig
// interface and system models that this iteration relies on:
//
//   - openconfig-interfaces
//   - openconfig-system
//
// Coverage in the current iteration is limited to interface counters
// (in/out octets, unicast/multicast/broadcast packet counts, CRC
// errors), interface state (admin-status, oper-status, description)
// and basic device metadata (hostname, software version, location
// from operator-supplied configuration).
//
// Queue, PFC, ECN, buffer occupancy, burst events and LLDP-derived
// peer discovery are deliberately left out of this iteration. Those
// fields are populated by adapters layered on top of Cisco-native
// gNMI paths (Cisco-IOS-XR-style Cisco-NX-OS-device YANG modules) and
// arrive in follow-up iterations.
//
// The adapter uses github.com/openconfig/gnmic/pkg/api as its gNMI
// client. The pure translation function [Translate] is exported so it
// can be exercised in isolation without a live target.
package nxos
