package auth

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdLogout returns the `circleci auth logout` command.
func NewCmdLogout(f *cmdutil.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored API token",
		Long: heredoc.Doc(`
			Remove the stored API token from ~/.circleci/cli.yml.

			After logout, commands that require authentication will fail until you
			run 'circleci auth login' again. In interactive mode a confirmation
			prompt is shown; pass --yes to skip it.
		`),
		Example: heredoc.Doc(`
			# Log out (prompts for confirmation in interactive mode):
			$ circleci auth logout

			# Log out without confirmation:
			$ circleci auth logout --yes

			# Log out in a script (non-interactive, no prompt):
			$ CI=true circleci auth logout
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			if cfg.Token() == "" {
				fmt.Fprintln(ios.Out, "Not currently authenticated.")
				return nil
			}

			// Prompt for confirmation in interactive mode.
			if ios.IsInteractive && !yes {
				fmt.Fprintf(ios.ErrOut, "Remove token for %s? [y/N] ", cfg.Host())
				var answer string
				fmt.Fscan(ios.In, &answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(ios.Out, "Cancelled.")
					return nil
				}
			}

			if err := cfg.Set("token", ""); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return cierrors.New("SAVE_ERROR", "Could not save config",
					err.Error(), cierrors.ExitGeneralError)
			}

			fmt.Fprintf(ios.Out, "✓ Logged out of %s\n", cfg.Host())
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}
