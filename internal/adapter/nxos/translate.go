package nxos

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"

	"github.com/hardrockhodl/illuminet/internal/core/model"
)

// DeviceContext carries the device-level metadata that the adapter
// knows independently of the gNMI feed: operator-supplied tags and
// the connection target. It is supplied by the adapter to Translate so
// the returned Sample is fully populated even before any
// /system/state notification arrives.
type DeviceContext struct {
	// Name is the operator-supplied device name. The adapter sets this
	// from the gNMI target address before the first notification has
	// been seen; once a /system/state/hostname notification arrives,
	// Translate replaces it with the device's own hostname.
	Name string

	// ManagementIP is the address the collector uses to reach the
	// device, recorded for diagnostics.
	ManagementIP string

	// Location is an operator-supplied tag.
	Location string
}

// Translate converts a batch of gNMI Notifications into a Sample. The
// function is pure: no I/O, no time.Now, deterministic given the same
// inputs.
//
// Unknown paths are silently skipped. A TypedValue that does not match
// the field's expected scalar type returns an error, since that
// indicates either a server bug or a YANG model mismatch the adapter
// must surface.
func Translate(devCtx DeviceContext, notifications []*gnmi.Notification) (*model.Sample, error) {
	t := newTranslator(devCtx)

	for _, n := range notifications {
		if n == nil {
			continue
		}
		ts := n.GetTimestamp()
		if ts > t.latestTimestamp {
			t.latestTimestamp = ts
		}
		prefixElems := pathElems(n.GetPrefix())
		for _, u := range n.GetUpdate() {
			elems := append([]*gnmi.PathElem(nil), prefixElems...)
			elems = append(elems, pathElems(u.GetPath())...)
			if err := t.applyUpdate(elems, u.GetVal(), ts); err != nil {
				return nil, err
			}
		}
	}

	return t.finalize(), nil
}

type translator struct {
	devCtx          DeviceContext
	device          model.Device
	interfaces      map[string]*model.Interface
	latestTimestamp int64
}

func newTranslator(devCtx DeviceContext) *translator {
	return &translator{
		devCtx: devCtx,
		device: model.Device{
			Name:         devCtx.Name,
			ManagementIP: devCtx.ManagementIP,
			Location:     devCtx.Location,
			Vendor:       model.VendorCisco,
		},
		interfaces: make(map[string]*model.Interface),
	}
}

func (t *translator) finalize() *model.Sample {
	s := &model.Sample{Device: t.device}
	if t.latestTimestamp > 0 {
		s.CollectedAt = time.Unix(0, t.latestTimestamp)
	}

	names := make([]string, 0, len(t.interfaces))
	for k := range t.interfaces {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		s.Interfaces = append(s.Interfaces, *t.interfaces[k])
	}
	return s
}

func (t *translator) applyUpdate(elems []*gnmi.PathElem, val *gnmi.TypedValue, ts int64) error {
	if len(elems) == 0 {
		return nil
	}
	switch elems[0].GetName() {
	case "interfaces":
		return t.applyInterfaces(elems[1:], val, ts)
	case "qos":
		return t.applyQoS(elems[1:], val, ts)
	case "lldp":
		return t.applyLLDP(elems[1:], val, ts)
	case "system":
		return t.applySystem(elems[1:], val)
	default:
		return nil
	}
}

func (t *translator) applyInterfaces(elems []*gnmi.PathElem, val *gnmi.TypedValue, ts int64) error {
	if len(elems) < 2 || elems[0].GetName() != "interface" {
		return nil
	}
	name, ok := interfaceNameFromElem(elems[0])
	if !ok {
		return nil
	}
	iface := t.getOrCreateInterface(name, ts)
	return t.applyInterfaceField(iface, elems[1:], val)
}

