// Package version exposes build/version information for tldg.
package version

// tldg-5xh
var (
	// Version is the semantic version, overridable via -ldflags.
	Version = "0.0.0-dev"
	// Commit is the git commit hash, overridable via -ldflags.
	Commit = "unknown"
	// Date is the build date, overridable via -ldflags.
	Date = "unknown"
)

// String returns a human-readable version string.
func String() string {
	return Version + " (" + Commit + ", " + Date + ")"
}
