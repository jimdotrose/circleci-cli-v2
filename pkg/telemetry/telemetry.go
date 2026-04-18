package telemetry

import "os"

// Enabled reports whether telemetry collection is active.
// Telemetry is disabled when any of the following are set:
//   - CIRCLECI_NO_TELEMETRY
//   - NO_ANALYTICS
//   - DO_NOT_TRACK
//   - CI=true (automated environments)
//
// When the config value is "false", telemetry is also disabled.
func Enabled(configValue string) bool {
	if os.Getenv("CIRCLECI_NO_TELEMETRY") != "" {
		return false
	}
	if os.Getenv("NO_ANALYTICS") != "" {
		return false
	}
	if os.Getenv("DO_NOT_TRACK") != "" {
		return false
	}
	if os.Getenv("CI") != "" {
		return false
	}
	return configValue != "false"
}
