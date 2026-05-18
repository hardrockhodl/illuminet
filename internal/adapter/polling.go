package adapter

import (
	"context"
	"log/slog"
	"time"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// PollFunc fetches a single Sample. It is invoked repeatedly by
// [PollingAdapter] on a fixed interval. Returning a nil Sample with a
// nil error signals "no data this tick" and is silently ignored.
type PollFunc func(ctx context.Context) (*model.Sample, error)

// PollingAdapter wraps a PollFunc and exposes it as an [Adapter] that
// polls on a fixed interval.
//
// The first fetch is invoked immediately when Run starts; subsequent
// fetches are driven by a time.Ticker. PollFunc errors are logged at
// warn level and do not terminate the adapter; only ctx cancellation
// causes Run to return.
type PollingAdapter struct {
	name     string
	interval time.Duration
	fetch    PollFunc
	logger   *slog.Logger
}

// NewPolling constructs a PollingAdapter. A nil logger is replaced by
// [slog.Default].
func NewPolling(name string, interval time.Duration, fetch PollFunc, logger *slog.Logger) *PollingAdapter {
	if logger == nil {
		logger = slog.Default()
	}
	return &PollingAdapter{
		name:     name,
		interval: interval,
		fetch:    fetch,
		logger:   logger,
	}
}

// Name returns the adapter's identifier as supplied to NewPolling.
func (p *PollingAdapter) Name() string { return p.name }

// Run drives the poll loop until ctx is cancelled. It always returns
// nil: cancellation is the expected termination signal, not an error.
func (p *PollingAdapter) Run(ctx context.Context, out chan<- *model.Sample) error {
	p.fetchOnce(ctx, out)
	if ctx.Err() != nil {
		return nil
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			p.fetchOnce(ctx, out)
		}
	}
}

// fetchOnce invokes the PollFunc and forwards a non-nil sample to out.
// Errors are logged but otherwise ignored. The send on out is
// cancellable so a wedged consumer cannot prevent shutdown.
func (p *PollingAdapter) fetchOnce(ctx context.Context, out chan<- *model.Sample) {
	sample, err := p.fetch(ctx)
	if err != nil {
		p.logger.Warn("adapter fetch failed", "adapter", p.name, "error", err)
		return
	}
	if sample == nil {
		return
	}
	select {
	case out <- sample:
	case <-ctx.Done():
	}
}
