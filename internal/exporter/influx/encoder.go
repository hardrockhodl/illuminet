package influx

import (
	"bytes"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Encoder builds InfluxDB Line Protocol records and writes them to an
// io.Writer.
//
// The Encoder is stateful: a record begins with BeginLine, accumulates
// tags and fields, and is committed by EndLine. A record with no fields
// is silently dropped, since Line Protocol requires at least one
// field. The Encoder is not safe for concurrent use.
type Encoder struct {
	w           io.Writer
	inLine      bool
	measurement string
	tags        []kv
	fields      []kv
}

type kv struct {
	key, value string
}

// NewEncoder returns an Encoder that writes Line Protocol records to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// BeginLine starts a new record for the given measurement. Any
// in-progress state is discarded. The measurement name is written
// verbatim; the caller is responsible for restricting it to characters
// that do not require escaping (the constants in measurements.go are
// all safe).
func (e *Encoder) BeginLine(measurement string) *Encoder {
	e.inLine = true
	e.measurement = measurement
	e.tags = e.tags[:0]
	e.fields = e.fields[:0]
	return e
}

// AddTag adds a tag to the in-progress record. Empty values are
// silently skipped, per Line Protocol's convention. Tag keys and
// values are escaped for comma, equals and space.
func (e *Encoder) AddTag(key, value string) *Encoder {
	if !e.inLine || value == "" {
		return e
	}
	e.tags = append(e.tags, kv{escapeIdent(key), escapeIdent(value)})
	return e
}

// AddStringField adds a string field. Empty values are silently
// skipped to keep records compact; callers that need to record the
// fact a string is empty should add an explicit sentinel value.
func (e *Encoder) AddStringField(key, value string) *Encoder {
	if !e.inLine || value == "" {
		return e
	}
	e.fields = append(e.fields, kv{escapeIdent(key), quoteString(value)})
	return e
}

// AddUintField adds an unsigned integer field. The value is emitted
// with the 'i' (signed integer) suffix because the 'u' (unsigned
// integer) suffix is not universally supported across the Telegraf
// pipeline that this exporter targets. The cast is from uint64 to
// int64; values above MaxInt64 wrap, which is acceptable for the
// counters we emit but documented here for the record.
func (e *Encoder) AddUintField(key string, v uint64) *Encoder {
	if !e.inLine {
		return e
	}
	val := strconv.FormatInt(int64(v), 10) + "i"
	e.fields = append(e.fields, kv{escapeIdent(key), val})
	return e
}

// AddIntField adds a signed integer field.
func (e *Encoder) AddIntField(key string, v int64) *Encoder {
	if !e.inLine {
		return e
	}
	val := strconv.FormatInt(v, 10) + "i"
	e.fields = append(e.fields, kv{escapeIdent(key), val})
	return e
}

// AddFloatField adds a float64 field. The value is rendered with
// strconv.FormatFloat in 'g' mode with -1 precision, which produces
// the shortest representation that round-trips.
func (e *Encoder) AddFloatField(key string, v float64) *Encoder {
	if !e.inLine {
		return e
	}
	val := strconv.FormatFloat(v, 'g', -1, 64)
	e.fields = append(e.fields, kv{escapeIdent(key), val})
	return e
}

// AddBoolField adds a boolean field rendered as the single character
// "t" or "f".
func (e *Encoder) AddBoolField(key string, v bool) *Encoder {
	if !e.inLine {
		return e
	}
	val := "f"
	if v {
		val = "t"
	}
	e.fields = append(e.fields, kv{escapeIdent(key), val})
	return e
}

// AddOptionalUint is the nil-skipping variant of AddUintField. A nil
// pointer is silently ignored, which is how the exporter represents
// "not reported".
func (e *Encoder) AddOptionalUint(key string, v *uint64) *Encoder {
	if v == nil {
		return e
	}
	return e.AddUintField(key, *v)
}

// AddOptionalInt is the nil-skipping variant of AddIntField.
func (e *Encoder) AddOptionalInt(key string, v *int64) *Encoder {
	if v == nil {
		return e
	}
	return e.AddIntField(key, *v)
}

// AddOptionalFloat is the nil-skipping variant of AddFloatField.
func (e *Encoder) AddOptionalFloat(key string, v *float64) *Encoder {
	if v == nil {
		return e
	}
	return e.AddFloatField(key, *v)
}

// EndLine writes the in-progress record to the underlying writer with
// the given timestamp (encoded as nanoseconds since the Unix epoch).
// If no fields have been added, the record is silently dropped and
// nothing is written. The Encoder is reset for a new record either
// way.
func (e *Encoder) EndLine(timestamp time.Time) error {
	if !e.inLine {
		return nil
	}
	defer e.reset()

	if len(e.fields) == 0 {
		return nil
	}

	var buf bytes.Buffer
	buf.WriteString(e.measurement)

	sort.Slice(e.tags, func(i, j int) bool {
		return e.tags[i].key < e.tags[j].key
	})
	for _, t := range e.tags {
		buf.WriteByte(',')
		buf.WriteString(t.key)
		buf.WriteByte('=')
		buf.WriteString(t.value)
	}

	buf.WriteByte(' ')
	for i, f := range e.fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(f.key)
		buf.WriteByte('=')
		buf.WriteString(f.value)
	}

	buf.WriteByte(' ')
	buf.WriteString(strconv.FormatInt(timestamp.UnixNano(), 10))
	buf.WriteByte('\n')

	_, err := e.w.Write(buf.Bytes())
	return err
}

func (e *Encoder) reset() {
	e.inLine = false
	e.measurement = ""
	e.tags = e.tags[:0]
	e.fields = e.fields[:0]
}

// escapeIdent escapes a tag key, tag value or field key per Line
// Protocol's tag-component rules: backslash-escape comma, equals and
// space.
func escapeIdent(s string) string {
	if !strings.ContainsAny(s, ", =") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 4)
	for _, r := range s {
		switch r {
		case ',', '=', ' ':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// quoteString wraps and escapes a string field value per Line
// Protocol's string-field rules: double-quoted, with backslash-escaped
// double-quote and backslash characters.
func quoteString(s string) string {
	if !strings.ContainsAny(s, `"\`) {
		var b strings.Builder
		b.Grow(len(s) + 2)
		b.WriteByte('"')
		b.WriteString(s)
		b.WriteByte('"')
		return b.String()
	}
	var b strings.Builder
	b.Grow(len(s) + 4)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '\\':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}
