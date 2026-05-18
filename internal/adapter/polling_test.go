package adapter

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

func TestPollingAdapter_ImmediateFirstFetch(t *testing.T) {
	var calls atomic.Int32
	fetch := func(_ context.Context) (*model.Sample, error) {
		calls.Add(1)
		return nil, nil
	}
	pa := NewPolling("test", time.Hour, fetch, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan *model.Sample, 1)
	done := make(chan error, 1)
	go func() { done <- pa.Run(ctx, out) }()

	// Poll the call counter briefly; the immediate first fetch should
	// happen well within the very long ticker interval.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) && calls.Load() == 0 {
		time.Sleep(5 * time.Millisecond)
	}
	if got := calls.Load(); got == 0 {
		t.Fatal("PollFunc was not called immediately at Run start")
	}
	cancel()
	if err := <-done; err != nil {
		t.Errorf("Run returned error on cancellation: %v", err)
	}
}

func TestPollingAdapter_IntervalPolling(t *testing.T) {
	var calls atomic.Int32
	fetch := func(_ context.Context) (*model.Sample, error) {
		calls.Add(1)
		return nil, nil
	}
	pa := NewPolling("test", 30*time.Millisecond, fetch, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	out := make(chan *model.Sample, 1)
	if err := pa.Run(ctx, out); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Expect at least 3 calls: one immediate + at least 2 ticks within
	// the 200ms window at 30ms interval (~6 ticks ideal).
	if n := calls.Load(); n < 3 {
		t.Errorf("expected >=3 calls, got %d", n)
	}
}

func TestPollingAdapter_ErrorContinues(t *testing.T) {
	var calls atomic.Int32
	fetch := func(_ context.Context) (*model.Sample, error) {
		calls.Add(1)
		return nil, errors.New("fetch boom")
	}
	pa := NewPolling("test", 20*time.Millisecond, fetch, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	out := make(chan *model.Sample, 1)
	if err := pa.Run(ctx, out); err != nil {
		t.Fatalf("Run returned error despite transient fetch failures: %v", err)
	}
	if n := calls.Load(); n < 2 {
		t.Errorf("expected loop to keep running through errors, got %d calls", n)
	}
}

func TestPollingAdapter_CtxCancelReturnsNil(t *testing.T) {
	fetch := func(_ context.Context) (*model.Sample, error) {
		return nil, nil
	}
	pa := NewPolling("test", time.Hour, fetch, nil)

	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan *model.Sample, 1)
	done := make(chan error, 1)
	go func() { done <- pa.Run(ctx, out) }()

	// Give the immediate fetch a moment to run, then cancel.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil error on cancellation, got %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Run did not return after cancellation")
	}
}

func TestPollingAdapter_FullChannelShutdown(t *testing.T) {
	fetch := func(_ context.Context) (*model.Sample, error) {
		return &model.Sample{Device: model.Device{Name: "x"}}, nil
	}
	pa := NewPolling("test", 10*time.Millisecond, fetch, nil)

	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan *model.Sample) // unbuffered, never read

	done := make(chan error, 1)
	go func() { done <- pa.Run(ctx, out) }()

	// Let the adapter try to send and block on the unread channel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil error on cancellation, got %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Run did not return after cancellation despite full channel")
	}
}
