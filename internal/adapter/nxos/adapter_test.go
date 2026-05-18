package nxos

import (
	"context"
	"io"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// mockGNMIServer is a minimal in-process gNMI server used to exercise
// the adapter end-to-end without a live target. It implements the
// Capabilities and Subscribe RPCs that the adapter touches; Get and
// Set come from the UnimplementedGNMIServer embed.
type mockGNMIServer struct {
	gnmi.UnimplementedGNMIServer

	mu             sync.Mutex
	capabilities   *gnmi.CapabilityResponse
	notifications  []*gnmi.Notification
	subscribeCount int
}

func (s *mockGNMIServer) Capabilities(_ context.Context, _ *gnmi.CapabilityRequest) (*gnmi.CapabilityResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.capabilities, nil
}

func (s *mockGNMIServer) Subscribe(stream gnmi.GNMI_SubscribeServer) error {
	if _, err := stream.Recv(); err != nil {
		return err
	}

	s.mu.Lock()
	s.subscribeCount++
	notifs := append([]*gnmi.Notification(nil), s.notifications...)
	s.mu.Unlock()

	for _, n := range notifs {
		if err := stream.Send(&gnmi.SubscribeResponse{
			Response: &gnmi.SubscribeResponse_Update{Update: n},
		}); err != nil {
			return err
		}
	}
	if err := stream.Send(&gnmi.SubscribeResponse{
		Response: &gnmi.SubscribeResponse_SyncResponse{SyncResponse: true},
	}); err != nil {
		return err
	}
	<-stream.Context().Done()
	return nil
}

// setupMockServer starts an in-process gRPC server on an ephemeral
// loopback port and returns the address plus a stop function. The
// mock is preloaded with the supplied notifications.
//
// gnmic's CreateGNMIClient installs its own grpc.WithContextDialer
// after caller-supplied options, which silently overrides any custom
// dialer the test tries to inject. The most robust workaround is to
// run a real loopback TCP listener and let gnmic dial it normally.
func setupMockServer(t *testing.T, notifications []*gnmi.Notification) (*mockGNMIServer, string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	mock := &mockGNMIServer{
		capabilities: &gnmi.CapabilityResponse{
			GNMIVersion: "0.10.0",
			SupportedModels: []*gnmi.ModelData{
				{Name: "openconfig-interfaces", Organization: "OpenConfig", Version: "3.0.0"},
				{Name: "openconfig-system", Organization: "OpenConfig", Version: "2.0.0"},
			},
			SupportedEncodings: []gnmi.Encoding{gnmi.Encoding_PROTO},
		},
		notifications: notifications,
	}
	gnmi.RegisterGNMIServer(srv, mock)

	go func() {
		_ = srv.Serve(lis)
	}()

	return mock, lis.Addr().String(), func() {
		srv.Stop()
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAdapter_DeliversSample(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	notifs := []*gnmi.Notification{
		sampleSystemState("leaf-01", "10.5(1)", ts),
		sampleInterfaceCounters("Ethernet1/1", 1024, 2048, ts),
		sampleInterfaceState("Ethernet1/1", "UP", "UP", "uplink to spine-1", ts),
	}

	_, addr, stop := setupMockServer(t, notifs)
	defer stop()

	cfg := Config{
		Address:  addr,
		Username: "u",
		Password: "p",
		Insecure: true,
		Interval: 80 * time.Millisecond,
		Timeout:  500 * time.Millisecond,
		Location: "lab-rack-A",
	}
	a, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan *model.Sample, 4)
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx, out) }()

	var s *model.Sample
	select {
	case s = <-out:
	case <-time.After(2 * time.Second):
		t.Fatal("no sample received within timeout")
	}

	if s.Device.Name != "leaf-01" {
		t.Errorf("Device.Name: got %q, want leaf-01", s.Device.Name)
	}
	if s.Device.OSVersion != "10.5(1)" {
		t.Errorf("Device.OSVersion: got %q, want 10.5(1)", s.Device.OSVersion)
	}
	if s.Device.Vendor != "cisco" {
		t.Errorf("Device.Vendor: got %q", s.Device.Vendor)
	}
	if s.Device.Location != "lab-rack-A" {
		t.Errorf("Device.Location: got %q, want lab-rack-A", s.Device.Location)
	}

	if len(s.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(s.Interfaces))
	}
	iface := s.Interfaces[0]
	if iface.Name != "Ethernet1/1" {
		t.Errorf("Interface.Name: got %q", iface.Name)
	}
	if iface.AdminState != model.AdminStateUp {
		t.Errorf("AdminState: got %q", iface.AdminState)
	}
	if iface.Counters == nil {
		t.Fatal("Counters: nil")
	}
	if iface.Counters.RxBytes == nil || *iface.Counters.RxBytes != 1024 {
		t.Errorf("RxBytes: got %v, want 1024", iface.Counters.RxBytes)
	}
	if iface.Counters.TxBytes == nil || *iface.Counters.TxBytes != 2048 {
		t.Errorf("TxBytes: got %v, want 2048", iface.Counters.TxBytes)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Run did not return after cancel")
	}
}

func TestAdapter_ShutsDownPromptlyOnCancel(t *testing.T) {
	_, addr, stop := setupMockServer(t, nil)
	defer stop()

	cfg := Config{
		Address:  addr,
		Username: "u",
		Insecure: true,
		Interval: 100 * time.Millisecond,
		Timeout:  500 * time.Millisecond,
	}
	a, err := New(cfg, discardLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan *model.Sample, 4)
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx, out) }()

	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return within 2s of cancel")
	}
}

func TestAdapter_NewRejectsBadConfig(t *testing.T) {
	_, err := New(Config{}, nil)
	if err == nil {
		t.Fatal("expected error from empty config, got nil")
	}
}
