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

// NewCmdList returns the `circleci context list` command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var orgSlug string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contexts for an organization",
		Long: heredoc.Doc(`
			List all CircleCI contexts visible to the authenticated user.

			Use --org-slug to filter to a specific organization. The slug
			format is <vcs-type>/<org-name>, e.g. github/myorg.
		`),
		Example: heredoc.Doc(`
			# List contexts for a GitHub org:
			$ circleci context list --org-slug github/myorg

			# List as JSON:
			$ circleci context list --org-slug github/myorg --json

			# Filter with jq:
			$ circleci context list --org-slug github/myorg --jq '.[].name'
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if orgSlug == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--org-slug is required",
					"Provide the organization slug to list contexts for.",
					cierrors.ExitBadArguments,
				).WithSuggestions("Format: --org-slug github/myorg")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching contexts...")
			var all []apiclient.Context
			pageToken := ""
			for {
				items, next, err := client.ListContexts(orgSlug, "", pageToken)
				if err != nil {
					stop()
					return err
				}
				all = append(all, items...)
				if next == "" {
					break
				}
				pageToken = next
			}
			stop()

			if err := opts.Write(f.IOStreams.Out, all); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(all) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintln(f.IOStreams.Out, "No contexts found.")
				}
				return nil
			}

			if opts.Plain {
				for _, ctx := range all {
					fmt.Fprintf(f.IOStreams.Out, "%s\t%s\n", ctx.ID, ctx.Name)
				}
				return nil
			}

			for _, ctx := range all {
				fmt.Fprintf(f.IOStreams.Out, "%-40s  %s\n", ctx.ID, ctx.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&orgSlug, "org-slug", "", "Organization slug (e.g. github/myorg)")
	output.AddFlags(cmd, &opts, &apiclient.Context{})
	return cmd
}
