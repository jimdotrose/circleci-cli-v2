package runner

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdToken returns the `circleci runner token` command group.
func NewCmdToken(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token <command>",
		Short: "Manage runner authentication tokens",
		Long: heredoc.Doc(`
			Commands for managing authentication tokens for self-hosted runners.

			Tokens are created per resource class and used by runner agents to
			authenticate with CircleCI. A token is only shown once at creation.
		`),
	}

	cmd.AddCommand(NewCmdTokenList(f))
	cmd.AddCommand(NewCmdTokenCreate(f))
	cmd.AddCommand(NewCmdTokenDelete(f))
	return cmd
}

// NewCmdTokenList returns `circleci runner token list`.
func NewCmdTokenList(f *cmdutil.Factory) *cobra.Command {
	var resourceClass string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List runner tokens for a resource class",
		Long: heredoc.Doc(`
			List authentication tokens for a self-hosted runner resource class.

			Token values are not returned by the API — only metadata (ID,
			nickname, creation time) is shown. To create a new token, use
			'circleci runner token create'.
		`),
		Example: heredoc.Doc(`
			# List tokens for a resource class:
			$ circleci runner token list --resource-class myorg/my-runner

			# List as JSON:
			$ circleci runner token list --resource-class myorg/my-runner --json

			# Extract token IDs:
			$ circleci runner token list --resource-class myorg/my-runner \
			    --jq '.[].id'
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resourceClass == "" {
				return cierrors.New("MISSING_ARG", "--resource-class is required",
					"Provide the resource class to list tokens for.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching tokens...")
			tokens, err := client.ListRunnerTokens(resourceClass)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, tokens); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(tokens) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintln(f.IOStreams.Out, "No tokens found.")
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-36s  %-20s  %s\n", "ID", "NICKNAME", "CREATED")
			for _, t := range tokens {
				fmt.Fprintf(f.IOStreams.Out, "%-36s  %-20s  %s\n",
					t.ID, t.Nickname, t.CreatedAt.Format("2006-01-02"))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&resourceClass, "resource-class", "", "Resource class to list tokens for")
	output.AddFlags(cmd, &opts, &apiclient.RunnerToken{})
	return cmd
}

// NewCmdTokenCreate returns `circleci runner token create`.
func NewCmdTokenCreate(f *cmdutil.Factory) *cobra.Command {
	var resourceClass string
	var nickname string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a runner authentication token",
		Long: heredoc.Doc(`
			Create a new authentication token for a self-hosted runner resource class.

			The token value is shown only once. Store it securely — it cannot
			be retrieved again after creation. Use --nickname to label the token
			for easy identification.
		`),
		Example: heredoc.Doc(`
			# Create a token for a resource class:
			$ circleci runner token create --resource-class myorg/my-runner

			# Create with a nickname:
			$ circleci runner token create --resource-class myorg/my-runner \
			    --nickname "prod-server-01"

			# Create and output as JSON (includes token value):
			$ circleci runner token create --resource-class myorg/my-runner --json
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resourceClass == "" {
				return cierrors.New("MISSING_ARG", "--resource-class is required",
					"Provide the resource class to create a token for.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Creating token...")
			tok, err := client.CreateRunnerToken(resourceClass, nickname)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, tok); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Token created (ID: %s)\n", tok.ID)
			if tok.Token != "" {
				fmt.Fprintf(f.IOStreams.Out, "  Token value (save this — not shown again):\n")
				fmt.Fprintf(f.IOStreams.Out, "  %s\n", tok.Token)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&resourceClass, "resource-class", "", "Resource class to create token for")
	cmd.Flags().StringVar(&nickname, "nickname", "", "Human-readable label for the token")
	output.AddFlags(cmd, &opts, &apiclient.RunnerToken{})
	return cmd
}

// NewCmdTokenDelete returns `circleci runner token delete`.
func NewCmdTokenDelete(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <token-id>",
		Short: "Delete a runner token",
		Long: heredoc.Doc(`
			Delete a self-hosted runner authentication token.

			Any runner agent using this token will immediately lose the ability
			to claim jobs. Use --force to skip the confirmation prompt.
		`),
		Example: heredoc.Doc(`
			# Delete a token (prompts for confirmation):
			$ circleci runner token delete 00000000-0000-0000-0000-000000000000

			# Delete without prompting:
			$ circleci runner token delete <id> --force

			# Delete in a CI script:
			$ circleci runner token delete <id> --force --no-prompt
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force && !f.IOStreams.IsInteractive {
				return cierrors.New("CONFIRMATION_REQUIRED", "Confirmation required",
					"Pass --force to delete a token non-interactively.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Deleting token...")
			err = client.DeleteRunnerToken(args[0])
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Deleted token %s\n", args[0])
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
