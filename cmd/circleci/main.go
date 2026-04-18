package main

import (
	"encoding/json"
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

	// SIGINT (Ctrl+C): exit with ExitCancelled (6) instead of Go's default 2.
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	go func() {
		<-sigint
		os.Exit(cierrors.ExitCancelled)
	}()

	if code := run(); code != cierrors.ExitSuccess {
		os.Exit(code)
	}
}

func run() int {
	f := cmdutil.New()
	rootCmd := root.NewCmdRoot(f, buildVersion)

	if err := rootCmd.Execute(); err != nil {
		exitCode := cierrors.GetExitCode(err)
		if f.JSONOutput {
			var cliErr *cierrors.CLIError
			type jsonErr struct {
				Code        string   `json:"code,omitempty"`
				Title       string   `json:"title"`
				Message     string   `json:"message"`
				Suggestions []string `json:"suggestions,omitempty"`
				ExitCode    int      `json:"exit_code"`
			}
			var je jsonErr
			if errors.As(err, &cliErr) {
				je = jsonErr{
					Code:        cliErr.Code,
					Title:       cliErr.Title,
					Message:     cliErr.Message,
					Suggestions: cliErr.Suggestions,
					ExitCode:    cliErr.ExitCode,
				}
			} else {
				je = jsonErr{Title: "Error", Message: err.Error(), ExitCode: exitCode}
			}
			if b, merr := json.Marshal(map[string]interface{}{"error": je}); merr == nil {
				fmt.Fprintln(f.IOStreams.ErrOut, string(b))
			}
		} else {
			var cliErr *cierrors.CLIError
			if errors.As(err, &cliErr) {
				fmt.Fprintln(f.IOStreams.ErrOut, cliErr.Error())
			} else {
				fmt.Fprintf(f.IOStreams.ErrOut, "Error: %s\n", err)
			}
		}
		return exitCode
	}
	return cierrors.ExitSuccess
}
