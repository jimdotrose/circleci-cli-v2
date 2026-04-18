package workflow

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdGet returns the `circleci workflow get` command.
func NewCmdGet(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "get <workflow-id>",
		Short: "Get details for a workflow",
		Long: heredoc.Doc(`
			Fetch and display details for a specific CircleCI workflow.

			Pass the workflow UUID, which can be obtained from
			'circleci workflow list' or from the CircleCI web app URL.
		`),
		Example: heredoc.Doc(`
			# Get workflow details:
			$ circleci workflow get 00000000-0000-0000-0000-000000000000

			# Get as JSON:
			$ circleci workflow get <id> --json

			# Check workflow status:
			$ circleci workflow get <id> --jq '.status'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching workflow...")
			w, err := client.GetWorkflow(id)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, w); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			stoppedAt := "—"
			if w.StoppedAt != nil {
				stoppedAt = w.StoppedAt.Format("2006-01-02 15:04:05")
			}

			fmt.Fprintf(f.IOStreams.Out, "ID:         %s\n", w.ID)
			fmt.Fprintf(f.IOStreams.Out, "Name:       %s\n", w.Name)
			fmt.Fprintf(f.IOStreams.Out, "Pipeline:   %s\n", w.PipelineID)
			fmt.Fprintf(f.IOStreams.Out, "Status:     %s\n", w.Status)
			fmt.Fprintf(f.IOStreams.Out, "Created:    %s\n", w.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(f.IOStreams.Out, "Stopped:    %s\n", stoppedAt)
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Workflow{})
	return cmd
}