func (t *translator) getOrCreateInterface(name string, ts int64) *model.Interface {
	if existing, ok := t.interfaces[name]; ok {
		if ts > 0 {
			notifyTime := time.Unix(0, ts)
			if existing.ObservedAt.IsZero() || notifyTime.After(existing.ObservedAt) {
				existing.ObservedAt = notifyTime
			}
		}
		return existing
	}
	iface := &model.Interface{Name: name}
	if ts > 0 {
		iface.ObservedAt = time.Unix(0, ts)
	}
	t.interfaces[name] = iface
	return iface
}

func (t *translator) applyInterfaceField(iface *model.Interface, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 || elems[0].GetName() != "state" {
		return nil
	}
	rest := elems[1:]
	if len(rest) == 0 {
		return nil
	}
	if rest[0].GetName() == "counters" {
		return t.applyCounter(iface, rest[1:], val)
	}
	switch rest[0].GetName() {
	case "admin-status":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("interface %s admin-status: %w", iface.Name, err)
		}
		iface.AdminState = parseAdminState(s)
	case "oper-status":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("interface %s oper-status: %w", iface.Name, err)
		}
		iface.OperState = parseOperState(s)
	case "description":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("interface %s description: %w", iface.Name, err)
		}
		iface.Description = s
	}
	return nil
}

func (t *translator) applyCounter(iface *model.Interface, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 {
		return nil
	}
	if iface.Counters == nil {
		iface.Counters = &model.InterfaceCounters{}
	}
	name := elems[0].GetName()

	if name == "last-clear" {
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("interface %s last-clear: %w", iface.Name, err)
		}
		parsed, perr := time.Parse(time.RFC3339Nano, s)
		if perr != nil {
			return fmt.Errorf("interface %s last-clear: parse %q: %w", iface.Name, s, perr)
		}
		iface.Counters.LastClear = &parsed
		return nil
	}

	v, err := uintVal(val)
	if err != nil {
		return fmt.Errorf("interface %s counters/%s: %w", iface.Name, name, err)
	}
	switch name {
	case "in-octets":
		iface.Counters.RxBytes = &v
	case "out-octets":
		iface.Counters.TxBytes = &v
	case "in-unicast-pkts":
		iface.Counters.RxUcastPkts = &v
	case "out-unicast-pkts":
		iface.Counters.TxUcastPkts = &v
	case "in-multicast-pkts":
		iface.Counters.RxMcastPkts = &v
	case "out-multicast-pkts":
		iface.Counters.TxMcastPkts = &v
	case "in-broadcast-pkts":
		iface.Counters.RxBcastPkts = &v
	case "out-broadcast-pkts":
		iface.Counters.TxBcastPkts = &v
	case "in-crc-errors":
		iface.Counters.RxCRC = &v
	}
	return nil
}

func (t *translator) applyQoS(elems []*gnmi.PathElem, val *gnmi.TypedValue, ts int64) error {
	if len(elems) == 0 || elems[0].GetName() != "interfaces" {
		return nil
	}
	return t.applyQoSInterfaces(elems[1:], val, ts)
}

func (t *translator) applyQoSInterfaces(elems []*gnmi.PathElem, val *gnmi.TypedValue, ts int64) error {
	if len(elems) == 0 || elems[0].GetName() != "interface" {
		return nil
	}
	name, ok := interfaceNameFromElem(elems[0])
	if !ok {
		return nil
	}
	iface := t.getOrCreateInterface(name, ts)
	return t.applyQoSInterfaceContent(iface, elems[1:], val)
}

func (t *translator) applyQoSInterfaceContent(iface *model.Interface, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 || elems[0].GetName() != "output" {
		return nil
	}
	return t.applyQoSOutput(iface, elems[1:], val)
}

func (t *translator) applyQoSOutput(iface *model.Interface, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 || elems[0].GetName() != "queues" {
		return nil
	}
	return t.applyQoSQueues(iface, elems[1:], val)
}

