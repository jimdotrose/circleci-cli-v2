package context

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdDelete returns the `circleci context delete` command.
func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <context-id>",
		Short: "Delete a context",
		Long: heredoc.Doc(`
			Permanently delete a CircleCI context and all its environment variables.

			This action is irreversible. All environment variables stored in the
			context will be lost. Any pipelines referencing this context will fail
			until updated to use a different context.

			Use --force to skip the confirmation prompt in non-interactive mode.
		`),
		Example: heredoc.Doc(`
			# Delete a context (prompts for confirmation):
			$ circleci context delete 00000000-0000-0000-0000-000000000000

			# Delete without confirmation:
			$ circleci context delete <id> --force

			# Delete in a CI script:
			$ circleci context delete <id> --force --no-prompt
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !force && f.IOStreams.IsInteractive {
				// Show context ID before deleting and ask for y/N confirmation.
				fmt.Fprintf(f.IOStreams.ErrOut, "Permanently delete context %s? [y/N] ", id)

				var answer string
				fmt.Fscan(f.IOStreams.In, &answer)
				if answer != "y" && answer != "Y" {
					return cierrors.New(
						"CANCELLED",
						"Delete cancelled",
						"No changes made.",
						cierrors.ExitCancelled,
					)
				}
			} else if !force && !f.IOStreams.IsInteractive {
				return cierrors.New(
					"CONFIRMATION_REQUIRED",
					"Confirmation required",
					"Pass --force to delete a context in non-interactive mode.",
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Deleting context...")
			err = client.DeleteContext(id)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Context %s deleted.\n", id)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
