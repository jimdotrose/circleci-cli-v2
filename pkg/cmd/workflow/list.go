package workflow

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdList returns the `circleci workflow list` command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list <pipeline-id>",
		Short: "List workflows for a pipeline",
		Long: heredoc.Doc(`
			List all workflows for a given pipeline.

			Pass the pipeline UUID, which can be obtained from
			'circleci pipeline list' or from the CircleCI web app URL.
		`),
		Example: heredoc.Doc(`
			# List workflows for a pipeline:
			$ circleci workflow list 00000000-0000-0000-0000-000000000000

			# List as JSON:
			$ circleci workflow list <pipeline-id> --json

			# Filter running workflows:
			$ circleci workflow list <pipeline-id> --jq '[.[] | select(.status=="running")]'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching workflows...")
			var all []apiclient.Workflow
			pageToken := ""
			for {
				items, next, err := client.ListWorkflows(pipelineID, pageToken)
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
					fmt.Fprintln(f.IOStreams.Out, "No workflows found.")
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-20s  %-12s  %s\n", "NAME", "STATUS", "ID")
			for _, w := range all {
				fmt.Fprintf(f.IOStreams.Out, "%-20s  %-12s  %s\n", w.Name, w.Status, w.ID)
			}
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Workflow{})
	return cmd
}
