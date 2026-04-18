package settings

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdSettings returns the `circleci settings` command group.
func NewCmdSettings(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings <command>",
		Short: "Manage CLI configuration settings",
		Long: heredoc.Doc(`
			Read and write CircleCI CLI configuration settings.

			Settings are stored in ~/.circleci/cli.yml. Environment variables
			(CIRCLECI_TOKEN, CIRCLECI_HOST) take precedence over file values.

			To manage API tokens use 'circleci auth'. Settings provides a
			lower-level interface to all configuration keys including token.
		`),
		Annotations: map[string]string{"group": "developer"},
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdSet(f))

	return cmd
}
