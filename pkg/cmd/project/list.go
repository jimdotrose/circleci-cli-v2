package project

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdList returns the `circleci project list` command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List followed projects",
		Long: heredoc.Doc(`
			List all CircleCI projects the authenticated user follows.

			Returns the project slugs and VCS metadata for every project
			associated with the current user's token.
		`),
		Example: heredoc.Doc(`
			# List all followed projects:
			$ circleci project list

			# List as JSON:
			$ circleci project list --json

			# Extract slugs with jq:
			$ circleci project list --jq '.[].slug'
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching projects...")
			var all []apiclient.Project
			pageToken := ""
			for {
				items, next, err := client.ListProjects(pageToken)
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
					fmt.Fprintln(f.IOStreams.Out, "No projects found.")
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-12s  %s\n", "VCS", "SLUG")
			for _, p := range all {
				fmt.Fprintf(f.IOStreams.Out, "%-12s  %s\n", p.VCSInfo.Provider, p.Slug)
			}
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Project{})
	return cmd
}
