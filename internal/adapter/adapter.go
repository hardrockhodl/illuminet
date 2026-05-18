package adapter

import (
	"context"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// Adapter is a source of Sample data. Adapters run in their own
// goroutine and push samples to a pipeline-owned channel until the
// context is cancelled.
//
// Implementations must not close the output channel; the pipeline
// owns its lifecycle. Implementations must return from Run promptly
// when ctx is done. Implementations are expected to recover from
// transient errors internally and to return an error from Run only
// when the adapter cannot continue.
type Adapter interface {
	// Name returns a short identifier for logging and diagnostics,
	// for example "fake" or "nxos".
	Name() string

	// Run starts the adapter and blocks until ctx is cancelled or a
	// fatal error occurs. Samples are pushed to out as they become
	// available.
	Run(ctx context.Context, out chan<- *model.Sample) error
}
