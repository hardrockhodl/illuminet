package model

import "time"

// BufferCounters holds shared-buffer occupancy and headroom telemetry
// for one buffer scope (per-interface or per-ASIC instance).
//
// Cells are the vendor-native unit on the platforms we initially
// target. The pipeline does not convert between cells and bytes; that
// conversion requires platform-specific knowledge of cell size and is
// deferred to enrichment.
type BufferCounters struct {
	// PeakCellDropPG is the peak number of buffer cells observed in
	// drop-eligible priority groups since counters were last cleared.
	PeakCellDropPG *uint64

	// PeakCellNoDrop is the peak number of buffer cells observed in
	// no-drop / lossless priority groups since counters were last
	// cleared.
	PeakCellNoDrop *uint64

	// CurrentCellDropPG is the instantaneous number of buffer cells
	// in drop-eligible priority groups at sample time.
	CurrentCellDropPG *uint64

	// CurrentCellNoDrop is the instantaneous number of buffer cells
	// in no-drop / lossless priority groups at sample time.
	CurrentCellNoDrop *uint64

	// LastClear is the time at which these buffer counters were last
	// cleared. Nil when the platform did not report it. The pipeline
	// MUST NOT compute deltas across a LastClear change; doing so
	// produces spurious negative deltas after a clear.
	LastClear *time.Time
}

// BufferInstance is a snapshot of one shared-buffer scope on a device,
// typically corresponding to one ASIC instance or forwarding engine.
// Platforms with a single monolithic buffer expose a single
// BufferInstance; multi-slice platforms expose one per slice.
type BufferInstance struct {
	// ID is the platform-native instance index. Numbering is vendor
	// specific.
	ID int

	// Name is the platform-native identifier (e.g. "module-1",
	// "slice-0"). Empty when the platform does not name its
	// instances.
	Name string

	// Counters holds the buffer occupancy counters for this instance.
	Counters BufferCounters
}
