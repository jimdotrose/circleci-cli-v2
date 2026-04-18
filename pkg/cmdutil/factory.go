package cmdutil

import (
	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

// Factory provides shared dependencies to every command constructor.
// Commands receive a *Factory rather than accessing globals, which makes
// test injection straightforward: swap IOStreams for a test buffer,
// APIClient for a mock, etc.
type Factory struct {
	IOStreams *iostreams.IOStreams

	// Config returns the loaded CLI configuration. Lazily evaluated so
	// commands that don't need config don't pay the file-read cost.
	Config func() (Config, error)

	// APIClient returns an authenticated CircleCI API client.
	// Populated in Sprint 2 when the apiclient package exists.
	// APIClient func() (*apiclient.Client, error)

	// BaseURL returns the configured CircleCI host URL.
	BaseURL func() string
}

// Config is the minimal interface for CLI configuration needed by Sprint 1.
// A full implementation lives in pkg/config (Sprint 2).
type Config interface {
	// Token returns the active API token.
	Token() string
	// Host returns the configured CircleCI host URL.
	Host() string
}

// New builds a Factory wired to real system streams and a default base URL.
// Call this from main(); tests build their own factory with mock streams.
func New() *Factory {
	f := &Factory{
		IOStreams: iostreams.System(),
	}
	f.BaseURL = func() string {
		return "https://circleci.com"
	}
	return f
}
