package cmdutil

import (
	"os"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/config"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

// Factory provides shared dependencies to every command constructor.
// Commands receive a *Factory rather than accessing globals, enabling
// straightforward test injection: swap IOStreams for a test buffer,
// Config for a MockConfig, etc.
type Factory struct {
	IOStreams *iostreams.IOStreams
	Debug    bool // set by --debug flag; enables HTTP request/response logging

	// Config returns the loaded CLI configuration. Lazily evaluated so
	// commands that don't need config don't pay the file-read cost.
	Config func() (config.Config, error)

	// BaseURL returns the active CircleCI host URL, respecting:
	//   --host flag > CIRCLECI_HOST env > config file > default
	BaseURL func() string

	// APIClient returns a configured REST client for the CircleCI API v2.
	APIClient func() (*apiclient.Client, error)
}

// New builds a Factory wired to real system streams and a default base URL.
// Call this from main(); tests build their own factory with mock streams and config.
func New() *Factory {
	f := &Factory{
		IOStreams: iostreams.System(),
	}

	var cachedCfg config.Config
	f.Config = func() (config.Config, error) {
		if cachedCfg != nil {
			return cachedCfg, nil
		}
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			// Return defaults so the CLI remains usable with a broken config file.
			cachedCfg = config.NewMockConfig()
			return cachedCfg, err
		}
		cachedCfg = cfg
		return cfg, nil
	}

	f.BaseURL = func() string {
		if h := os.Getenv("CIRCLECI_HOST"); h != "" {
			return h
		}
		cfg, err := f.Config()
		if err != nil {
			return "https://circleci.com"
		}
		return cfg.Host()
	}

	f.APIClient = func() (*apiclient.Client, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		token := cfg.Token()
		if token == "" {
			return nil, cierrors.ErrAuthRequired
		}
		return apiclient.NewWithDebug(f.BaseURL(), token, f.Debug, f.IOStreams.ErrOut), nil
	}

	return f
}
