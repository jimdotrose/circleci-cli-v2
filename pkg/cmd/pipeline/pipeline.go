package pipeline

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdPipeline returns the `circleci pipeline` command group.
func NewCmdPipeline(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline <command>",
		Short: "Manage CircleCI pipelines",
		Long: heredoc.Doc(`
			Commands for working with CircleCI pipelines.

			Pipelines are the top-level unit of work in CircleCI. Each push
			to a VCS branch or tag creates a pipeline, which contains one or
			more workflows defined in .circleci/config.yml.
		`),
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdTrigger(f))
	return cmd
}
