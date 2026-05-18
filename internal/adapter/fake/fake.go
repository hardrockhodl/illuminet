package fake

import (
	"context"
	"log/slog"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/hardrockhodl/illuminet/internal/adapter"
	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// fakeSeed is the constant seed for the deterministic jitter PRNG.
// Two fake instances constructed with the same interval produce
// identical counter sequences because they start from this seed.
const fakeSeed uint64 = 42

type fake struct {
	interval time.Duration
	logger   *slog.Logger

	mu   sync.Mutex
	rand *rand.Rand
	tick int

	eth1Rx       uint64
	eth1Tx       uint64
	eth1QTx      uint64
	eth1RxPause  uint64
	eth49Rx      uint64
	eth49Tx      uint64
	eth49QTx     uint64
	eth49TxPause uint64

	bufPeakDrop   uint64
	bufPeakNoDrop uint64
}

// New constructs a fake adapter that emits deterministic Sample data
// on the given interval. A nil logger is replaced by slog.Default
// inside the underlying PollingAdapter.
func New(interval time.Duration, logger *slog.Logger) adapter.Adapter {
	return newFake(interval, logger)
}

// newFake returns the concrete *fake so package-internal tests can
// drive the fetch path directly.
func newFake(interval time.Duration, logger *slog.Logger) *fake {
	return &fake{
		interval: interval,
		logger:   logger,
		rand:     rand.New(rand.NewPCG(fakeSeed, fakeSeed)),
	}
}

// Name returns "fake".
func (f *fake) Name() string { return "fake" }

// Run wraps the fetch loop in a PollingAdapter and runs until ctx is
// cancelled.
func (f *fake) Run(ctx context.Context, out chan<- *model.Sample) error {
	return adapter.NewPolling(f.Name(), f.interval, f.fetch, f.logger).Run(ctx, out)
}

// fetch produces one Sample, advancing internal counters by an amount
// proportional to the configured interval (with ±10% jitter).
func (f *fake) fetch(_ context.Context) (*model.Sample, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.tick++
	sec := f.interval.Seconds()

	delta := func(ratePerSec float64) uint64 {
		return uint64(ratePerSec * sec * f.jitterFactor())
	}

	eth1RxDelta := delta(1_000_000)
	eth1TxDelta := delta(800_000)
	eth49RxDelta := delta(10_000_000)
	eth49TxDelta := delta(8_000_000)

	f.eth1Rx += eth1RxDelta
	f.eth1Tx += eth1TxDelta
	f.eth1QTx += eth1TxDelta / 100
	f.eth1RxPause += uint64(f.jitterFactor() * 2)

	f.eth49Rx += eth49RxDelta
	f.eth49Tx += eth49TxDelta
	f.eth49QTx += eth49TxDelta / 100
	f.eth49TxPause += uint64(f.jitterFactor() * 2)

	f.bufPeakDrop += uint64(f.jitterFactor() * 10)
	f.bufPeakNoDrop += uint64(f.jitterFactor() * 5)

	now := time.Now()
	cpuKernel := f.cpuLoad()
	cpuUser := cpuKernel * 0.4

	return f.buildSample(now, cpuKernel, cpuUser), nil
}

// cpuLoad returns the synthetic CPU utilization for the current tick.
// The wave is centered at 30% with ±20% amplitude and a 60-second
// period, so a Grafana panel set to a 5-minute window sees five
// complete oscillations.
func (f *fake) cpuLoad() float64 {
	elapsed := float64(f.tick) * f.interval.Seconds()
	return 30 + 20*math.Sin(2*math.Pi*elapsed/60)
}

// jitterFactor returns a value in [0.9, 1.1).
func (f *fake) jitterFactor() float64 {
	return 0.9 + 0.2*f.rand.Float64()
}

func (f *fake) buildSample(now time.Time, cpuKernel, cpuUser float64) *model.Sample {
	// Copy state into local variables before taking addresses, so the
	// pointers in the emitted Sample remain stable when fetch advances
	// the adapter's internal state on subsequent ticks.
	eth1Rx := f.eth1Rx
	eth1Tx := f.eth1Tx
	eth1QTx := f.eth1QTx
	eth1RxPause := f.eth1RxPause
	eth49Rx := f.eth49Rx
	eth49Tx := f.eth49Tx
	eth49QTx := f.eth49QTx
	eth49TxPause := f.eth49TxPause
	bufPeakDrop := f.bufPeakDrop
	bufPeakNoDrop := f.bufPeakNoDrop

	speed25G := uint64(25_000_000_000)
	speed100G := uint64(100_000_000_000)

	return &model.Sample{
		Device: model.Device{
			Name:         "fake-switch",
			ManagementIP: "10.0.0.1",
			Vendor:       "fake",
			Model:        "FAKE-9000",
			OSVersion:    "1.0.0",
			Role:         model.DeviceRoleLeaf,
			Location:     "lab",
			CPUKernel:    &cpuKernel,
			CPUUser:      &cpuUser,
		},
		Interfaces: []model.Interface{
			{
				Name:           "Ethernet1/1",
				Description:    "host-facing edge port",
				AdminState:     model.AdminStateUp,
				OperState:      model.OperStateUp,
				OperMode:       model.OperModeAccess,
				Classification: model.PortClassificationEdge,
				OperSpeed:      &speed25G,
				Peer: &model.Peer{
					Name:       "server-001",
					Type:       model.PeerTypeHost,
					LearnedVia: "lldp",
				},
				Counters: &model.InterfaceCounters{
					RxBytes: &eth1Rx,
					TxBytes: &eth1Tx,
				},
				Queues: []model.Queue{{
					ID:       0,
					Name:     "default",
					Counters: model.QueueCounters{TxPkts: &eth1QTx},
					PFC:      &model.PFCCounter{RxPause: &eth1RxPause},
				}},
				ObservedAt: now,
			},
			{
				Name:           "Ethernet1/49",
				Description:    "fabric uplink",
				AdminState:     model.AdminStateUp,
				OperState:      model.OperStateUp,
				OperMode:       model.OperModeRouted,
				Classification: model.PortClassificationCore,
				OperSpeed:      &speed100G,
				Peer: &model.Peer{
					Name:       "spine-1",
					Type:       model.PeerTypeSwitch,
					LearnedVia: "lldp",
				},
				Counters: &model.InterfaceCounters{
					RxBytes: &eth49Rx,
					TxBytes: &eth49Tx,
				},
				Queues: []model.Queue{{
					ID:       0,
					Name:     "default",
					Counters: model.QueueCounters{TxPkts: &eth49QTx},
					PFC:      &model.PFCCounter{TxPause: &eth49TxPause},
				}},
				ObservedAt: now,
			},
		},
		Buffers: []model.BufferInstance{{
			ID:   1,
			Name: "module-1",
			Counters: model.BufferCounters{
				PeakCellDropPG: &bufPeakDrop,
				PeakCellNoDrop: &bufPeakNoDrop,
			},
		}},
		// Bursts intentionally left empty in this iteration. Future
		// scenarios can inject synthetic BurstEvents here when we want
		// to exercise the burst path end-to-end.
		CollectedAt: now,
	}
}
