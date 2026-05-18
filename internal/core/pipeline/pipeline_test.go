package pipeline

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/hardrockhodl/illuminet/internal/adapter"
	"github.com/hardrockhodl/illuminet/internal/adapter/fake"
	"github.com/hardrockhodl/illuminet/internal/core/model"
	"github.com/hardrockhodl/illuminet/internal/exporter"
)

// recordingExporter is a test-only exporter that captures every sample
// it receives and can be configured to fail Export deterministically.
type recordingExporter struct {
	mu         sync.Mutex
	samples    []*model.Sample
	closed     bool
	failExport bool
}

func (e *recordingExporter) Export(_ context.Context, s *model.Sample) error {
	if e.failExport {
		return errors.New("recordingExporter: configured to fail")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.samples = append(e.samples, s)
	return nil
}

func (e *recordingExporter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	return nil
}

func (e *recordingExporter) count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.samples)
}

func (e *recordingExporter) isClosed() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.closed
}

// stubAdapter blocks on ctx.Done; useful for tests that exercise
// Pipeline construction without producing samples.
type stubAdapter struct{}

func (s *stubAdapter) Name() string { return "stub" }
func (s *stubAdapter) Run(ctx context.Context, _ chan<- *model.Sample) error {
	<-ctx.Done()
	return nil
}

func TestPipeline_New_RequiresAdapters(t *testing.T) {
	_, err := New(Options{
		Exporters: []exporter.Exporter{&recordingExporter{}},
	})
	if err == nil {
		t.Fatal("expected error when no adapters configured")
	}
}

func TestPipeline_New_RequiresExporters(t *testing.T) {
	_, err := New(Options{
		Adapters: []adapter.Adapter{&stubAdapter{}},
	})
	if err == nil {
		t.Fatal("expected error when no exporters configured")
	}
}

func TestPipeline_Run_DeliversSamples(t *testing.T) {
	e := &recordingExporter{}
	p, err := New(Options{
		Adapters:  []adapter.Adapter{fake.New(20*time.Millisecond, nil)},
		Exporters: []exporter.Exporter{e},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := p.Run(ctx); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if e.count() < 1 {
		t.Errorf("expected at least one sample delivered, got %d", e.count())
	}
	if !e.isClosed() {
		t.Error("exporter Close was not called after shutdown")
	}
}

func TestPipeline_Run_FailingExporterDoesNotBlockOthers(t *testing.T) {
	failing := &recordingExporter{failExport: true}
	healthy := &recordingExporter{}

	p, err := New(Options{
		Adapters:  []adapter.Adapter{fake.New(20*time.Millisecond, nil)},
		Exporters: []exporter.Exporter{failing, healthy},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := p.Run(ctx); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if healthy.count() < 1 {
		t.Errorf("healthy exporter received no samples (failing exporter blocked dispatch): %d", healthy.count())
	}
	if !failing.isClosed() || !healthy.isClosed() {
		t.Error("Close not called on both exporters")
	}
}

func TestPipeline_Run_ShutdownTimely(t *testing.T) {
	e := &recordingExporter{}
	p, err := New(Options{
		Adapters:  []adapter.Adapter{fake.New(10*time.Millisecond, nil)},
		Exporters: []exporter.Exporter{e},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error on cancellation: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("pipeline did not shut down within 1 second of context cancellation")
	}

	if !e.isClosed() {
		t.Error("exporter Close was not called after shutdown")
	}
}
