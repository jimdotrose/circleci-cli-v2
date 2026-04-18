package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/root"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/version"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// These are set at build time via -ldflags.
var (
	buildVersion = "dev"
	buildDate    = "unknown"
)

func main() {
	// Silence broken pipe errors when output is piped to head, grep -m 1, etc.
	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)

	code := run()
	if code != cierrors.ExitSuccess {
		os.Exit(code)
	}
}

func run() int {
	f := cmdutil.New()
	ios := f.IOStreams

	rootCmd := root.NewCmdRoot(f)

	// Attach --version / -V flag at root level.
	rootCmd.Version = buildVersion
	rootCmd.InitDefaultVersionFlag()

	// Also wire `circleci version` as a subcommand for discoverability.
	rootCmd.AddCommand(version.NewCmdVersion(f, buildVersion))

	// Cobra's built-in completion command (bash/zsh/fish/powershell).
	rootCmd.InitDefaultCompletionCmd()

	if err := rootCmd.Execute(); err != nil {
		// CLIErrors already have user-facing formatting; plain errors get a
		// generic prefix.
		var cliErr *cierrors.CLIError
		if errors.As(err, &cliErr) {
			fmt.Fprintln(ios.ErrOut, cliErr.Error())
		} else {
			fmt.Fprintf(ios.ErrOut, "Error: %s\n", err)
		}
		return cierrors.GetExitCode(err)
	}

	return cierrors.ExitSuccess
}
