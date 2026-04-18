package project

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdFollow returns the `circleci project follow` command.
func NewCmdFollow(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "follow <project-slug>",
		Short: "Follow a project to enable builds",
		Long: heredoc.Doc(`
			Follow a CircleCI project, enabling builds for new commits.

			The project slug format is <vcs-type>/<org>/<repo>, for example
			github/myorg/myrepo. Following a project links it to your account
			and allows the CLI to trigger pipelines for it.
		`),
		Example: heredoc.Doc(`
			# Follow a GitHub project:
			$ circleci project follow github/myorg/myrepo

			# Follow a Bitbucket project:
			$ circleci project follow bitbucket/myteam/myservice

			# Follow and output result as JSON:
			$ circleci project follow github/myorg/myrepo --json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Following project...")
			p, err := client.FollowProject(slug)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, p); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Now following %s\n", slug)
			}
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Project{})
	return cmd
}