func (t *translator) applyQoSQueues(iface *model.Interface, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 || elems[0].GetName() != "queue" {
		return nil
	}
	name := elems[0].GetKey()["name"]
	if name == "" {
		return nil
	}
	queue := getOrCreateQueue(iface, name)
	return t.applyQueueContent(queue, elems[1:], val)
}

func (t *translator) applyQueueContent(queue *model.Queue, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 || elems[0].GetName() != "state" {
		return nil
	}
	return t.applyQueueState(queue, elems[1:], val)
}

func (t *translator) applyQueueState(queue *model.Queue, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 {
		return nil
	}
	name := elems[0].GetName()
	v, err := uintVal(val)
	if err != nil {
		return fmt.Errorf("queue %s %s: %w", queue.Name, name, err)
	}
	switch name {
	case "transmit-pkts":
		queue.Counters.TxPkts = &v
	case "transmit-octets":
		queue.Counters.TxBytes = &v
	case "dropped-pkts":
		queue.Counters.DropPkts = &v
	case "dropped-octets":
		queue.Counters.DropBytes = &v
	case "ecn-marked-pkts":
		if queue.ECN == nil {
			queue.ECN = &model.ECNCounter{}
		}
		queue.ECN.MarkedPkts = &v
	}
	return nil
}

func (t *translator) applyLLDP(elems []*gnmi.PathElem, val *gnmi.TypedValue, ts int64) error {
	if len(elems) == 0 || elems[0].GetName() != "interfaces" {
		return nil
	}
	return t.applyLLDPInterfaces(elems[1:], val, ts)
}

func (t *translator) applyLLDPInterfaces(elems []*gnmi.PathElem, val *gnmi.TypedValue, ts int64) error {
	if len(elems) == 0 || elems[0].GetName() != "interface" {
		return nil
	}
	name, ok := interfaceNameFromElem(elems[0])
	if !ok {
		return nil
	}
	iface := t.getOrCreateInterface(name, ts)
	return t.applyLLDPInterfaceContent(iface, elems[1:], val)
}

func (t *translator) applyLLDPInterfaceContent(iface *model.Interface, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 || elems[0].GetName() != "neighbors" {
		return nil
	}
	return t.applyLLDPNeighbors(iface, elems[1:], val)
}

func (t *translator) applyLLDPNeighbors(iface *model.Interface, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 || elems[0].GetName() != "neighbor" {
		return nil
	}
	if elems[0].GetKey()["id"] == "" {
		return nil
	}
	if iface.Peer == nil {
		iface.Peer = &model.Peer{LearnedVia: "lldp"}
	}
	return t.applyNeighborContent(iface.Peer, elems[1:], val)
}

func (t *translator) applyNeighborContent(peer *model.Peer, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 || elems[0].GetName() != "state" {
		return nil
	}
	return t.applyNeighborState(peer, elems[1:], val)
}

func (t *translator) applyNeighborState(peer *model.Peer, elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) == 0 {
		return nil
	}
	switch elems[0].GetName() {
	case "system-name":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("lldp system-name: %w", err)
		}
		peer.Name = s
	case "system-description":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("lldp system-description: %w", err)
		}
		peer.SystemDescription = s
	case "management-address":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("lldp management-address: %w", err)
		}
		peer.MgmtIP = s
	case "chassis-id":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("lldp chassis-id: %w", err)
		}
		// chassis-id only fills Peer.Interface when port-id has not
		// been seen yet; port-id is the preferred identifier.
		if peer.Interface == "" {
			peer.Interface = s
		}
	case "port-id":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("lldp port-id: %w", err)
		}
		peer.Interface = s
	case "system-capabilities":
		caps, err := stringListVal(val)
		if err != nil {
			return fmt.Errorf("lldp system-capabilities: %w", err)
		}
		peer.Capabilities = caps
	}
	return nil
}

