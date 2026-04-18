package workflow

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdCancel returns the `circleci workflow cancel` command.
func NewCmdCancel(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "cancel <workflow-id>",
		Short: "Cancel a running workflow",
		Long: heredoc.Doc(`
			Cancel a currently running CircleCI workflow.

			Only workflows in a running state can be cancelled. Use --force to
			skip the confirmation prompt when running non-interactively.
		`),
		Example: heredoc.Doc(`
			# Cancel a workflow (prompts for confirmation):
			$ circleci workflow cancel 00000000-0000-0000-0000-000000000000

			# Cancel without prompting:
			$ circleci workflow cancel <id> --force

			# Cancel in a CI script:
			$ circleci workflow cancel <id> --force --no-prompt
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !force && !f.IOStreams.IsInteractive {
				return cierrors.New(
					"CONFIRMATION_REQUIRED",
					"Confirmation required",
					"Pass --force to cancel a workflow in non-interactive mode.",
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Cancelling workflow...")
			err = client.CancelWorkflow(id)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Workflow %s cancelled.\n", id)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
