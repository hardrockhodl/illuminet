package nxos

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	defaultInterval = 10 * time.Second
	defaultTimeout  = 10 * time.Second
)

// Config holds the connection parameters for a single Cisco NX-OS
// gNMI target.
type Config struct {
	// Address is the gNMI server address in host:port form. NX-OS
	// listens on TCP/50051 by default but the operator can customize.
	// Required.
	Address string

	// Username for gNMI authentication. Required.
	Username string

	// Password for gNMI authentication. Required unless Insecure mode
	// is in use against a lab target with no authentication.
	Password string

	// Insecure disables TLS. Default false. Set true only for lab use
	// where the NX-OS gNMI server has been configured without TLS.
	Insecure bool

	// SkipVerify disables TLS certificate verification. Lab-only.
	SkipVerify bool

	// Interval is the sample interval for STREAM SAMPLE subscriptions.
	// Zero is replaced by 10 seconds via applyDefaults.
	Interval time.Duration

	// Timeout is the per-operation gRPC timeout (Capabilities, initial
	// Subscribe handshake). Subscription streaming itself runs without
	// timeout. Zero is replaced by 10 seconds via applyDefaults.
	Timeout time.Duration

	// Location is an operator-supplied tag that propagates into
	// Device.Location on every emitted Sample. Optional.
	Location string
}

// Validate reports the first configuration error encountered. The
// receiver is not modified; callers that want defaults applied must
// call applyDefaults explicitly.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		return errors.New("nxos config: Address is required")
	}
	if !strings.Contains(c.Address, ":") {
		return fmt.Errorf("nxos config: Address %q must include a port (host:port)", c.Address)
	}
	if strings.TrimSpace(c.Username) == "" {
		return errors.New("nxos config: Username is required")
	}
	if !c.Insecure && strings.TrimSpace(c.Password) == "" {
		return errors.New("nxos config: Password is required (use Insecure=true only for lab targets without auth)")
	}
	if c.Interval < 0 {
		return fmt.Errorf("nxos config: Interval must be non-negative, got %s", c.Interval)
	}
	if c.Timeout < 0 {
		return fmt.Errorf("nxos config: Timeout must be non-negative, got %s", c.Timeout)
	}
	return nil
}

// applyDefaults fills zero-value fields with package defaults.
func (c *Config) applyDefaults() {
	if c.Interval == 0 {
		c.Interval = defaultInterval
	}
	if c.Timeout == 0 {
		c.Timeout = defaultTimeout
	}
}
