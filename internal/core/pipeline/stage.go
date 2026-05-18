package pipeline

import (
	"context"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// Stage processes a Sample between adapter receipt and exporter
// dispatch. Stages run in sequence, in the order configured in
// [Options]. Each stage may modify the Sample in place.
//
// Stages should be pure transformations on the Sample data: idempotent
// and free of side effects on external systems. A stage that returns
// an error is logged at warn level by the pipeline and the remaining
// stages still run on the (possibly partial) Sample. This is
// intentional best-effort behavior: one broken stage must not prevent
// the rest of the pipeline from working.
type Stage interface {
	// Name returns a short identifier used in log entries and
	// diagnostics.
	Name() string

	// Process transforms the Sample in place. ctx may carry
	// cancellation; long-running stages should check it. A nil
	// sample is a programmer error and stages may panic on it.
	Process(ctx context.Context, sample *model.Sample) error
}
