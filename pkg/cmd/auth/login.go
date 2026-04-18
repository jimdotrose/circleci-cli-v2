package auth

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdLogin returns the `circleci auth login` command.
func NewCmdLogin(f *cmdutil.Factory) *cobra.Command {
	var withToken bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a CircleCI API token",
		Long: heredoc.Doc(`
			Store a CircleCI personal API token for use by the CLI.

			The token is saved to ~/.circleci/cli.yml and used for all subsequent
			commands that require authentication. To authenticate with a CircleCI
			Server instance, set the host with:

			  circleci settings set host https://circleci.mycompany.com

			In non-interactive mode (CI=true or --no-prompt), the token must be
			provided via the CIRCLECI_TOKEN environment variable or piped to stdin
			with --with-token. The command exits with code 3 if no token is found.
		`),
		Example: heredoc.Doc(`
			# Interactive login (prompts for token with masked input):
			$ circleci auth login

			# Non-interactive login via environment variable:
			$ CIRCLECI_TOKEN=mytoken circleci auth login --no-prompt

			# Pipe a token from stdin:
			$ echo "$MY_TOKEN" | circleci auth login --with-token

			# Log in to a CircleCI Server instance:
			$ circleci settings set host https://circleci.mycompany.com
			$ circleci auth login
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			var token string

			switch {
			case withToken:
				data, err := io.ReadAll(ios.In)
				if err != nil {
					return cierrors.New("READ_ERROR", "Could not read token from stdin",
						err.Error(), cierrors.ExitGeneralError)
				}
				token = strings.TrimRight(string(data), "\r\n")

			case !ios.IsInteractive:
				token = os.Getenv("CIRCLECI_TOKEN")
				if token == "" {
					token = os.Getenv("CIRCLECI_CLI_TOKEN")
				}
				if token == "" {
					return cierrors.New(
						"AUTH_REQUIRED",
						"No token provided",
						"In non-interactive mode the CIRCLECI_TOKEN environment variable must be set.",
						cierrors.ExitAuthError,
					).WithSuggestions(
						"Set CIRCLECI_TOKEN=<your-token>",
						"Or run interactively: circleci auth login",
					)
				}

			default:
				var err error
				token, err = ios.ReadPassword("Paste your CircleCI API token: ")
				if err != nil {
					return cierrors.New("READ_ERROR", "Could not read token",
						err.Error(), cierrors.ExitGeneralError)
				}
			}

			token = strings.TrimSpace(token)
			if token == "" {
				return cierrors.New(
					"EMPTY_TOKEN",
					"Token cannot be empty",
					"Provide a valid CircleCI personal API token.",
					cierrors.ExitBadArguments,
				).WithSuggestions("Generate a token at https://app.circleci.com/settings/user/tokens")
			}

			if err := cfg.Set("token", token); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return cierrors.New("SAVE_ERROR", "Could not save config",
					err.Error(), cierrors.ExitGeneralError)
			}

			if !ios.Quiet {
				fmt.Fprintf(ios.Out, "✓ Token stored to %s\n", cfg.Path())
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&withToken, "with-token", false, "Read token from stdin")
	return cmd
}
