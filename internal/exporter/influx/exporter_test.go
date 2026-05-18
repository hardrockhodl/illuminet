package influx

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

func ptrUint(v uint64) *uint64 { return &v }

func TestExporter_EmptySample_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	exp := New(&buf)
	if err := exp.Export(context.Background(), &model.Sample{}); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestExporter_NilSample(t *testing.T) {
	var buf bytes.Buffer
	exp := New(&buf)
	if err := exp.Export(context.Background(), nil); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestExporter_MinimalSample_ExactOutput(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	rx := uint64(1000)

	s := &model.Sample{
		Device: model.Device{
			Name:     "leaf-01",
			Location: "rack-A",
			Model:    "TestModel",
		},
		Interfaces: []model.Interface{{
			Name:     "Ethernet1/1",
			Counters: &model.InterfaceCounters{RxBytes: &rx},
		}},
		CollectedAt: ts,
	}

	var buf bytes.Buffer
	exp := New(&buf)
	if err := exp.Export(context.Background(), s); err != nil {
		t.Fatalf("Export: %v", err)
	}

	nano := ts.UnixNano()
	want := fmt.Sprintf("device,host=leaf-01,location=rack-A,role=unknown,vendor=unknown model=\"TestModel\" %d\n", nano) +
		fmt.Sprintf("interface,admin_state=unknown,classification=unknown,host=leaf-01,interface=Ethernet1/1,oper_mode=unknown,oper_state=unknown rx_bytes=1000i %d\n", nano)

	if buf.String() != want {
		t.Errorf("output mismatch:\ngot:  %q\nwant: %q", buf.String(), want)
	}
}

func TestExporter_PeerInfoAsTags(t *testing.T) {
	s := &model.Sample{
		Device: model.Device{Name: "h1"},
		Interfaces: []model.Interface{{
			Name: "eth1",
			Peer: &model.Peer{
				Name: "server-42",
				Type: model.PeerTypeHost,
			},
			Counters: &model.InterfaceCounters{RxBytes: ptrUint(100)},
		}},
		CollectedAt: time.Unix(1, 0),
	}

	var buf bytes.Buffer
	exp := New(&buf)
	if err := exp.Export(context.Background(), s); err != nil {
		t.Fatalf("Export: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "peer_name=server-42") {
		t.Errorf("missing peer_name tag in %q", out)
	}
	if !strings.Contains(out, "peer_type=host") {
		t.Errorf("missing peer_type tag in %q", out)
	}
}

func TestExporter_QueuePFCSeparation(t *testing.T) {
	rxPause := uint64(50)
	s := &model.Sample{
		Device: model.Device{Name: "h1"},
		Interfaces: []model.Interface{{
			Name:     "eth1",
			Counters: &model.InterfaceCounters{RxBytes: ptrUint(100)},
			Queues: []model.Queue{{
				ID:       3,
				Counters: model.QueueCounters{TxPkts: ptrUint(7)},
				PFC:      &model.PFCCounter{RxPause: &rxPause},
			}},
		}},
		CollectedAt: time.Unix(1, 0),
	}

	var buf bytes.Buffer
	exp := New(&buf)
	if err := exp.Export(context.Background(), s); err != nil {
		t.Fatalf("Export: %v", err)
	}

	var ifaceLine, queueLine string
	for _, l := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		switch {
		case strings.HasPrefix(l, "interface,"):
			ifaceLine = l
		case strings.HasPrefix(l, "queue,"):
			queueLine = l
		}
	}
	if ifaceLine == "" {
		t.Fatalf("no interface line in output:\n%s", buf.String())
	}
	if queueLine == "" {
		t.Fatalf("no queue line in output:\n%s", buf.String())
	}
	if strings.Contains(ifaceLine, "rx_pause") {
		t.Errorf("rx_pause leaked into interface line: %q", ifaceLine)
	}
	if !strings.Contains(queueLine, "rx_pause=50i") {
		t.Errorf("queue line missing rx_pause=50i: %q", queueLine)
	}
	if !strings.Contains(queueLine, "queue_id=3") {
		t.Errorf("queue line missing queue_id tag: %q", queueLine)
	}
}

func TestExporter_BurstUsesPeakTime(t *testing.T) {
	collected := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	peak := time.Date(2026, 5, 18, 11, 59, 50, 0, time.UTC)

	s := &model.Sample{
		Device: model.Device{Name: "h1"},
		Bursts: []model.BurstEvent{{
			Interface: "eth1",
			QueueID:   2,
			Duration:  500 * time.Microsecond,
			PeakTime:  peak,
			PeakDepth: ptrUint(99999),
		}},
		CollectedAt: collected,
	}

	var buf bytes.Buffer
	exp := New(&buf)
	if err := exp.Export(context.Background(), s); err != nil {
		t.Fatalf("Export: %v", err)
	}

	var burstLine string
	for _, l := range strings.Split(buf.String(), "\n") {
		if strings.HasPrefix(l, "burst,") {
			burstLine = l
			break
		}
	}
	if burstLine == "" {
		t.Fatalf("no burst line in %q", buf.String())
	}

	burstNano := strconv.FormatInt(peak.UnixNano(), 10)
	collectedNano := strconv.FormatInt(collected.UnixNano(), 10)
	if !strings.HasSuffix(burstLine, " "+burstNano) {
		t.Errorf("burst line %q does not end with peak timestamp %s", burstLine, burstNano)
	}
	if strings.Contains(burstLine, " "+collectedNano) {
		t.Errorf("burst line %q wrongly contains collected timestamp", burstLine)
	}
	if !strings.Contains(burstLine, "duration_us=500i") {
		t.Errorf("burst line missing duration_us=500i: %q", burstLine)
	}
}

func TestExporter_ComplexSampleRoundtrip(t *testing.T) {
	collected := time.Unix(1700000000, 0).UTC()
	observed := collected.Add(50 * time.Millisecond)
	peak := collected.Add(-2 * time.Second)
	cpu := 42.5
	mem := uint64(8 << 30)
	uptime := 72 * time.Hour
	responseTime := 120 * time.Millisecond

	s := &model.Sample{
		Device: model.Device{
			Name:         "spine-1",
			ManagementIP: "10.0.0.1",
			Vendor:       model.VendorCisco,
			Model:        "N9K-C9332D-GX2B",
			OSVersion:    "10.5(1)",
			Location:     "rack=A 1",
			Role:         model.DeviceRoleSpine,
			KernelUptime: &uptime,
			CPUKernel:    &cpu,
			MemoryTotal:  &mem,
			MemoryUsed:   ptrUint(1 << 30),
			ResponseTime: &responseTime,
		},
		Interfaces: []model.Interface{{
			Name:           "Ethernet1/1",
			Description:    `link to "leaf-1"`,
			AdminState:     model.AdminStateUp,
			OperState:      model.OperStateUp,
			OperMode:       model.OperModeRouted,
			Classification: model.PortClassificationCore,
			OperSpeed:      ptrUint(100_000_000_000),
			Peer: &model.Peer{
				Name: "leaf-1",
				Type: model.PeerTypeSwitch,
			},
			Counters: &model.InterfaceCounters{
				RxBytes:     ptrUint(1 << 40),
				TxBytes:     ptrUint(2 << 40),
				RxUcastPkts: ptrUint(100),
				PacketSizeRx: &model.PacketSizeBuckets{
					Pkts64:      ptrUint(10),
					Pkts65to127: ptrUint(20),
				},
			},
			Queues: []model.Queue{{
				ID:       0,
				Name:     "default",
				Counters: model.QueueCounters{TxPkts: ptrUint(1234), DropPkts: ptrUint(5)},
				PFC:      &model.PFCCounter{RxPause: ptrUint(3), TxPause: ptrUint(7)},
				ECN:      &model.ECNCounter{MarkedPkts: ptrUint(11)},
			}},
			ObservedAt: observed,
		}},
		Buffers: []model.BufferInstance{{
			ID:   1,
			Name: "slice-0",
			Counters: model.BufferCounters{
				CurrentCellDropPG: ptrUint(42),
				PeakCellNoDrop:    ptrUint(99),
			},
		}},
		Bursts: []model.BurstEvent{{
			Interface: "Ethernet1/1",
			QueueID:   0,
			PeakDepth: ptrUint(123456),
			Duration:  2 * time.Millisecond,
			PeakTime:  peak,
		}},
		CollectedAt: collected,
	}

	var buf bytes.Buffer
	exp := New(&buf)
	if err := exp.Export(context.Background(), s); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if err := exp.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	out := buf.String()

	wantFragments := []string{
		"device,",
		`host=spine-1`,
		`location=rack\=A\ 1`,
		"vendor=cisco",
		"role=spine",
		`model="N9K-C9332D-GX2B"`,
		`os_version="10.5(1)"`,
		"cpu_kernel=42.5",
		"kernel_uptime_seconds=259200i",
		"response_time_ms=120i",
		"interface,",
		"admin_state=up",
		"oper_state=up",
		"oper_mode=routed",
		"classification=core",
		"peer_name=leaf-1",
		"peer_type=switch",
		`description="link to \"leaf-1\""`,
		"oper_speed=100000000000i",
		"rx_bytes=",
		"rx_pkts_64=10i",
		"queue,",
		"queue_id=0",
		"queue_name=default",
		"rx_pause=3i",
		"tx_pause=7i",
		"ecn_marked_pkts=11i",
		"buffer,",
		"instance_id=1",
		"instance_name=slice-0",
		"current_cell_drop_pg=42i",
		"peak_cell_no_drop=99i",
		"burst,",
		"queue_id=0",
		"peak_depth=123456i",
		"duration_us=2000i",
	}
	for _, frag := range wantFragments {
		if !strings.Contains(out, frag) {
			t.Errorf("output missing fragment %q\nfull output:\n%s", frag, out)
		}
	}

	// Sanity check: every non-empty line should match the basic
	// "measurement[,tags] fields timestamp" shape (three space-
	// separated parts after splitting on the *first* two spaces, or
	// two parts when no tags are present).
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		// At minimum the line must contain " " (field separator) and
		// end with a numeric timestamp.
		parts := strings.Split(line, " ")
		if len(parts) < 3 {
			t.Errorf("line %q does not have at least three space-separated segments", line)
			continue
		}
		if _, err := strconv.ParseInt(parts[len(parts)-1], 10, 64); err != nil {
			t.Errorf("line %q does not end with a numeric timestamp: %v", line, err)
		}
	}
}

func TestExporter_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	exp := New(&buf)
	err := exp.Export(ctx, &model.Sample{Device: model.Device{Name: "h1"}})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if buf.Len() != 0 {
		t.Errorf("output written after cancellation: %q", buf.String())
	}
}

func TestExporter_InterfaceObservedAtFallback(t *testing.T) {
	collected := time.Unix(1700000000, 0).UTC()
	s := &model.Sample{
		Device: model.Device{Name: "h1"},
		Interfaces: []model.Interface{{
			Name:     "eth1",
			Counters: &model.InterfaceCounters{RxBytes: ptrUint(1)},
			// ObservedAt left zero — exporter should fall back to CollectedAt.
		}},
		CollectedAt: collected,
	}

	var buf bytes.Buffer
	exp := New(&buf)
	if err := exp.Export(context.Background(), s); err != nil {
		t.Fatalf("Export: %v", err)
	}

	want := " " + strconv.FormatInt(collected.UnixNano(), 10) + "\n"
	if !strings.Contains(buf.String(), want) {
		t.Errorf("expected interface line to fall back to CollectedAt timestamp %q, got:\n%s", want, buf.String())
	}
}
