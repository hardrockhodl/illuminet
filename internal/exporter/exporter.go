package exporter

import (
	"context"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// Exporter writes a Sample to its configured destination.
//
// Implementations are not required to be safe for concurrent use
// unless explicitly documented to be so.
type Exporter interface {
	// Export writes the sample. Errors from the underlying transport
	// or writer are returned to the caller. Partial writes are
	// possible on error; callers must not assume the write was atomic.
	// A nil sample is a no-op.
	Export(ctx context.Context, sample *model.Sample) error

	// Close flushes any buffered state and releases resources. After
	// Close returns, Export must not be called.
	Close() error
}
