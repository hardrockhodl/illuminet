package influx

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func endLine(t *testing.T, enc *Encoder, ts time.Time) {
	t.Helper()
	if err := enc.EndLine(ts); err != nil {
		t.Fatalf("EndLine: %v", err)
	}
}

func TestEncoder_NoFieldsDropsLine(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.BeginLine("test").AddTag("host", "h1")
	endLine(t, enc, time.Unix(0, 1234567890))
	if buf.Len() != 0 {
		t.Errorf("buffer non-empty: %q", buf.String())
	}
}

func TestEncoder_TagEscaping(t *testing.T) {
	cases := []struct {
		name     string
		key, val string
		want     string
	}{
		{"comma in value", "k", "a,b", `k=a\,b`},
		{"equals in value", "k", "a=b", `k=a\=b`},
		{"space in value", "k", "a b", `k=a\ b`},
		{"all three in value", "k", "a, b=c", `k=a\,\ b\=c`},
		{"comma in key", "k,1", "v", `k\,1=v`},
		{"clean", "k", "v", `k=v`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.BeginLine("m").AddTag(tc.key, tc.val).AddUintField("f", 1)
			endLine(t, enc, time.Unix(0, 1))
			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("output %q missing %q", buf.String(), tc.want)
			}
		})
	}
}

func TestEncoder_EmptyTagSkipped(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.BeginLine("m").
		AddTag("kept", "v").
		AddTag("skipped", "").
		AddUintField("f", 1)
	endLine(t, enc, time.Unix(0, 1))
	out := buf.String()
	if !strings.Contains(out, "kept=v") {
		t.Errorf("missing kept tag in %q", out)
	}
	if strings.Contains(out, "skipped") {
		t.Errorf("empty tag leaked into output: %q", out)
	}
}

func TestEncoder_TagOrderLexicographic(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.BeginLine("m").
		AddTag("zulu", "z").
		AddTag("alpha", "a").
		AddTag("mike", "m").
		AddUintField("f", 1)
	endLine(t, enc, time.Unix(0, 0))
	want := "m,alpha=a,mike=m,zulu=z f=1i 0\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestEncoder_StringFieldEscaping(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{"clean", "hello", `f="hello"`},
		{"double quote", `say "hi"`, `f="say \"hi\""`},
		{"backslash", `a\b`, `f="a\\b"`},
		{"both", `\"`, `f="\\\""`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.BeginLine("m").AddStringField("f", tc.value)
			endLine(t, enc, time.Unix(0, 1))
			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("output %q missing %q", buf.String(), tc.want)
			}
		})
	}
}

func TestEncoder_EmptyStringFieldSkipped(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.BeginLine("m").
		AddStringField("empty", "").
		AddUintField("present", 1)
	endLine(t, enc, time.Unix(0, 1))
	if strings.Contains(buf.String(), "empty") {
		t.Errorf("empty string field leaked: %q", buf.String())
	}
}

func TestEncoder_OptionalNilSkipped(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.BeginLine("m").
		AddOptionalUint("nilu", nil).
		AddOptionalInt("nili", nil).
		AddOptionalFloat("nilf", nil).
		AddUintField("present", 7)
	endLine(t, enc, time.Unix(0, 1))
	out := buf.String()
	if strings.Contains(out, "nilu") || strings.Contains(out, "nili") || strings.Contains(out, "nilf") {
		t.Errorf("nil optional leaked into output: %q", out)
	}
	if !strings.Contains(out, "present=7i") {
		t.Errorf("missing present field in %q", out)
	}
}

func TestEncoder_IntFieldHasISuffix(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.BeginLine("m").AddUintField("u", 42).AddIntField("s", -7)
	endLine(t, enc, time.Unix(0, 1))
	out := buf.String()
	if !strings.Contains(out, "u=42i") {
		t.Errorf("missing u=42i in %q", out)
	}
	if !strings.Contains(out, "s=-7i") {
		t.Errorf("missing s=-7i in %q", out)
	}
}

func TestEncoder_FloatFieldHasNoSuffix(t *testing.T) {
	cases := []struct {
		name string
		v    float64
		want string
	}{
		{"integral", 1.0, "f=1"},
		{"fraction", 3.14, "f=3.14"},
		{"small", 0.001, "f=0.001"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.BeginLine("m").AddFloatField("f", tc.v)
			endLine(t, enc, time.Unix(0, 1))
			if !strings.Contains(buf.String(), tc.want) {
				t.Errorf("output %q missing %q", buf.String(), tc.want)
			}
		})
	}
}

func TestEncoder_TimestampNanoseconds(t *testing.T) {
	var buf bytes.Buffer
	ts := time.Unix(1700000000, 123456789)
	enc := NewEncoder(&buf)
	enc.BeginLine("m").AddUintField("f", 1)
	endLine(t, enc, ts)
	out := buf.String()
	if !strings.HasSuffix(out, " 1700000000123456789\n") {
		t.Errorf("output %q does not end with expected nano timestamp", out)
	}
}

func TestEncoder_MultipleLines(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.BeginLine("m1").AddUintField("f", 1)
	endLine(t, enc, time.Unix(0, 1))
	enc.BeginLine("m2").AddUintField("f", 2)
	endLine(t, enc, time.Unix(0, 2))
	got := buf.String()
	want := "m1 f=1i 1\nm2 f=2i 2\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEncoder_BoolField(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.BeginLine("m").AddBoolField("up", true).AddBoolField("down", false)
	endLine(t, enc, time.Unix(0, 1))
	out := buf.String()
	if !strings.Contains(out, "up=t") || !strings.Contains(out, "down=f") {
		t.Errorf("bool encoding wrong in %q", out)
	}
}
