package settings

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdList returns the `circleci settings list` command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all CLI settings",
		Long: heredoc.Doc(`
			List all CLI configuration settings and their current effective values.

			Values reflect the full precedence chain:
			  CIRCLECI_* env vars > ~/.circleci/cli.yml > built-in defaults

			Available settings:
			  host          CircleCI host URL (default: https://circleci.com)
			  token         API token — manage with 'circleci auth login'
			  update_check  Enable version update notifications (default: true)
			  telemetry     Enable anonymous usage telemetry (default: true)
		`),
		Example: heredoc.Doc(`
			# List all settings:
			$ circleci settings list

			# Check which host the CLI is pointed at:
			$ circleci settings list | grep host

			# Use in a shell script:
			$ circleci settings list
			host         https://circleci.com
			token        [set]
			update_check true
			telemetry    true
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			for _, key := range cfg.Keys() {
				val, _ := cfg.Get(key)
				display := val
				if key == "token" {
					if val == "" {
						display = "[not set]"
					} else {
						display = "[set]"
					}
				}
				fmt.Fprintf(f.IOStreams.Out, "%-14s %s\n", key, display)
			}
			return nil
		},
	}
}
