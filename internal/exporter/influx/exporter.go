package influx

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// Exporter writes Sample data as InfluxDB Line Protocol records to a
// caller-supplied io.Writer.
//
// One Sample produces:
//
//   - At most one device record.
//   - One interface record per Interface in Sample.Interfaces.
//   - One queue record per Queue per Interface.
//   - One buffer record per BufferInstance in Sample.Buffers.
//   - One burst record per BurstEvent in Sample.Bursts.
//
// A record is silently dropped when it has no fields to emit; this
// happens, for example, for a zero-value Device with no counters.
//
// The Exporter is not safe for concurrent use.
type Exporter struct {
	w io.Writer
}

// New returns an Exporter that writes records to w. The caller retains
// ownership of w; the Exporter does not close it.
func New(w io.Writer) *Exporter {
	return &Exporter{w: w}
}

// Export emits Line Protocol records for the given sample. A nil
// sample is a no-op. A non-nil context error short-circuits the
// export with the context's error.
func (e *Exporter) Export(ctx context.Context, s *model.Sample) error {
	if s == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	enc := NewEncoder(e.w)

	if err := writeDevice(enc, s); err != nil {
		return err
	}

	for i := range s.Interfaces {
		iface := &s.Interfaces[i]
		ts := iface.ObservedAt
		if ts.IsZero() {
			ts = s.CollectedAt
		}
		if err := writeInterface(enc, s.Device, iface, ts); err != nil {
			return err
		}
		for j := range iface.Queues {
			if err := writeQueue(enc, s.Device, iface, &iface.Queues[j], ts); err != nil {
				return err
			}
		}
	}

	for i := range s.Buffers {
		if err := writeBufferInstance(enc, s.Device, &s.Buffers[i], s.CollectedAt); err != nil {
			return err
		}
	}

	for i := range s.Bursts {
		if err := writeBurst(enc, s.Device, &s.Bursts[i]); err != nil {
			return err
		}
	}

	return nil
}

// Close releases resources. It is currently a no-op because the
// underlying writer is owned by the caller.
func (e *Exporter) Close() error {
	return nil
}

func writeDevice(enc *Encoder, s *model.Sample) error {
	d := s.Device
	enc.BeginLine(MeasurementDevice).
		AddTag(TagHost, d.Name).
		AddTag(TagLocation, d.Location).
		AddTag(TagVendor, d.Vendor).
		AddTag(TagRole, d.Role.String()).
		AddStringField(FieldModel, d.Model).
		AddStringField(FieldOSVersion, d.OSVersion).
		AddOptionalFloat(FieldCPUKernel, d.CPUKernel).
		AddOptionalFloat(FieldCPUUser, d.CPUUser).
		AddOptionalUint(FieldMemTotal, d.MemoryTotal).
		AddOptionalUint(FieldMemUsed, d.MemoryUsed)

	if u := d.KernelUptime; u != nil {
		enc.AddUintField(FieldKernelUptimeSeconds, uint64(u.Seconds()))
	}
	if r := d.ResponseTime; r != nil {
		enc.AddUintField(FieldResponseTimeMs, uint64(r.Milliseconds()))
	}

	return enc.EndLine(s.CollectedAt)
}

func writeInterface(enc *Encoder, d model.Device, iface *model.Interface, ts time.Time) error {
	enc.BeginLine(MeasurementInterface).
		AddTag(TagHost, d.Name).
		AddTag(TagInterface, iface.Name).
		AddTag(TagAdminState, iface.AdminState.String()).
		AddTag(TagOperState, iface.OperState.String()).
		AddTag(TagOperMode, iface.OperMode.String()).
		AddTag(TagClassification, iface.Classification.String())

	if p := iface.Peer; p != nil {
		enc.AddTag(TagPeerName, p.Name).
			AddTag(TagPeerType, p.Type.String())
	}

	enc.AddOptionalUint(FieldOperSpeed, iface.OperSpeed).
		AddStringField(FieldDescription, iface.Description).
		AddStringField(FieldDownReason, iface.DownReason)

	if c := iface.Counters; c != nil {
		enc.AddOptionalUint(FieldRxBytes, c.RxBytes).
			AddOptionalUint(FieldTxBytes, c.TxBytes).
			AddOptionalUint(FieldRxUcastPkts, c.RxUcastPkts).
			AddOptionalUint(FieldTxUcastPkts, c.TxUcastPkts).
			AddOptionalUint(FieldRxMcastPkts, c.RxMcastPkts).
			AddOptionalUint(FieldTxMcastPkts, c.TxMcastPkts).
			AddOptionalUint(FieldRxBcastPkts, c.RxBcastPkts).
			AddOptionalUint(FieldTxBcastPkts, c.TxBcastPkts).
			AddOptionalUint(FieldRxCRC, c.RxCRC).
			AddOptionalUint(FieldRxCRCStomped, c.RxCRCStomped).
			AddOptionalUint(FieldRxJumbo, c.RxJumbo).
			AddOptionalUint(FieldTxJumbo, c.TxJumbo)
		writeRxPacketSizes(enc, c.PacketSizeRx)
		writeTxPacketSizes(enc, c.PacketSizeTx)
	}

	return enc.EndLine(ts)
}

