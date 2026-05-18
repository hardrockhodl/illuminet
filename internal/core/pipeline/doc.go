// Package pipeline orchestrates the IllumiNET collection loop.
//
// Responsibilities of this package, once implemented, will include:
//
//   - Driving the lifecycle of vendor adapters (start, stop, reconnect,
//     backoff) and consuming the streams of normalized samples they
//     produce.
//   - Correlation of per-interface, per-queue, and per-buffer counters
//     across telemetry sources and time windows.
//   - Enrichment of samples with topology context derived from LLDP, and
//     classification of neighbors (host, switch, other) including the
//     Core vs. Edge port split that paregupt's nexus_traffic_monitor
//     established as a useful operational view.
//   - Dispatching enriched samples to the configured exporters in
//     internal/exporter.
//
// No collection logic is implemented in this initial scaffold.
package pipeline
