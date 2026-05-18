// Package pipeline orchestrates the IllumiNET collection loop.
//
// A [Pipeline] fans Sample data out from one or more Adapters to one
// or more Exporters. Adapters run in their own goroutines and push
// onto a buffered channel; a central dispatcher reads the channel and
// invokes Export on each configured exporter sequentially. A failing
// exporter logs and is skipped; it does not block siblings.
//
// On context cancellation the pipeline waits up to a few seconds for
// adapters to drain, then closes its exporters. The exit ordering is:
//
//  1. Context cancellation propagates to all adapters.
//  2. Adapter goroutines return; their Run errors are collected.
//  3. The sample channel is closed and the dispatcher drains.
//  4. Each exporter's Close is invoked.
//
// Future iterations will add enrichment stages (LLDP-driven peer
// classification, port-role inference) between the channel reader and
// the exporter dispatch loop. The current shape is intentionally
// minimal so the contract between adapter, pipeline and exporter is
// easy to reason about.
package pipeline
