package context

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdCreate returns the `circleci context create` command.
func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var orgID string
	var orgType string
	var dryRun bool
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new context",
		Long: heredoc.Doc(`
			Create a new CircleCI context for an organization.

			Contexts are scoped to an organization. Provide the organization's
			ID and type (organization or account). The created context can then
			be referenced in config.yml and have environment variables added.

			Use --dry-run to preview the creation without making an API call.

			JSON Fields: id, name, createdAt
		`),
		Example: heredoc.Doc(`
			# Create a context for an organization:
			$ circleci context create staging --org-id 00000000-0000-0000-0000-000000000000

			# Create and output as JSON:
			$ circleci context create production --org-id <id> --json

			# Preview creation without making an API call:
			$ circleci context create staging --org-id <id> --dry-run

			# Create for an account (user) context:
			$ circleci context create my-ctx --org-id <id> --org-type account
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if orgID == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--org-id is required",
					"Provide the organization ID to create the context under.",
					cierrors.ExitBadArguments,
				)
			}

			if dryRun {
				fmt.Fprintf(f.IOStreams.Out, "Would create context:\n")
				fmt.Fprintf(f.IOStreams.Out, "  name:     %s\n", name)
				fmt.Fprintf(f.IOStreams.Out, "  org-id:   %s\n", orgID)
				fmt.Fprintf(f.IOStreams.Out, "  org-type: %s\n", orgType)
				return nil
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Creating context...")
			ctx, err := client.CreateContext(name, orgID, orgType)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, ctx); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Context %q created (ID: %s)\n", ctx.Name, ctx.ID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&orgID, "org-id", "", "Organization ID")
	cmd.Flags().StringVar(&orgType, "org-type", "organization", "Organization type: organization or account (default: organization)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Print what would be created without making API call")
	output.AddFlags(cmd, &opts, &apiclient.Context{})
	return cmd
}