// interfaceNameFromElem extracts the interface name from a PathElem's
// key map. OpenConfig uses "name" under /interfaces/interface but
// "interface-id" under /qos/interfaces/interface. The value is the
// same interface string on NX-OS; only the key name differs.
func interfaceNameFromElem(elem *gnmi.PathElem) (string, bool) {
	if name, ok := elem.GetKey()["name"]; ok && name != "" {
		return name, true
	}
	if id, ok := elem.GetKey()["interface-id"]; ok && id != "" {
		return id, true
	}
	return "", false
}

// getOrCreateQueue returns a pointer to the Queue with the given name
// in iface.Queues, appending a new zero-value Queue if no match
// exists. The returned pointer is only valid until the next call that
// might append to iface.Queues; Translate uses it within a single
// notification application so this is safe.
func getOrCreateQueue(iface *model.Interface, name string) *model.Queue {
	for i := range iface.Queues {
		if iface.Queues[i].Name == name {
			return &iface.Queues[i]
		}
	}
	iface.Queues = append(iface.Queues, model.Queue{Name: name})
	return &iface.Queues[len(iface.Queues)-1]
}

func (t *translator) applySystem(elems []*gnmi.PathElem, val *gnmi.TypedValue) error {
	if len(elems) < 2 || elems[0].GetName() != "state" {
		return nil
	}
	switch elems[1].GetName() {
	case "hostname":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("system hostname: %w", err)
		}
		t.device.Name = s
	case "software-version":
		s, err := stringVal(val)
		if err != nil {
			return fmt.Errorf("system software-version: %w", err)
		}
		t.device.OSVersion = s
	}
	return nil
}

func pathElems(p *gnmi.Path) []*gnmi.PathElem {
	if p == nil {
		return nil
	}
	return p.GetElem()
}

func parseAdminState(s string) model.AdminState {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UP":
		return model.AdminStateUp
	case "DOWN":
		return model.AdminStateDown
	default:
		return model.AdminStateUnknown
	}
}

func parseOperState(s string) model.OperState {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UP":
		return model.OperStateUp
	case "DOWN":
		return model.OperStateDown
	case "TESTING":
		return model.OperStateTesting
	default:
		return model.OperStateUnknown
	}
}

func uintVal(v *gnmi.TypedValue) (uint64, error) {
	if v == nil {
		return 0, errors.New("nil TypedValue")
	}
	switch x := v.GetValue().(type) {
	case *gnmi.TypedValue_UintVal:
		return x.UintVal, nil
	case *gnmi.TypedValue_IntVal:
		if x.IntVal < 0 {
			return 0, fmt.Errorf("expected non-negative integer, got %d", x.IntVal)
		}
		return uint64(x.IntVal), nil
	default:
		return 0, fmt.Errorf("expected integer TypedValue, got %T", x)
	}
}

func stringVal(v *gnmi.TypedValue) (string, error) {
	if v == nil {
		return "", errors.New("nil TypedValue")
	}
	switch x := v.GetValue().(type) {
	case *gnmi.TypedValue_StringVal:
		return x.StringVal, nil
	case *gnmi.TypedValue_AsciiVal:
		return x.AsciiVal, nil
	default:
		return "", fmt.Errorf("expected string TypedValue, got %T", x)
	}
}

// stringListVal extracts a list of strings from a gNMI leaflist
// TypedValue. Used for fields like LLDP system-capabilities that YANG
// models as a leaf-list.
func stringListVal(v *gnmi.TypedValue) ([]string, error) {
	if v == nil {
		return nil, errors.New("nil TypedValue")
	}
	switch x := v.GetValue().(type) {
	case *gnmi.TypedValue_LeaflistVal:
		elems := x.LeaflistVal.GetElement()
		out := make([]string, 0, len(elems))
		for _, el := range elems {
			ev, ok := el.GetValue().(*gnmi.TypedValue_StringVal)
			if !ok {
				return nil, fmt.Errorf("expected string element in leaflist, got %T", el.GetValue())
			}
			out = append(out, ev.StringVal)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected leaflist TypedValue, got %T", x)
	}
}
