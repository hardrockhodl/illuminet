// Package fake provides a deterministic in-process Adapter used for
// end-to-end pipeline tests and demos.
//
// The adapter emits a synthetic two-interface switch with monotonic,
// jitter-bounded counters and a sinusoidal CPU load. Counter
// trajectories are reproducible across runs: two fake adapters
// constructed at the same logical moment produce identical streams.
// This lets tests assert exact behavior and lets operators sanity-
// check downstream dashboards without a real device.
package fake
