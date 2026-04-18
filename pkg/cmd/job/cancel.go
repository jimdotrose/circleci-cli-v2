package job

import (
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdCancel returns the `circleci job cancel` command.
func NewCmdCancel(f *cmdutil.Factory) *cobra.Command {
	var project string
	var force bool

	cmd := &cobra.Command{
		Use:   "cancel <job-number>",
		Short: "Cancel a running job",
		Long: heredoc.Doc(`
			Cancel a currently running CircleCI job.

			Requires the project slug and job number. Use --force to skip the
			confirmation prompt in non-interactive mode.
		`),
		Example: heredoc.Doc(`
			# Cancel a job (prompts for confirmation):
			$ circleci job cancel 42 --project github/myorg/myrepo

			# Cancel without prompting:
			$ circleci job cancel 42 --project github/myorg/myrepo --force

			# Cancel in a CI script:
			$ circleci job cancel 42 --project github/myorg/myrepo --force --no-prompt
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--project is required",
					"Provide the project slug to cancel the job.",
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

			if !force && !f.IOStreams.IsInteractive {
				return cierrors.New(
					"CONFIRMATION_REQUIRED",
					"Confirmation required",
					"Pass --force to cancel a job in non-interactive mode.",
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Cancelling job...")
			err = client.CancelJob(project, jobNum)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Job %d in %s cancelled.\n", jobNum, project)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project slug (e.g. github/myorg/myrepo)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
