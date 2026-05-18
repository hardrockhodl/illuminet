package pipeline

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/hardrockhodl/illuminet/internal/adapter"
	"github.com/hardrockhodl/illuminet/internal/core/model"
	"github.com/hardrockhodl/illuminet/internal/exporter"
)

const (
	defaultBuffer        = 64
	shutdownDrainTimeout = 5 * time.Second
)

// Pipeline runs adapters and dispatches their samples to exporters.
// It is safe to construct one Pipeline per process and call Run once.
type Pipeline struct {
	adapters  []adapter.Adapter
	exporters []exporter.Exporter
	logger    *slog.Logger
	buffer    int
}

// Options configures a Pipeline. A zero or negative Buffer is replaced
// by the package default (64). A nil Logger is replaced by
// slog.Default.
type Options struct {
	Adapters  []adapter.Adapter
	Exporters []exporter.Exporter
	Logger    *slog.Logger
	Buffer    int
}

// New constructs a Pipeline from Options. It returns an error when no
// adapter or no exporter is configured.
func New(opts Options) (*Pipeline, error) {
	if len(opts.Adapters) == 0 {
		return nil, errors.New("pipeline: at least one adapter required")
	}
	if len(opts.Exporters) == 0 {
		return nil, errors.New("pipeline: at least one exporter required")
	}
	buf := opts.Buffer
	if buf <= 0 {
		buf = defaultBuffer
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Pipeline{
		adapters:  opts.Adapters,
		exporters: opts.Exporters,
		logger:    logger,
		buffer:    buf,
	}, nil
}

// Run starts all adapters and blocks until ctx is cancelled or every
// adapter terminates. On shutdown, Run waits up to ~5 seconds for
// adapters to drain, then closes the sample channel and calls Close
// on every exporter. It returns the first fatal adapter error, or
// nil.
func (p *Pipeline) Run(ctx context.Context) error {
	samples := make(chan *model.Sample, p.buffer)

	var adapterWG sync.WaitGroup
	errCh := make(chan error, len(p.adapters))

	for _, a := range p.adapters {
		adapterWG.Add(1)
		go func(a adapter.Adapter) {
			defer adapterWG.Done()
			if err := a.Run(ctx, samples); err != nil {
				p.logger.Error("adapter terminated with error", "adapter", a.Name(), "error", err)
				errCh <- err
			}
		}(a)
	}

	dispatchDone := make(chan struct{})
	go func() {
		defer close(dispatchDone)
		// The dispatcher uses a fresh context so that, during
		// shutdown, buffered samples are still delivered to exporters.
		// Adapters already saw the original context's cancellation
		// and have begun returning; the dispatcher only finishes when
		// the sample channel closes.
		exportCtx := context.Background()
		for s := range samples {
			for _, exp := range p.exporters {
				if err := exp.Export(exportCtx, s); err != nil {
					p.logger.Error("export failed", "error", err)
				}
			}
		}
	}()

	<-ctx.Done()

	drained := make(chan struct{})
	go func() {
		adapterWG.Wait()
		close(drained)
	}()

	select {
	case <-drained:
		close(samples)
		<-dispatchDone
	case <-time.After(shutdownDrainTimeout):
		// Adapters did not return within the drain window. Closing
		// samples would race their pending sends and panic; leak the
		// goroutines and the channel. The process is exiting anyway.
		p.logger.Warn("adapter drain timeout, leaking goroutines on shutdown",
			"timeout", shutdownDrainTimeout)
	}

	for _, exp := range p.exporters {
		if err := exp.Close(); err != nil {
			p.logger.Error("exporter close failed", "error", err)
		}
	}

	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
