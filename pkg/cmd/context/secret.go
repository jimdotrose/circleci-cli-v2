package context

import (
	"fmt"
	"io"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdSecret returns the `circleci context secret` command group.
func NewCmdSecret(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret <command>",
		Short: "Manage environment variables in a context",
		Long: heredoc.Doc(`
			Commands for managing environment variables (secrets) stored
			in a CircleCI context.

			Variable values are never returned by the API — only names are
			readable. Use 'set' to create or update a variable.
		`),
	}

	cmd.AddCommand(NewCmdSecretSet(f))
	cmd.AddCommand(NewCmdSecretRemove(f))
	return cmd
}

// NewCmdSecretSet returns the `circleci context secret set` command.
func NewCmdSecretSet(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <context-id> <variable-name>",
		Short: "Set an environment variable in a context",
		Long: heredoc.Doc(`
			Create or update an environment variable in a CircleCI context.

			The value is read from stdin (masked input when interactive).
			To provide the value non-interactively, pipe it:

			  echo "$VALUE" | circleci context secret set <id> <name>
		`),
		Example: heredoc.Doc(`
			# Set a variable interactively (value masked):
			$ circleci context secret set <context-id> AWS_SECRET_KEY

			# Set a variable from stdin:
			$ echo "$MY_SECRET" | circleci context secret set <ctx-id> MY_VAR

			# Set via environment variable:
			$ CIRCLECI_SECRET="$val" circleci context secret set <ctx-id> MYVAR
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			contextID := args[0]
			varName := args[1]

			var value string
			var err error

			if f.IOStreams.IsInteractive {
				value, err = f.IOStreams.ReadPassword(fmt.Sprintf("Value for %s: ", varName))
				if err != nil {
					return cierrors.New(
						"READ_ERROR",
						"Could not read value",
						fmt.Sprintf("Error reading secret value: %v", err),
						cierrors.ExitGeneralError,
					)
				}
			} else {
				// Non-interactive: read from stdin.
				data, err := io.ReadAll(f.IOStreams.In)
				if err != nil {
					return err
				}
				value = strings.TrimRight(string(data), "\r\n")
			}

			if value == "" {
				return cierrors.New(
					"EMPTY_VALUE",
					"Value cannot be empty",
					"Provide a non-empty value for the environment variable.",
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Setting variable...")
			err = client.SetContextVariable(contextID, varName, value)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Set %s in context %s.\n", varName, contextID)
			}
			return nil
		},
	}

	return cmd
}

// NewCmdSecretRemove returns the `circleci context secret remove` command.
func NewCmdSecretRemove(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove <context-id> <variable-name>",
		Short: "Remove an environment variable from a context",
		Long: heredoc.Doc(`
			Delete an environment variable from a CircleCI context.

			This action is irreversible. Use --force to skip the confirmation
			prompt in non-interactive mode.
		`),
		Example: heredoc.Doc(`
			# Remove a variable (prompts for confirmation):
			$ circleci context secret remove <context-id> AWS_SECRET_KEY

			# Remove without prompting:
			$ circleci context secret remove <ctx-id> MY_VAR --force

			# Remove in a CI script:
			$ circleci context secret remove <ctx-id> OLD_KEY --force --no-prompt
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			contextID := args[0]
			varName := args[1]

			if !force && !f.IOStreams.IsInteractive {
				return cierrors.New(
					"CONFIRMATION_REQUIRED",
					"Confirmation required",
					"Pass --force to remove a variable in non-interactive mode.",
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Removing variable...")
			err = client.RemoveContextVariable(contextID, varName)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Removed %s from context %s.\n", varName, contextID)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
