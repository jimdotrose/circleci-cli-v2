package pipeline

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdGet returns the `circleci pipeline get` command.
func NewCmdGet(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "get <pipeline-id>",
		Short: "Get details for a pipeline",
		Long: heredoc.Doc(`
			Fetch and display details for a specific CircleCI pipeline.

			Pass the pipeline UUID, which can be obtained from
			'circleci pipeline list' or from the CircleCI web app URL.
		`),
		Example: heredoc.Doc(`
			# Get pipeline details:
			$ circleci pipeline get 00000000-0000-0000-0000-000000000000

			# Get as JSON:
			$ circleci pipeline get <id> --json

			# Get the pipeline state:
			$ circleci pipeline get <id> --jq '.state'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching pipeline...")
			p, err := client.GetPipeline(id)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, p); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			branch, tag := "", ""
			if p.VCS != nil {
				branch = p.VCS.Branch
				tag = p.VCS.Tag
			}
			ref := branch
			if ref == "" {
				ref = tag
			}

			fmt.Fprintf(f.IOStreams.Out, "ID:      %s\n", p.ID)
			fmt.Fprintf(f.IOStreams.Out, "Number:  %d\n", p.Number)
			fmt.Fprintf(f.IOStreams.Out, "Project: %s\n", p.ProjectSlug)
			fmt.Fprintf(f.IOStreams.Out, "State:   %s\n", p.State)
			fmt.Fprintf(f.IOStreams.Out, "Ref:     %s\n", ref)
			fmt.Fprintf(f.IOStreams.Out, "Created: %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Pipeline{})
	return cmd
}
