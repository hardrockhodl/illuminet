// Package exporter is the parent package for output exporters.
//
// Each supported output format lives in a subpackage and consumes the
// enriched sample stream produced by the core pipeline. Initial
// planned subpackages:
//
//   - otlp: OpenTelemetry Line Protocol over gRPC. This is the primary
//     supported output and the one the rest of the pipeline is designed
//     around.
//   - prom: Prometheus remote write, for environments that already
//     standardize on Prometheus as the time series database.
//   - influx: InfluxDB Line Protocol emitted on stdout. This output
//     exists for backward compatibility with Telegraf's inputs.exec
//     plugin, which is how paregupt's nexus_traffic_monitor integrates
//     with InfluxDB today.
//
// No exporter is implemented in this initial scaffold.
package exporter
