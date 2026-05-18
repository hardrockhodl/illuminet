package model

import "time"

// BurstEvent is a peak-time annotated record of a transient queue
// utilization spike. Burst detection happens inside the adapter or
// inside enrichment; this type is the carrier across the pipeline.
//
// PeakTime is the timestamp at which the burst itself occurred, not
// the timestamp of the collection tick that surfaced it. The two can
// differ by seconds when bursts are reconstructed from on-device
// micro-burst history. See docs/ARCHITECTURE.md.
type BurstEvent struct {
	// Interface is the platform-native interface name on which the
	// burst occurred.
	Interface string

	// QueueID is the platform-native queue index on which the burst
	// occurred.
	QueueID int

	// StartDepth is the queue depth observed at the start of the
	// burst, in the same unit as Queue.Counters.CurrentDepth. Nil
	// when not reported.
	StartDepth *uint64

	// EndDepth is the queue depth observed at the end of the burst,
	// in the same unit as StartDepth. Nil when not reported.
	EndDepth *uint64

	// PeakDepth is the maximum queue depth observed during the burst,
	// in the same unit as StartDepth. Nil when not reported.
	PeakDepth *uint64

	// Duration is the duration of the burst. The zero value indicates
	// the burst was instantaneous within the platform's measurement
	// resolution.
	Duration time.Duration

	// PeakTime is the timestamp at which the peak depth was observed.
	// This is the time of the burst, not the time of the collection
	// tick that reported it.
	PeakTime time.Time
}
