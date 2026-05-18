package model

import "time"

// Sample is the top-level container for one collection tick against
// one Device. It is the unit produced by the collector and consumed by
// every downstream stage: pipeline, enrichment, exporters.
//
// A Sample is immutable. Each tick produces a new Sample; consumers
// that need to compare against a previous tick keep their own
// references rather than mutating in place.
type Sample struct {
	// Device holds the device-level metadata and gauges for this tick.
	Device Device

	// Interfaces holds the per-interface snapshots taken during this
	// tick. May be empty if no interface data was collected.
	Interfaces []Interface

	// Buffers holds the per-ASIC-instance buffer snapshots for this
	// tick. May be empty on platforms that expose buffer state only
	// at the per-interface level (see Interface.Buffer).
	Buffers []BufferInstance

	// Bursts holds the burst events observed during this tick. May be
	// empty when no bursts were detected or the platform does not
	// surface burst telemetry.
	Bursts []BurstEvent

	// CollectedAt is the timestamp at which this collection tick
	// started, taken from the collector's clock. A zero time.Time
	// means the start time was not recorded.
	CollectedAt time.Time

	// Latency is the wall-clock duration the collector spent
	// producing this Sample, from CollectedAt to the moment the
	// Sample was assembled.
	Latency time.Duration
}
