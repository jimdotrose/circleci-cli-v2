package project

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdProject returns the `circleci project` command group.
func NewCmdProject(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project <command>",
		Short: "Manage CircleCI projects",
		Long: heredoc.Doc(`
			Commands for working with CircleCI projects.

			Projects are the unit of work associated with a VCS repository.
			Each project can have environment variables, follow/unfollow state,
			and pipeline triggers.
		`),
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdFollow(f))
	cmd.AddCommand(NewCmdEnv(f))
	return cmd
}
