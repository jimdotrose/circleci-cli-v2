package job

import (
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdLogs returns the `circleci job logs` command.
func NewCmdLogs(f *cmdutil.Factory) *cobra.Command {
	var project string

	cmd := &cobra.Command{
		Use:   "logs <job-number>",
		Short: "Stream step output for a job",
		Long: heredoc.Doc(`
			Fetch and stream the step-by-step output for a CircleCI job.

			Each step's output is written to stdout in the order steps ran.
			Requires the project slug and job number. The job number can be
			obtained from 'circleci job list' or the CircleCI web UI.

			Output is streamed as plain text — pipe it to a pager or file
			as needed. Non-output lines (progress) are written to stderr.
		`),
		Example: heredoc.Doc(`
			# Stream logs for job 42:
			$ circleci job logs 42 --project github/myorg/myrepo

			# Save logs to a file:
			$ circleci job logs 42 --project github/myorg/myrepo > job-42.log

			# Page through long output:
			$ circleci job logs 42 --project github/myorg/myrepo | less

			# Show only lines containing "error":
			$ circleci job logs 42 --project github/myorg/myrepo | grep -i error
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--project is required",
					"Provide the project slug to look up the job.",
					cierrors.ExitBadArguments,
				).WithSuggestions("Format: --project github/myorg/myrepo")
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

			stop := f.IOStreams.StartSpinner("Fetching job logs...")
			err = client.GetJobLogs(project, jobNum, f.IOStreams.Out)
			stop()
			return err
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project slug (e.g. github/myorg/myrepo)")
	return cmd
}
