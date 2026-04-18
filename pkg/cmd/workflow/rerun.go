package workflow

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdRerun returns the `circleci workflow rerun` command.
func NewCmdRerun(f *cmdutil.Factory) *cobra.Command {
	var fromFailed bool

	cmd := &cobra.Command{
		Use:   "rerun <workflow-id>",
		Short: "Rerun a workflow",
		Long: heredoc.Doc(`
			Rerun a CircleCI workflow from the beginning, or from failed jobs.

			By default reruns all jobs. Use --failed to rerun only the jobs
			that failed in the original run, saving credit and time.
		`),
		Example: heredoc.Doc(`
			# Rerun all jobs in a workflow:
			$ circleci workflow rerun 00000000-0000-0000-0000-000000000000

			# Rerun only failed jobs:
			$ circleci workflow rerun <id> --failed

			# Rerun in a script:
			$ circleci workflow rerun <id> --failed --no-prompt
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			msg := "Rerunning workflow..."
			if fromFailed {
				msg = "Rerunning failed jobs..."
			}
			stop := f.IOStreams.StartSpinner(msg)
			err = client.RerunWorkflow(id, fromFailed)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				if fromFailed {
					fmt.Fprintf(f.IOStreams.Out, "✓ Rerunning failed jobs for workflow %s.\n", id)
				} else {
					fmt.Fprintf(f.IOStreams.Out, "✓ Rerunning workflow %s.\n", id)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&fromFailed, "failed", false, "Rerun only failed jobs")
	return cmd
}
