# IllumiNET

Multi-vendor datacenter fabric observability.

> **Status: early development.** APIs, the internal data model, and output
> formats will change without notice until the v0.1.0 release.

## What it is

IllumiNET is a Go collector that pulls operational telemetry from datacenter
switches (gNMI first, with vendor-specific fallbacks), normalizes it into a
vendor-agnostic domain model, enriches it with fabric topology context, and
emits the result on the OpenTelemetry wire (with optional Prometheus and
InfluxDB outputs). The initial focus is RoCEv2 lossless fabrics, where the
useful signal lives in per-priority PFC pauses, ECN markings, microburst
peaks, and buffer occupancy rather than in aggregate interface counters.

## What it isn't yet

Nothing collects anything yet. This repository currently contains only the
project scaffold: a buildable `cmd/illuminet` skeleton with `version` and
`collect` placeholders, package layout for the core pipeline, vendor adapters,
and exporters, plus CI and tooling. The `collect` command exits non-zero with
"not yet implemented".

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
