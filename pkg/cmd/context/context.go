package context

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdContext returns the `circleci context` command group.
func NewCmdContext(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context <command>",
		Short: "Manage CircleCI contexts",
		Long: heredoc.Doc(`
			Commands for managing CircleCI contexts and their environment variables.

			Contexts provide a mechanism for securing and sharing environment
			variables across projects. They are attached to an organization
			and can be referenced in your config with the context key.
		`),
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdShow(f))
	cmd.AddCommand(NewCmdDelete(f))
	cmd.AddCommand(NewCmdSecret(f))
	return cmd
}
