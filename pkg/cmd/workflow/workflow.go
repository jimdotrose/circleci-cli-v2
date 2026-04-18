package workflow

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdWorkflow returns the `circleci workflow` command group.
func NewCmdWorkflow(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow <command>",
		Short: "Manage CircleCI workflows",
		Long: heredoc.Doc(`
			Commands for working with CircleCI workflows.

			Workflows orchestrate a set of jobs within a pipeline. Each pipeline
			contains one or more workflows as defined in your config.yml.
		`),
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdCancel(f))
	cmd.AddCommand(NewCmdRerun(f))
	return cmd
}
