package runner

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdRunner returns the `circleci runner` command group.
func NewCmdRunner(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runner <command>",
		Short: "Manage CircleCI self-hosted runners",
		Long: heredoc.Doc(`
			Commands for managing CircleCI self-hosted runners.

			Self-hosted runners allow you to execute CircleCI jobs on your own
			infrastructure. Each runner is associated with a resource class and
			authenticated with a token.
		`),
	}

	cmd.AddCommand(NewCmdResourceClass(f))
	cmd.AddCommand(NewCmdToken(f))
	cmd.AddCommand(NewCmdInstance(f))
	return cmd
}
