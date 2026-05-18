// Command illuminet is the entry point for the IllumiNET collector binary.
//
// The first iteration is intentionally a thin skeleton. It accepts two
// subcommands, "version" and "collect", and a small set of global flags
// for verbosity. No telemetry collection is performed yet.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hardrockhodl/illuminet/pkg/version"
)

const usage = `illuminet is a multi-vendor datacenter fabric observability collector.

Usage:
  illuminet [global flags] <command> [command flags]

Commands:
  version   Print build identification (version, commit, build date).
  collect   Start the collection pipeline (not yet implemented).

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
		// flag already wrote the error and usage to stderr.
		return 2
	}

	// Resolve the verbosity level. The verbosity is wired into a logger
	// in a later iteration; for now we only retain the resolved level.
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

// resolveVerbosity returns the effective verbosity level. Higher
// numbers indicate more verbose output. The highest flag wins.
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

// runVersion prints build identification produced by the linker.
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

// runCollect is a placeholder for the collection pipeline command. It
// reports that the functionality is not yet available and exits non
// zero so scripts notice.
func runCollect(args []string, _, stderr io.Writer) int {
	fs := flag.NewFlagSet("illuminet collect", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	fmt.Fprintln(stderr, "illuminet collect: not yet implemented")
	return 1
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
