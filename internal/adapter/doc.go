// Package adapter is the parent package for vendor-specific telemetry
// adapters.
//
// Each supported platform lives in a subpackage and is responsible for
// translating native telemetry into the vendor-agnostic types defined
// in internal/core/model. Initial planned subpackages:
//
//   - nxos: Cisco Nexus 9000 family, primarily via gNMI streaming
//     telemetry with NX-API as a fallback for state that is not yet
//     exposed via gNMI on the target NX-OS train.
//   - eos: Arista EOS, primarily via gNMI with eAPI as a fallback.
//   - junos: Juniper Junos, primarily via gNMI with NETCONF as a
//     fallback.
//
// No adapter is implemented in this initial scaffold.
package adapter
