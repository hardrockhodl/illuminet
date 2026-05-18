package nxos

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmic/pkg/api"
	"github.com/openconfig/gnmic/pkg/api/target"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// Adapter is the NX-OS gNMI adapter. It maintains a long-lived
// subscription to one target and pushes Samples onto the pipeline.
type Adapter struct {
	cfg    Config
	logger *slog.Logger
}

// New constructs an Adapter. It validates the supplied Config and
// applies package defaults to zero-valued fields.
func New(cfg Config, logger *slog.Logger) (*Adapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.applyDefaults()
	if logger == nil {
		logger = slog.Default()
	}
	return &Adapter{cfg: cfg, logger: logger}, nil
}

// Name returns the adapter identifier "nxos".
func (a *Adapter) Name() string { return "nxos" }

// Run connects to the gNMI target, subscribes to the configured paths,
// and pushes a Sample to out on every cfg.Interval tick. Returns nil
// when ctx is cancelled; returns an error only on fatal setup failure
// (connection refused, bad credentials at handshake, malformed
// subscribe request).
func (a *Adapter) Run(ctx context.Context, out chan<- *model.Sample) error {
	tgt, err := a.newTarget()
	if err != nil {
		return fmt.Errorf("nxos: create target: %w", err)
	}
	defer func() {
		if err := tgt.Close(); err != nil {
			a.logger.Warn("nxos: target close failed", "error", err)
		}
	}()

	if err := a.connectAndProbe(ctx, tgt); err != nil {
		return err
	}

	subReq, err := api.NewSubscribeRequest(
		api.Encoding("proto"),
		api.SubscriptionListMode("stream"),
		api.Subscription(
			api.Path(PathInterfaces),
			api.SubscriptionMode("sample"),
			api.SampleInterval(a.cfg.Interval),
		),
		api.Subscription(
			api.Path(PathSystem),
			api.SubscriptionMode("sample"),
			api.SampleInterval(a.cfg.Interval),
		),
	)
	if err != nil {
		return fmt.Errorf("nxos: build subscribe request: %w", err)
	}

	respCh, errCh := tgt.SubscribeStreamChan(ctx, subReq, "nxos-main")

	cache := newNotificationCache()
	ticker := time.NewTicker(a.cfg.Interval)
	defer ticker.Stop()

	devCtx := DeviceContext{
		Name:         hostFromAddress(a.cfg.Address),
		ManagementIP: hostFromAddress(a.cfg.Address),
		Location:     a.cfg.Location,
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case resp, ok := <-respCh:
			if !ok {
				a.logger.Info("nxos: subscription channel closed")
				return nil
			}
			if n := resp.GetUpdate(); n != nil {
				cache.ingest(n)
			}

		case err, ok := <-errCh:
			if !ok {
				continue
			}
			if err != nil {
				a.logger.Warn("nxos: subscription error", "error", err)
			}

		case <-ticker.C:
			notification := cache.snapshot()
			if notification == nil {
				continue
			}
			sample, terr := Translate(devCtx, []*gnmi.Notification{notification})
			if terr != nil {
				a.logger.Error("nxos: translate failed", "error", terr)
				continue
			}
			select {
			case out <- sample:
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func (a *Adapter) newTarget() (*target.Target, error) {
	opts := []api.TargetOption{
		api.Name("nxos"),
		api.Address(a.cfg.Address),
		api.Username(a.cfg.Username),
		api.Timeout(a.cfg.Timeout),
	}
	if a.cfg.Password != "" {
		opts = append(opts, api.Password(a.cfg.Password))
	}
	if a.cfg.Insecure {
		opts = append(opts, api.Insecure(true))
	}
	if a.cfg.SkipVerify {
		opts = append(opts, api.SkipVerify(true))
	}
	return api.NewTarget(opts...)
}

func (a *Adapter) connectAndProbe(ctx context.Context, tgt *target.Target) error {
	connectCtx, cancel := context.WithTimeout(ctx, a.cfg.Timeout)
	err := tgt.CreateGNMIClient(connectCtx)
	cancel()
	if err != nil {
		return fmt.Errorf("nxos: create gnmi client: %w", err)
	}

	capCtx, capCancel := context.WithTimeout(ctx, a.cfg.Timeout)
	capResp, err := tgt.Capabilities(capCtx)
	capCancel()
	if err != nil {
		// Capabilities failure is non-fatal: some lab setups disable
		// the RPC. Log and continue to Subscribe.
		a.logger.Warn("nxos: capabilities probe failed", "error", err)
		return nil
	}
	a.logger.Debug("nxos: capabilities",
		"gnmi_version", capResp.GetGNMIVersion(),
		"encodings", capResp.GetSupportedEncodings(),
		"model_count", len(capResp.GetSupportedModels()))
	return nil
}

// hostFromAddress returns the host portion of an address. Used as a
// preliminary Device.Name when /system/state/hostname has not yet been
// observed.
func hostFromAddress(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

// notificationCache keeps the most recently observed Update per
// absolute path. The cache is read in tick handlers as a single
// gnmi.Notification suitable for [Translate]; the deterministic key
// ordering produced by snapshot makes the resulting Sample
// reproducible for tests.
type notificationCache struct {
	mu       sync.Mutex
	updates  map[string]*gnmi.Update
	latestTS int64
}

func newNotificationCache() *notificationCache {
	return &notificationCache{updates: make(map[string]*gnmi.Update)}
}

func (c *notificationCache) ingest(n *gnmi.Notification) {
	if n == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if n.GetTimestamp() > c.latestTS {
		c.latestTS = n.GetTimestamp()
	}
	prefixElems := n.GetPrefix().GetElem()
	for _, u := range n.GetUpdate() {
		merged := mergePathElems(prefixElems, u.GetPath().GetElem())
		key := pathKey(merged)
		c.updates[key] = &gnmi.Update{
			Path: &gnmi.Path{Elem: merged},
			Val:  u.GetVal(),
		}
	}
}

func (c *notificationCache) snapshot() *gnmi.Notification {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.updates) == 0 {
		return nil
	}
	keys := make([]string, 0, len(c.updates))
	for k := range c.updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	updates := make([]*gnmi.Update, len(keys))
	for i, k := range keys {
		updates[i] = c.updates[k]
	}
	return &gnmi.Notification{Timestamp: c.latestTS, Update: updates}
}

func mergePathElems(prefix, path []*gnmi.PathElem) []*gnmi.PathElem {
	out := make([]*gnmi.PathElem, 0, len(prefix)+len(path))
	out = append(out, prefix...)
	out = append(out, path...)
	return out
}

func pathKey(elems []*gnmi.PathElem) string {
	var b strings.Builder
	for _, e := range elems {
		b.WriteByte('/')
		b.WriteString(e.GetName())
		if len(e.GetKey()) > 0 {
			keys := make([]string, 0, len(e.GetKey()))
			for k := range e.GetKey() {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				b.WriteByte('[')
				b.WriteString(k)
				b.WriteByte('=')
				b.WriteString(e.GetKey()[k])
				b.WriteByte(']')
			}
		}
	}
	return b.String()
}
