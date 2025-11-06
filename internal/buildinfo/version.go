package buildinfo

var (
	Version string
	Commit  string
)

// FormattedVersion returns the version with commit hash if available
func FormattedVersion() string {
	version := Version
	if Commit != "" && len(Commit) >= 8 {
		version += "-" + Commit[:8]
	}
	return version
}
