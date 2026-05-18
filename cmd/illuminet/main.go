// Command illuminet is the entry point for the IllumiNET collector binary.
//
// The binary accepts two subcommands, "version" and "collect", plus a
// small set of global flags. The "collect" subcommand runs the
// configured adapter through the pipeline and writes the configured
// exporter's output to stdout.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hardrockhodl/illuminet/internal/adapter"
	"github.com/hardrockhodl/illuminet/internal/adapter/fake"
	"github.com/hardrockhodl/illuminet/internal/core/pipeline"
	"github.com/hardrockhodl/illuminet/internal/exporter"
	"github.com/hardrockhodl/illuminet/internal/exporter/influx"
	"github.com/hardrockhodl/illuminet/pkg/version"
)

const usage = `illuminet is a multi-vendor datacenter fabric observability collector.

Usage:
  illuminet [global flags] <command> [command flags]

Commands:
  version   Print build identification (version, commit, build date).
  collect   Start the collection pipeline.

Global flags:
  -v        Verbose output (info level).
  -vv       More verbose output (debug level).
  -vvv      Most verbose output (trace level).
  -h, --help
            Show this help message.
`

// run is the testable entry point. It returns the process exit code so
// main stays a one-liner.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("illuminet", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var v, vv, vvv bool
	fs.BoolVar(&v, "v", false, "verbose output (info level)")
	fs.BoolVar(&vv, "vv", false, "more verbose output (debug level)")
	fs.BoolVar(&vvv, "vvv", false, "most verbose output (trace level)")

	fs.Usage = func() {
		fmt.Fprint(stderr, usage)
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	level := resolveVerbosity(v, vv, vvv)
	_ = level

	rest := fs.Args()
	if len(rest) == 0 {
		fs.Usage()
		return 2
	}

	cmd, cmdArgs := rest[0], rest[1:]
	switch cmd {
	case "version":
		return runVersion(cmdArgs, stdout, stderr)
	case "collect":
		return runCollect(cmdArgs, stdout, stderr)
	case "help", "-h", "--help":
		fs.Usage()
		return 0
	default:
		fmt.Fprintf(stderr, "illuminet: unknown command %q\n\n", cmd)
		fs.Usage()
		return 2
	}
}

func resolveVerbosity(v, vv, vvv bool) int {
	switch {
	case vvv:
		return 3
	case vv:
		return 2
	case v:
		return 1
	default:
		return 0
	}
}

func runVersion(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("illuminet version", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	fmt.Fprintf(stdout, "illuminet %s (commit %s, built %s)\n",
		version.Version, version.Commit, version.Date)
	return 0
}

func runCollect(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("illuminet collect", flag.ContinueOnError)
	fs.SetOutput(stderr)

	adapterName := fs.String("adapter", "fake", `adapter to start (currently only "fake")`)
	interval := fs.Duration("interval", 5*time.Second, "poll interval for polling adapters")
	exporterName := fs.String("exporter", "stdout", `exporter to use (stdout = InfluxDB Line Protocol)`)
	logLevel := fs.String("log-level", "info", "slog level: debug, info, warn, error")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	level, err := parseLogLevel(*logLevel)
	if err != nil {
		fmt.Fprintf(stderr, "illuminet collect: %v\n", err)
		return 2
	}
	logger := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level}))

	src, code := buildAdapter(*adapterName, *interval, logger, stderr)
	if code != 0 {
		return code
	}
	sink, code := buildExporter(*exporterName, stdout, stderr)
	if code != 0 {
		return code
	}

	p, err := pipeline.New(pipeline.Options{
		Adapters:  []adapter.Adapter{src},
		Exporters: []exporter.Exporter{sink},
		Logger:    logger,
	})
	if err != nil {
		fmt.Fprintf(stderr, "illuminet collect: %v\n", err)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("starting collector",
		"adapter", *adapterName,
		"exporter", *exporterName,
		"interval", *interval)

	if err := p.Run(ctx); err != nil {
		fmt.Fprintf(stderr, "illuminet collect: %v\n", err)
		return 1
	}
	return 0
}

func buildAdapter(name string, interval time.Duration, logger *slog.Logger, stderr io.Writer) (adapter.Adapter, int) {
	switch name {
	case "fake":
		return fake.New(interval, logger), 0
	default:
		fmt.Fprintf(stderr, "illuminet collect: unknown adapter %q\n", name)
		return nil, 2
	}
}

func buildExporter(name string, stdout, stderr io.Writer) (exporter.Exporter, int) {
	switch name {
	case "stdout":
		return influx.New(stdout), 0
	default:
		fmt.Fprintf(stderr, "illuminet collect: unknown exporter %q\n", name)
		return nil, 2
	}
}

func parseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown log level %q (want debug|info|warn|error)", s)
	}
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
