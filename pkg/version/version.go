// Package version exposes build-time identification metadata for IllumiNET.
//
// The Version, Commit, and Date variables are intended to be set at link
// time using the Go linker's -ldflags option, for example:
//
//	go build -ldflags "-X github.com/hardrockhodl/illuminet/pkg/version.Version=v0.0.1 \
//	                   -X github.com/hardrockhodl/illuminet/pkg/version.Commit=$(git rev-parse --short HEAD) \
//	                   -X github.com/hardrockhodl/illuminet/pkg/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
//
// When the binary is built without these flags (for example during
// development or via "go run"), the variables fall back to placeholder
// values.
package version

// Build metadata. These variables are overwritten via -ldflags at build
// time. Defaults indicate a non-release build.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
