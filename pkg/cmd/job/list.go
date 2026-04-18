package job

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdList returns the `circleci job list` command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list <workflow-id>",
		Short: "List jobs for a workflow",
		Long: heredoc.Doc(`
			List all jobs for a given CircleCI workflow.

			Pass the workflow UUID, which can be obtained from
			'circleci workflow list' or from the CircleCI web app URL.
		`),
		Example: heredoc.Doc(`
			# List jobs in a workflow:
			$ circleci job list 00000000-0000-0000-0000-000000000000

			# List as JSON:
			$ circleci job list <workflow-id> --json

			# List failed jobs only:
			$ circleci job list <workflow-id> --jq '[.[] | select(.status=="failed")]'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching jobs...")
			var all []apiclient.Job
			pageToken := ""
			for {
				items, next, err := client.ListJobs(workflowID, pageToken)
				if err != nil {
					stop()
					return err
				}
				all = append(all, items...)
				if next == "" {
					break
				}
				pageToken = next
			}
			stop()

			if err := opts.Write(f.IOStreams.Out, all); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(all) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintln(f.IOStreams.Out, "No jobs found.")
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-30s  %-12s  %s\n", "NAME", "STATUS", "ID")
			for _, j := range all {
				fmt.Fprintf(f.IOStreams.Out, "%-30s  %-12s  %s\n", j.Name, j.Status, j.ID)
			}
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Job{})
	return cmd
}
