// Package influx implements the InfluxDB Line Protocol exporter.
//
// The exporter consumes immutable [model.Sample] values and emits one
// Line Protocol record per logical entity (device, interface, queue,
// buffer instance, burst). Output is written to a caller-supplied
// [io.Writer]; the typical destination is os.Stdout when running
// under Telegraf's inputs.exec plugin.
//
// Field naming follows InfluxDB-conventional snake_case, which is a
// deliberate breaking change from paregupt/nexus_traffic_monitor (NTM)
// where the two differ. The migration cost is recovered by avoiding
// case-mixed identifiers downstream; the mapping is documented in the
// constants defined in measurements.go.
//
// The exporter is not safe for concurrent use. Callers that need to
// fan out from multiple goroutines should funnel writes through a
// single Exporter goroutine.
package influx
