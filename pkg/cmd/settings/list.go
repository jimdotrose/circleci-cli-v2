package settings

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdList returns the `circleci settings list` command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
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

			JSON Fields: host, token, update_check, telemetry
		`),
		Example: heredoc.Doc(`
			# List all settings:
			$ circleci settings list

			# List as JSON:
			$ circleci settings list --json

			# Use in a shell script:
			$ HOST=$(circleci settings get host)
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			if asJSON {
				m := map[string]string{}
				for _, key := range cfg.Keys() {
					val, _ := cfg.Get(key)
					if key == "token" {
						if val == "" {
							val = ""
						}
					}
					m[key] = val
				}
				out, _ := json.MarshalIndent(m, "", "  ")
				fmt.Fprintln(f.IOStreams.Out, string(out))
				return nil
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

	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Output as JSON")
	return cmd
}
