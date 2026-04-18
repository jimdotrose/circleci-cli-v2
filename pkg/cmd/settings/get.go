package settings

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdGet returns the `circleci settings get` command.
func NewCmdGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a setting value",
		Long: heredoc.Doc(`
			Get the current value of a CLI configuration setting.

			The returned value reflects the full precedence chain:
			  CIRCLECI_* env vars > ~/.circleci/cli.yml > built-in defaults

			Available settings: host, token, update_check, telemetry
		`),
		Example: heredoc.Doc(`
			# Get the configured host:
			$ circleci settings get host

			# Get the active token (full value, for scripting):
			$ circleci settings get token

			# Use a setting value in a script:
			$ HOST=$(circleci settings get host)
			$ curl "$HOST/api/v2/me"
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			val, ok := cfg.Get(key)
			if !ok {
				return cierrors.New(
					"UNKNOWN_SETTING",
					fmt.Sprintf("Unknown setting %q", key),
					fmt.Sprintf("Valid settings: %s", "host, token, update_check, telemetry"),
					cierrors.ExitBadArguments,
				)
			}

			fmt.Fprintln(f.IOStreams.Out, val)
			return nil
		},
	}
}
