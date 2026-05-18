// Package exporter is the parent package for output exporters.
//
// Each supported output format lives in a subpackage and consumes the
// enriched sample stream produced by the core pipeline. The
// [Exporter] interface in this package is the common contract every
// subpackage implements.
//
// Implemented subpackages:
//
//   - influx: InfluxDB Line Protocol emitted to an io.Writer. This
//     output exists for backward compatibility with Telegraf's
//     inputs.exec plugin, which is how paregupt's nexus_traffic_monitor
//     integrates with InfluxDB today.
//
// Planned subpackages:
//
//   - otlp: OpenTelemetry Line Protocol over gRPC. The primary intended
//     output and the one the rest of the pipeline is designed around.
//   - prom: Prometheus remote write, for environments that already
//     standardize on Prometheus as the time series database.
package exporter
