package model

import "time"

// PacketSizeBuckets holds per-frame-size packet counts in the buckets
// commonly exposed by switching ASICs. All buckets are optional; a nil
// pointer means the platform did not report that bucket.
type PacketSizeBuckets struct {
	// Pkts64 counts frames of exactly 64 bytes.
	Pkts64 *uint64
	// Pkts65to127 counts frames in the 65-127 byte range.
	Pkts65to127 *uint64
	// Pkts128to255 counts frames in the 128-255 byte range.
	Pkts128to255 *uint64
	// Pkts256to511 counts frames in the 256-511 byte range.
	Pkts256to511 *uint64
	// Pkts512to1023 counts frames in the 512-1023 byte range.
	Pkts512to1023 *uint64
	// Pkts1024to1518 counts frames in the 1024-1518 byte range.
	Pkts1024to1518 *uint64
	// Pkts1519to2047 counts jumbo-range frames in the 1519-2047 byte
	// range. Not exposed by all platforms.
	Pkts1519to2047 *uint64
	// Pkts2048to4095 counts jumbo-range frames in the 2048-4095 byte
	// range. Not exposed by all platforms.
	Pkts2048to4095 *uint64
	// Pkts4096to9216 counts jumbo-range frames in the 4096-9216 byte
	// range. Not exposed by all platforms.
	Pkts4096to9216 *uint64
}

// InterfaceCounters is the link-layer counter snapshot for an
// Interface. Every counter is optional; nil means the platform did
// not report it. A reported zero is distinct from nil.
type InterfaceCounters struct {
	// RxBytes is the cumulative received byte count.
	RxBytes *uint64
	// TxBytes is the cumulative transmitted byte count.
	TxBytes *uint64

	// RxUcastPkts is the cumulative received unicast packet count.
	RxUcastPkts *uint64
	// TxUcastPkts is the cumulative transmitted unicast packet count.
	TxUcastPkts *uint64
	// RxMcastPkts is the cumulative received multicast packet count.
	RxMcastPkts *uint64
	// TxMcastPkts is the cumulative transmitted multicast packet count.
	TxMcastPkts *uint64
	// RxBcastPkts is the cumulative received broadcast packet count.
	RxBcastPkts *uint64
	// TxBcastPkts is the cumulative transmitted broadcast packet count.
	TxBcastPkts *uint64

	// RxCRC is the cumulative count of frames received with a CRC
	// error.
	RxCRC *uint64
	// RxCRCStomped is the cumulative count of frames received with a
	// stomped CRC, i.e. a CRC that was deliberately invalidated by an
	// upstream cut-through forwarder when it discovered corruption
	// mid-frame. Cisco NX-OS specific today; left nil by other
	// vendors.
	RxCRCStomped *uint64

	// RxJumbo is the cumulative count of received jumbo frames.
	RxJumbo *uint64
	// TxJumbo is the cumulative count of transmitted jumbo frames.
	TxJumbo *uint64

	// PacketSizeRx is the per-size receive histogram. Nil when the
	// platform did not report a size histogram.
	PacketSizeRx *PacketSizeBuckets
	// PacketSizeTx is the per-size transmit histogram. Nil when the
	// platform did not report a size histogram.
	PacketSizeTx *PacketSizeBuckets

	// LastClear is the time at which counters on this interface were
	// last cleared. Nil when the platform did not report it. The
	// pipeline MUST NOT compute deltas across a LastClear change.
	LastClear *time.Time
}
