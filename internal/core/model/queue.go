package model

// QueueCounters holds the traffic-management counters for a single
// queue. Drop counters are split between tail drops (DropBytes /
// DropPkts) and active queue management drops (RandomDropBytes /
// RandomDropPkts, typically WRED) so the cause of loss can be
// distinguished downstream.
type QueueCounters struct {
	// TxBytes is the cumulative bytes transmitted out of the queue.
	TxBytes *uint64
	// TxPkts is the cumulative packets transmitted out of the queue.
	TxPkts *uint64

	// DropBytes is the cumulative bytes dropped from the queue,
	// excluding random / WRED drops.
	DropBytes *uint64
	// DropPkts is the cumulative packets dropped from the queue,
	// excluding random / WRED drops.
	DropPkts *uint64

	// RandomDropBytes is the cumulative bytes dropped by the queue's
	// active queue management algorithm (WRED / AFD).
	RandomDropBytes *uint64
	// RandomDropPkts is the cumulative packets dropped by the queue's
	// active queue management algorithm (WRED / AFD).
	RandomDropPkts *uint64

	// CurrentDepth is the instantaneous queue depth at sample time.
	// The unit (bytes vs. cells vs. packets) is vendor specific; the
	// adapter records the value the platform exposes natively and the
	// pipeline treats the unit per-vendor. Nil when not reported.
	CurrentDepth *uint64

	// PeakDepth is the highest depth observed since counters were last
	// cleared, in the same unit as CurrentDepth. Nil when not
	// reported.
	PeakDepth *uint64
}

// Queue is one transmit (or, on a few platforms, receive) queue on an
// Interface. A Queue is recorded only when at least some data is
// available for it; an Interface with no measured queues carries an
// empty slice rather than zero-value Queue entries.
type Queue struct {
	// ID is the platform-native queue index. Numbering is vendor
	// specific and is not normalized.
	ID int

	// Name is the platform-native queue or class-map identifier. Empty
	// when the platform does not name its queues.
	Name string

	// Counters holds the traffic-management counters for this queue.
	// Stored by value rather than pointer because a Queue is only
	// instantiated when there is data to populate.
	Counters QueueCounters

	// PFC holds the PFC pause counters for this queue. Nil when PFC
	// is not configured for the queue or the platform did not report
	// PFC state.
	PFC *PFCCounter

	// ECN holds the ECN marking counters for this queue. Nil when the
	// platform did not report ECN state.
	ECN *ECNCounter
}
