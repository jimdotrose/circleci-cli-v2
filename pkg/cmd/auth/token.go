package auth

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdToken returns the `circleci auth token` command.
func NewCmdToken(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print the active API token",
		Long: heredoc.Doc(`
			Print the active CircleCI API token to stdout.

			The output contains no trailing newline, making it safe for use in
			command substitution. Returns exit code 3 if no token is configured.

			The effective token is resolved in priority order:
			  CIRCLECI_TOKEN env var > CIRCLECI_CLI_TOKEN env var > ~/.circleci/cli.yml
		`),
		Example: heredoc.Doc(`
			# Print the active token:
			$ circleci auth token

			# Export the token into an environment variable:
			$ export CIRCLECI_TOKEN=$(circleci auth token)

			# Pass the token directly to curl:
			$ curl -H "Circle-Token: $(circleci auth token)" \
			    https://circleci.com/api/v2/me
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			token := cfg.Token()
			if token == "" {
				return cierrors.ErrAuthRequired
			}

			// Intentionally no trailing newline — safe for $(circleci auth token).
			fmt.Fprint(f.IOStreams.Out, token)
			return nil
		},
	}
}
