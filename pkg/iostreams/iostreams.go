package iostreams

import (
	"bytes"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// IOStreams holds the three standard streams and all computed output-mode flags.
// Every command receives an IOStreams via the Factory; no command reads env vars
// directly for output decisions.
type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	// Computed once at construction time from env + TTY state.
	IsInteractive  bool // false when CI=true, no TTY, or CIRCLECI_NO_INTERACTIVE
	ColorEnabled   bool // false when NO_COLOR, CIRCLECI_NO_COLOR, CLICOLOR=0, TERM=dumb
	SpinnerEnabled bool // false in CI mode or when stderr is not a TTY
	UpdatesEnabled bool // false in CI mode or CIRCLECI_NO_UPDATE_NOTIFIER
	Quiet          bool // suppress progress/informational output (--quiet/-q)
}

// System returns IOStreams wired to the real os.Std{in,out,err}, with all flags
// resolved from the current process environment and TTY state.
func System() *IOStreams {
	ios := &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	ios.resolve()
	return ios
}

// Test returns an IOStreams backed by in-memory buffers with color and
// interactivity disabled — suitable for unit tests.
// Returns (ios, stdin, stdout, stderr).
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &IOStreams{
		In:             io.NopCloser(in),
		Out:            out,
		ErrOut:         errOut,
		IsInteractive:  false,
		ColorEnabled:   false,
		SpinnerEnabled: false,
		UpdatesEnabled: false,
	}, in, out, errOut
}

// resolve computes all output-mode flags from the environment and TTY state.
func (s *IOStreams) resolve() {
	outFile, outIsFile := s.Out.(*os.File)
	errFile, errIsFile := s.ErrOut.(*os.File)

	stdoutIsTTY := outIsFile && isatty.IsTerminal(outFile.Fd())
	stderrIsTTY := errIsFile && isatty.IsTerminal(errFile.Fd())

	// CI mode when CI env var is set, stdout is not a TTY, or non-interactive is forced.
	ciMode := os.Getenv("CI") != "" ||
		!stdoutIsTTY ||
		os.Getenv("CIRCLECI_NO_INTERACTIVE") != ""

	s.IsInteractive = !ciMode

	// Color resolution follows no-color.org plus Heroku/CLICOLOR conventions.
	colorForced := os.Getenv("CLICOLOR_FORCE") == "1"
	noColor := os.Getenv("NO_COLOR") != "" ||
		os.Getenv("CIRCLECI_NO_COLOR") != "" ||
		os.Getenv("CLICOLOR") == "0" ||
		os.Getenv("TERM") == "dumb"

	switch {
	case colorForced:
		s.ColorEnabled = true
	case noColor:
		s.ColorEnabled = false
	default:
		s.ColorEnabled = stdoutIsTTY
	}

	// Spinner requires stderr TTY, non-CI mode, and no explicit disable.
	s.SpinnerEnabled = stderrIsTTY && !ciMode && os.Getenv("CIRCLECI_SPINNER_DISABLED") == ""

	// Update notifications disabled in CI mode or explicitly suppressed.
	s.UpdatesEnabled = !ciMode && os.Getenv("CIRCLECI_NO_UPDATE_NOTIFIER") == ""
}

// SetColorEnabled overrides the computed color setting (called by --no-color flag).
// Disabling color also disables the spinner since it relies on ANSI codes.
func (s *IOStreams) SetColorEnabled(enabled bool) {
	s.ColorEnabled = enabled
	if !enabled {
		s.SpinnerEnabled = false
	}
}

// SetInteractive overrides the interactive flag (called by --no-prompt flag).
func (s *IOStreams) SetInteractive(enabled bool) {
	s.IsInteractive = enabled
}
