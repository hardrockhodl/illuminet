// Package model defines the normalized, vendor-agnostic domain model
// used throughout IllumiNET.
//
// The eventual surface of this package is expected to include the
// following types:
//
//   - Device: a single switch or router, identified by a stable
//     management address and platform metadata (vendor, OS family, OS
//     version, serial number).
//   - Interface: a physical or logical port on a Device, with operational
//     state, speed, MTU, and link-layer counters.
//   - Queue: a per-interface egress or ingress queue, holding QoS
//     metadata and drop counters.
//   - BufferInstance: a shared or dedicated buffer pool snapshot, sized
//     in cells or bytes depending on platform.
//   - PFCCounter: per-priority-class Priority Flow Control pause counters,
//     captured as a pair (RX pauses received, TX pauses generated) so
//     that the asymmetric ingress vs. egress signal can be analyzed.
//   - ECNCounter: per-queue Explicit Congestion Notification marking
//     counters.
//   - BurstEvent: a peak-time annotated record of a transient utilization
//     spike, including the timestamp at which the peak was observed.
//
// All values in this package are intended to be vendor agnostic. Vendor
// adapters in internal/adapter are responsible for mapping platform
// specific telemetry into these types.
//
// No types are implemented in this initial scaffold.
package model
