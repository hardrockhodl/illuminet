# IllumiNET

Multi-vendor datacenter fabric observability.

## What it is

IllumiNET is a Go collector that pulls operational telemetry from datacenter
switches (gNMI first, with vendor-specific fallbacks), normalizes it into a
vendor-agnostic domain model, enriches it with fabric topology context, and
emits the result on the OpenTelemetry wire (with optional Prometheus and
InfluxDB outputs). The initial focus is RoCEv2 lossless fabrics, where the
useful signal lives in per-priority PFC pauses, ECN markings, microburst
peaks, and buffer occupancy rather than in aggregate interface counters.

## Status

Early development. The collector runs end-to-end against a Cisco NX-OS
gNMI target and emits InfluxDB Line Protocol on stdout, suitable for
ingestion by Telegraf's `inputs.exec` plugin. Coverage so far:

- NX-OS adapter: device metadata, interface counters (RxBytes, TxBytes,
  unicast/multicast/broadcast packet counts, CRC errors), admin/oper
  state, descriptions.
- Fake adapter for testing without a switch.
- InfluxDB Line Protocol exporter.

Not yet implemented: queue counters, PFC, ECN, buffer state, microburst
detection, LLDP topology enrichment, additional vendors. APIs, the
internal data model, and output formats will change without notice
until the v0.1.0 release.

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the planned data flow,
domain model, and output strategy.

## Inspiration

The project takes operational inspiration from
[paregupt/nexus_traffic_monitor](https://github.com/paregupt/nexus_traffic_monitor)
and its sibling projects `ucs_traffic_monitor` and `mds_traffic_monitor`, which
established the per-port and per-queue dashboards this work aims to generalize
across vendors.

## License

Apache License 2.0. See [LICENSE](LICENSE).
