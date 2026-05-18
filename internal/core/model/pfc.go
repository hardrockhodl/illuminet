package model

// PFCCounter holds Priority Flow Control pause counters for one queue
// or priority class. The RX/TX asymmetry is preserved deliberately:
// received pauses indicate the local egress was throttled by a
// congested downstream peer, while transmitted pauses indicate local
// ingress congestion that we are pushing back on the upstream peer.
type PFCCounter struct {
	// RxPause is the cumulative count of PFC pause frames received on
	// the associated queue / priority. Indicates downstream congestion
	// at the peer's ingress.
	RxPause *uint64

	// TxPause is the cumulative count of PFC pause frames transmitted
	// on the associated queue / priority. Indicates local ingress
	// congestion being signaled back to the upstream peer.
	TxPause *uint64

	// WatchdogEvents is the cumulative count of PFC watchdog
	// activations on the associated queue. Nil when the platform does
	// not implement a PFC watchdog or did not report it.
	WatchdogEvents *uint64
}
