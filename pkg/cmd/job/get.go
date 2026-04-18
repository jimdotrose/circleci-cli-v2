package job

import (
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdGet returns the `circleci job get` command.
func NewCmdGet(f *cmdutil.Factory) *cobra.Command {
	var project string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "get <job-number>",
		Short: "Get details for a job",
		Long: heredoc.Doc(`
			Fetch and display details for a specific CircleCI job.

			Requires the project slug and job number. The job number can be
			obtained from 'circleci job list' (the job_number field).
		`),
		Example: heredoc.Doc(`
			# Get job details:
			$ circleci job get 42 --project github/myorg/myrepo

			# Get as JSON:
			$ circleci job get 42 --project github/myorg/myrepo --json

			# Get job status:
			$ circleci job get 42 --project github/myorg/myrepo --jq '.status'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--project is required",
					"Provide the project slug to look up the job.",
					cierrors.ExitBadArguments,
				)
			}

			jobNum, err := strconv.Atoi(args[0])
			if err != nil {
				return cierrors.New(
					"INVALID_ARG",
					"Invalid job number",
					fmt.Sprintf("%q is not a valid job number.", args[0]),
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching job...")
			j, err := client.GetJob(project, jobNum)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, j); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "ID:      %s\n", j.ID)
			fmt.Fprintf(f.IOStreams.Out, "Name:    %s\n", j.Name)
			fmt.Fprintf(f.IOStreams.Out, "Status:  %s\n", j.Status)
			fmt.Fprintf(f.IOStreams.Out, "Type:    %s\n", j.Type)
			if j.JobNumber != nil {
				fmt.Fprintf(f.IOStreams.Out, "Number:  %d\n", *j.JobNumber)
			}
			if j.StartedAt != nil {
				fmt.Fprintf(f.IOStreams.Out, "Started: %s\n", j.StartedAt.Format("2006-01-02 15:04:05"))
			}
			if j.StoppedAt != nil {
				fmt.Fprintf(f.IOStreams.Out, "Stopped: %s\n", j.StoppedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project slug (e.g. github/myorg/myrepo)")
	output.AddFlags(cmd, &opts, &apiclient.Job{})
	return cmd
}
