package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/root"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// Set at build time via -ldflags.
var (
	buildVersion = "dev"
	buildDate    = "unknown"
)

func main() {
	// Silence broken pipe errors when output is piped to head, grep -m 1, etc.
	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)

	if code := run(); code != cierrors.ExitSuccess {
		os.Exit(code)
	}
}

func run() int {
	f := cmdutil.New()
	rootCmd := root.NewCmdRoot(f, buildVersion)

	if err := rootCmd.Execute(); err != nil {
		var cliErr *cierrors.CLIError
		if errors.As(err, &cliErr) {
			fmt.Fprintln(f.IOStreams.ErrOut, cliErr.Error())
		} else {
			fmt.Fprintf(f.IOStreams.ErrOut, "Error: %s\n", err)
		}
		return cierrors.GetExitCode(err)
	}
	return cierrors.ExitSuccess
}
