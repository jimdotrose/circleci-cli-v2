package job

import (
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdArtifacts returns the `circleci job artifacts` command.
func NewCmdArtifacts(f *cmdutil.Factory) *cobra.Command {
	var project string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "artifacts <job-number>",
		Short: "List artifacts for a job",
		Long: heredoc.Doc(`
			List all artifacts produced by a CircleCI job.

			Requires the project slug and job number. Artifact URLs can be
			used to download files produced by the job (test reports,
			coverage, binaries, etc.).
		`),
		Example: heredoc.Doc(`
			# List artifacts for a job:
			$ circleci job artifacts 42 --project github/myorg/myrepo

			# List artifact URLs only:
			$ circleci job artifacts 42 --project github/myorg/myrepo --jq '.[].url'

			# List as JSON:
			$ circleci job artifacts 42 --project github/myorg/myrepo --json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--project is required",
					"Provide the project slug to list artifacts for.",
					cierrors.ExitBadArguments,
				)
			}

			jobNum, err := strconv.Atoi(args[0])
			if err != nil {
				return cierrors.New(
					"INVALID_ARG",
					"Invalid job number",
					fmt.Sprintf("%q is not a valid job number.", args[0]),
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching artifacts...")
			var all []apiclient.Artifact
			pageToken := ""
			for {
				items, next, err := client.ListArtifacts(project, jobNum, pageToken)
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
					fmt.Fprintln(f.IOStreams.Out, "No artifacts found.")
				}
				return nil
			}

			for _, a := range all {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", a.URL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project slug (e.g. github/myorg/myrepo)")
	output.AddFlags(cmd, &opts, &apiclient.Artifact{})
	return cmd
}
