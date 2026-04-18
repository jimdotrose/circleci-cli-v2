package telemetry

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const noticePath = ".circleci/.telemetry_notice_shown"

// ShowNoticeIfNeeded prints the first-run disclosure once, then records that
// it has been shown. No-ops when CI=true or any opt-out env var is set.
func ShowNoticeIfNeeded(w io.Writer, homeDir string) {
	// Never show in automated environments.
	if os.Getenv("CI") != "" ||
		os.Getenv("CIRCLECI_NO_TELEMETRY") != "" ||
		os.Getenv("NO_ANALYTICS") != "" ||
		os.Getenv("DO_NOT_TRACK") != "" {
		return
	}

	flagFile := filepath.Join(homeDir, noticePath)
	if _, err := os.Stat(flagFile); err == nil {
		return // already shown
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "CircleCI collects anonymous usage data to improve the CLI.")
	fmt.Fprintln(w, "No personal information, tokens, or file paths are collected.")
	fmt.Fprintln(w, "To opt out: circleci telemetry disable")
	fmt.Fprintln(w, "")

	// Record that the notice has been shown.
	_ = os.MkdirAll(filepath.Dir(flagFile), 0700)
	_ = os.WriteFile(flagFile, []byte("shown\n"), 0600)
}
