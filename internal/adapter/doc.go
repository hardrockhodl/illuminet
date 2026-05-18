// Package adapter is the parent package for telemetry adapters.
//
// An [Adapter] is a source of [model.Sample] data. Adapters run in
// their own goroutine, owned by the pipeline, and push samples onto a
// channel the pipeline supplies. The pipeline owns the channel's
// lifecycle; adapters must never close it.
//
// This package also provides [PollingAdapter], a helper that turns a
// fetch function into an Adapter that polls on a fixed interval. Most
// adapters that wrap a request/response API (NX-API, eAPI, NETCONF)
// can be built on top of PollingAdapter; streaming adapters (gNMI
// subscriptions) implement Adapter directly.
//
// Implemented subpackages:
//
//   - fake: deterministic in-process adapter used for end-to-end
//     pipeline tests and demos.
//
// Planned subpackages:
//
//   - nxos: Cisco Nexus 9000 family, primarily via gNMI streaming
//     telemetry with NX-API as a fallback.
//   - eos: Arista EOS, primarily via gNMI with eAPI as a fallback.
//   - junos: Juniper Junos, primarily via gNMI with NETCONF as a
//     fallback.
package adapter
