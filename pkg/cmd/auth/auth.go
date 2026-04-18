package auth

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdAuth returns the `circleci auth` command group.
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Authenticate with CircleCI",
		Long: heredoc.Doc(`
			Manage authentication credentials for the CircleCI CLI.

			Use 'circleci auth login' to store a personal API token.
			Use 'circleci auth status' to verify the current credentials.
		`),
		Annotations: map[string]string{"group": "core"},
	}

	cmd.AddCommand(NewCmdLogin(f))
	cmd.AddCommand(NewCmdLogout(f))
	cmd.AddCommand(NewCmdStatus(f))
	cmd.AddCommand(NewCmdToken(f))

	return cmd
}
