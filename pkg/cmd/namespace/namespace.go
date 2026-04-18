package namespace

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdNamespace returns the `circleci namespace` command group.
func NewCmdNamespace(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespace <command>",
		Short: "Manage orb namespaces",
		Long: heredoc.Doc(`
			Commands for managing CircleCI orb namespaces.

			A namespace is a unique prefix used to publish orbs. It is owned
			by an organization or account and must be claimed before publishing.
		`),
	}

	cmd.AddCommand(NewCmdNamespaceCreate(f))
	return cmd
}

// NewCmdNamespaceCreate returns `circleci namespace create`.
func NewCmdNamespaceCreate(f *cmdutil.Factory) *cobra.Command {
	var ownerID string
	var ownerType string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an orb namespace",
		Long: heredoc.Doc(`
			Claim a new orb namespace for an organization or account.

			The namespace name must be globally unique across CircleCI. Once
			claimed, all orbs published under it will be prefixed with the
			namespace name (e.g. myorg/myorb).

			Use --owner-id with your organization UUID and --owner-type set
			to "organization" or "account".
		`),
		Example: heredoc.Doc(`
			# Create a namespace for an organization:
			$ circleci namespace create myorg \
			    --owner-id $ORG_ID --owner-type organization

			# Create a namespace for a personal account:
			$ circleci namespace create myname \
			    --owner-id $ACCOUNT_ID --owner-type account

			# Create and inspect the response:
			$ circleci namespace create myorg \
			    --owner-id $ORG_ID --owner-type organization --json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization or account UUID.", cierrors.ExitBadArguments)
			}
			if ownerType == "" {
				return cierrors.New("MISSING_ARG", "--owner-type is required",
					"Provide 'organization' or 'account'.", cierrors.ExitBadArguments)
			}
			if ownerType != "organization" && ownerType != "account" {
				return cierrors.New("INVALID_ARG", "Invalid --owner-type",
					"Must be 'organization' or 'account'.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Creating namespace...")
			result, err := client.CreateNamespace(args[0], ownerID, ownerType)
			stop()
			if err != nil {
				return err
			}

			showJSON, _ := cmd.Flags().GetBool("json")
			if showJSON {
				out, _ := json.MarshalIndent(result, "", "  ")
				fmt.Fprintln(f.IOStreams.Out, string(out))
				return nil
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Created namespace %s\n", args[0])
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization or account UUID")
	cmd.Flags().StringVar(&ownerType, "owner-type", "", "Owner type: 'organization' or 'account'")
	cmd.Flags().Bool("json", false, "Output result as JSON")
	return cmd
}
