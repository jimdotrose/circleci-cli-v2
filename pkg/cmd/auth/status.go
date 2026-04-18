package auth

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdStatus returns the `circleci auth status` command.
func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	var showToken bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication state",
		Long: heredoc.Doc(`
			Display the current authentication state for the configured CircleCI host.

			Shows which host the CLI is pointed at and whether a token is configured.
			Use --show-token to reveal the full token value; by default only the last
			four characters are shown.

			This command does not make an API call to validate the token. Use
			'circleci diagnostic' to verify connectivity and token validity.
		`),
		Example: heredoc.Doc(`
			# Show current auth status:
			$ circleci auth status

			# Show the full token value:
			$ circleci auth status --show-token

			# Check status in a script (exits 3 if not authenticated):
			$ circleci auth status || echo "run: circleci auth login"
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			token := cfg.Token()
			host := cfg.Host()

			if token == "" {
				fmt.Fprintf(ios.Out, "✗ Not authenticated to %s\n", host)
				fmt.Fprintln(ios.Out, "  Run: circleci auth login")
				return cierrors.ErrAuthRequired
			}

			var tokenDisplay string
			if showToken {
				tokenDisplay = token
			} else if len(token) >= 4 {
				tokenDisplay = "••••" + token[len(token)-4:]
			} else {
				tokenDisplay = "••••"
			}

			fmt.Fprintf(ios.Out, "✓ Authenticated to %s (token: %s)\n", host, tokenDisplay)
			return nil
		},
	}

	cmd.Flags().BoolVar(&showToken, "show-token", false, "Print the full token value")
	return cmd
}
