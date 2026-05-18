// Package model defines the normalized, vendor-agnostic domain model
// used throughout IllumiNET.
//
// The model is intentionally a "rich union": it carries every counter
// and attribute that any supported vendor can plausibly expose, even
// when only one platform reports a given field today. Adapters in
// internal/adapter populate the fields they can derive and leave the
// rest at the zero value. This is what separates IllumiNET from a
// generic Telegraf/gNMI configuration.
//
// Conventions:
//
//   - Optional numeric fields (counters, gauges, percentages) are
//     pointer types. A nil pointer means "not reported by this vendor
//     on this platform"; that is semantically distinct from a reported
//     zero. Consumers MUST treat nil and zero differently.
//   - Optional string fields are plain strings; an empty string means
//     "not reported". Pointer-to-string is only used when an empty
//     string itself is a meaningful value, which is not the case in
//     this package today.
//   - Timestamps are time.Time. A zero time.Time (IsZero) means "not
//     reported".
//   - Samples are immutable snapshots. The model exposes no Update or
//     delta methods; the pipeline computes deltas across successive
//     samples, never inside a sample.
//   - Interface and platform identifiers are kept in their original
//     vendor-specific form. The model does not attempt to canonicalize
//     "Ethernet1/1" against "ge-0/0/0"; downstream enrichment may.
//
// This package depends only on the Go standard library.
package model
