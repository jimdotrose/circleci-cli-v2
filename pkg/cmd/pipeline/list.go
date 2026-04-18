package pipeline

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdList returns the `circleci pipeline list` command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var project string
	var branch string
	var limit int
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipelines for a project",
		Long: heredoc.Doc(`
			List recent pipelines for a CircleCI project.

			The project slug format is <vcs-type>/<org>/<repo>, for example
			github/myorg/myrepo. Use --branch to filter to a specific branch
			and --limit to cap the number of results returned.
		`),
		Example: heredoc.Doc(`
			# List pipelines for a project:
			$ circleci pipeline list --project github/myorg/myrepo

			# Filter to a specific branch:
			$ circleci pipeline list --project github/myorg/myrepo --branch main

			# List as JSON and filter with jq:
			$ circleci pipeline list --project github/myorg/myrepo --jq '.[].id'
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--project is required",
					"Provide the project slug to list pipelines for.",
					cierrors.ExitBadArguments,
				).WithSuggestions("Format: --project github/myorg/myrepo")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching pipelines...")
			var all []apiclient.Pipeline
			pageToken := ""
			for {
				items, next, err := client.ListPipelines(project, branch, pageToken)
				if err != nil {
					stop()
					return err
				}
				all = append(all, items...)
				if next == "" || (limit > 0 && len(all) >= limit) {
					break
				}
				pageToken = next
			}
			stop()

			if limit > 0 && len(all) > limit {
				all = all[:limit]
			}

			if err := opts.Write(f.IOStreams.Out, all); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(all) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintln(f.IOStreams.Out, "No pipelines found.")
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-6s  %-10s  %-20s  %s\n", "NUMBER", "STATE", "CREATED", "ID")
			for _, p := range all {
				branch := ""
				if p.VCS != nil {
					branch = p.VCS.Branch
				}
				_ = branch
				fmt.Fprintf(f.IOStreams.Out, "%-6d  %-10s  %-20s  %s\n",
					p.Number, p.State, p.CreatedAt.Format("2006-01-02 15:04"), p.ID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project slug (e.g. github/myorg/myrepo)")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Filter by branch name")
	cmd.Flags().IntVarP(&limit, "limit", "L", 20, "Maximum number of pipelines to return")
	output.AddFlags(cmd, &opts, &apiclient.Pipeline{})
	return cmd
}
