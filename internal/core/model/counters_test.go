package model

import "testing"

func TestInterfaceCounters_ZeroValue(t *testing.T) {
	var c InterfaceCounters

	if c.RxBytes != nil {
		t.Errorf("RxBytes: got non-nil, want nil")
	}
	if c.TxBytes != nil {
		t.Errorf("TxBytes: got non-nil, want nil")
	}
	if c.RxUcastPkts != nil {
		t.Errorf("RxUcastPkts: got non-nil, want nil")
	}
	if c.TxUcastPkts != nil {
		t.Errorf("TxUcastPkts: got non-nil, want nil")
	}
	if c.RxMcastPkts != nil {
		t.Errorf("RxMcastPkts: got non-nil, want nil")
	}
	if c.TxMcastPkts != nil {
		t.Errorf("TxMcastPkts: got non-nil, want nil")
	}
	if c.RxBcastPkts != nil {
		t.Errorf("RxBcastPkts: got non-nil, want nil")
	}
	if c.TxBcastPkts != nil {
		t.Errorf("TxBcastPkts: got non-nil, want nil")
	}
	if c.RxCRC != nil {
		t.Errorf("RxCRC: got non-nil, want nil")
	}
	if c.RxCRCStomped != nil {
		t.Errorf("RxCRCStomped: got non-nil, want nil")
	}
	if c.RxJumbo != nil {
		t.Errorf("RxJumbo: got non-nil, want nil")
	}
	if c.TxJumbo != nil {
		t.Errorf("TxJumbo: got non-nil, want nil")
	}
	if c.PacketSizeRx != nil {
		t.Errorf("PacketSizeRx: got non-nil, want nil")
	}
	if c.PacketSizeTx != nil {
		t.Errorf("PacketSizeTx: got non-nil, want nil")
	}
	if c.LastClear != nil {
		t.Errorf("LastClear: got non-nil, want nil")
	}
}

func TestPacketSizeBuckets_ZeroValue(t *testing.T) {
	var b PacketSizeBuckets

	if b.Pkts64 != nil {
		t.Errorf("Pkts64: got non-nil, want nil")
	}
	if b.Pkts65to127 != nil {
		t.Errorf("Pkts65to127: got non-nil, want nil")
	}
	if b.Pkts128to255 != nil {
		t.Errorf("Pkts128to255: got non-nil, want nil")
	}
	if b.Pkts256to511 != nil {
		t.Errorf("Pkts256to511: got non-nil, want nil")
	}
	if b.Pkts512to1023 != nil {
		t.Errorf("Pkts512to1023: got non-nil, want nil")
	}
	if b.Pkts1024to1518 != nil {
		t.Errorf("Pkts1024to1518: got non-nil, want nil")
	}
	if b.Pkts1519to2047 != nil {
		t.Errorf("Pkts1519to2047: got non-nil, want nil")
	}
	if b.Pkts2048to4095 != nil {
		t.Errorf("Pkts2048to4095: got non-nil, want nil")
	}
	if b.Pkts4096to9216 != nil {
		t.Errorf("Pkts4096to9216: got non-nil, want nil")
	}
}
