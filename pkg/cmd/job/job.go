package job

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdJob returns the `circleci job` command group.
func NewCmdJob(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job <command>",
		Short: "Manage CircleCI jobs",
		Long: heredoc.Doc(`
			Commands for working with CircleCI jobs.

			Jobs are the individual units of work within a workflow. Each job
			runs in a separate executor (docker, machine, etc.) and produces
			artifacts and test results.
		`),
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdCancel(f))
	cmd.AddCommand(NewCmdArtifacts(f))
	return cmd
}