func writeQueue(enc *Encoder, d model.Device, iface *model.Interface, q *model.Queue, ts time.Time) error {
	enc.BeginLine(MeasurementQueue).
		AddTag(TagHost, d.Name).
		AddTag(TagInterface, iface.Name).
		AddTag(TagQueueID, strconv.Itoa(q.ID)).
		AddTag(TagQueueName, q.Name).
		AddOptionalUint(FieldQueueTxBytes, q.Counters.TxBytes).
		AddOptionalUint(FieldQueueTxPkts, q.Counters.TxPkts).
		AddOptionalUint(FieldQueueDropBytes, q.Counters.DropBytes).
		AddOptionalUint(FieldQueueDropPkts, q.Counters.DropPkts).
		AddOptionalUint(FieldQueueRandomDropBytes, q.Counters.RandomDropBytes).
		AddOptionalUint(FieldQueueRandomDropPkts, q.Counters.RandomDropPkts).
		AddOptionalUint(FieldQueueCurrentDepth, q.Counters.CurrentDepth).
		AddOptionalUint(FieldQueuePeakDepth, q.Counters.PeakDepth)

	if p := q.PFC; p != nil {
		enc.AddOptionalUint(FieldQueueRxPause, p.RxPause).
			AddOptionalUint(FieldQueueTxPause, p.TxPause).
			AddOptionalUint(FieldQueueWatchdogEvents, p.WatchdogEvents)
	}
	if ecn := q.ECN; ecn != nil {
		enc.AddOptionalUint(FieldQueueECNMarkedPkts, ecn.MarkedPkts)
	}

	return enc.EndLine(ts)
}

func writeBufferInstance(enc *Encoder, d model.Device, b *model.BufferInstance, ts time.Time) error {
	enc.BeginLine(MeasurementBuffer).
		AddTag(TagHost, d.Name).
		AddTag(TagInstanceID, strconv.Itoa(b.ID)).
		AddTag(TagInstanceName, b.Name).
		AddOptionalUint(FieldBufferPeakCellDropPG, b.Counters.PeakCellDropPG).
		AddOptionalUint(FieldBufferPeakCellNoDrop, b.Counters.PeakCellNoDrop).
		AddOptionalUint(FieldBufferCurrentCellDropPG, b.Counters.CurrentCellDropPG).
		AddOptionalUint(FieldBufferCurrentCellNoDrop, b.Counters.CurrentCellNoDrop)

	return enc.EndLine(ts)
}

func writeBurst(enc *Encoder, d model.Device, b *model.BurstEvent) error {
	durationUs := uint64(b.Duration.Microseconds())
	enc.BeginLine(MeasurementBurst).
		AddTag(TagHost, d.Name).
		AddTag(TagInterface, b.Interface).
		AddTag(TagQueueID, strconv.Itoa(b.QueueID)).
		AddOptionalUint(FieldBurstStartDepth, b.StartDepth).
		AddOptionalUint(FieldBurstEndDepth, b.EndDepth).
		AddOptionalUint(FieldBurstPeakDepth, b.PeakDepth).
		AddUintField(FieldBurstDurationUs, durationUs)

	return enc.EndLine(b.PeakTime)
}

func writeRxPacketSizes(enc *Encoder, b *model.PacketSizeBuckets) {
	if b == nil {
		return
	}
	enc.AddOptionalUint(FieldRxPkts64, b.Pkts64).
		AddOptionalUint(FieldRxPkts65to127, b.Pkts65to127).
		AddOptionalUint(FieldRxPkts128to255, b.Pkts128to255).
		AddOptionalUint(FieldRxPkts256to511, b.Pkts256to511).
		AddOptionalUint(FieldRxPkts512to1023, b.Pkts512to1023).
		AddOptionalUint(FieldRxPkts1024to1518, b.Pkts1024to1518).
		AddOptionalUint(FieldRxPkts1519to2047, b.Pkts1519to2047).
		AddOptionalUint(FieldRxPkts2048to4095, b.Pkts2048to4095).
		AddOptionalUint(FieldRxPkts4096to9216, b.Pkts4096to9216)
}

func writeTxPacketSizes(enc *Encoder, b *model.PacketSizeBuckets) {
	if b == nil {
		return
	}
	enc.AddOptionalUint(FieldTxPkts64, b.Pkts64).
		AddOptionalUint(FieldTxPkts65to127, b.Pkts65to127).
		AddOptionalUint(FieldTxPkts128to255, b.Pkts128to255).
		AddOptionalUint(FieldTxPkts256to511, b.Pkts256to511).
		AddOptionalUint(FieldTxPkts512to1023, b.Pkts512to1023).
		AddOptionalUint(FieldTxPkts1024to1518, b.Pkts1024to1518).
		AddOptionalUint(FieldTxPkts1519to2047, b.Pkts1519to2047).
		AddOptionalUint(FieldTxPkts2048to4095, b.Pkts2048to4095).
		AddOptionalUint(FieldTxPkts4096to9216, b.Pkts4096to9216)
}
