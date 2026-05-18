package model

// ECNCounter holds Explicit Congestion Notification marking counts for
// one queue. A high mark count under sustained load is the canonical
// signal that the queue is approaching its WRED / AFD thresholds.
type ECNCounter struct {
	// MarkedPkts is the cumulative count of packets that were marked
	// with ECN-CE while traversing the associated queue. Nil when the
	// platform did not report it.
	MarkedPkts *uint64
}
