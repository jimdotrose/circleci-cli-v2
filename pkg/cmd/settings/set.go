package settings

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdSet returns the `circleci settings set` command.
func NewCmdSet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a setting value",
		Long: heredoc.Doc(`
			Set a CLI configuration setting and persist it to ~/.circleci/cli.yml.

			Changes take effect immediately for subsequent commands. Environment
			variables (CIRCLECI_TOKEN, CIRCLECI_HOST) always take precedence over
			file values regardless of what 'settings set' writes.

			Available settings: host, token, update_check, telemetry

			To manage API tokens interactively, prefer 'circleci auth login'.
		`),
		Example: heredoc.Doc(`
			# Point the CLI at a CircleCI Server instance:
			$ circleci settings set host https://circleci.mycompany.com

			# Disable telemetry:
			$ circleci settings set telemetry false

			# Disable update notifications:
			$ circleci settings set update_check false
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			if err := cfg.Set(key, value); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return cierrors.New("SAVE_ERROR", "Could not save config",
					err.Error(), cierrors.ExitGeneralError)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ %s set to %s\n", key, value)
			return nil
		},
	}
}
