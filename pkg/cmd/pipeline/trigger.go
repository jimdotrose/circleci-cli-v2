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

// NewCmdTrigger returns the `circleci pipeline trigger` command.
func NewCmdTrigger(f *cmdutil.Factory) *cobra.Command {
	var project string
	var branch string
	var tag string
	var parameters string
	var dryRun bool
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger a new pipeline",
		Long: heredoc.Doc(`
			Trigger a new CircleCI pipeline for a project.

			Specify a branch or tag to trigger against. Optionally pass pipeline
			parameters as a JSON object. Parameters must match those declared in
			the 'parameters' section of your config.yml.

			Use --dry-run to preview the request without actually triggering.

			JSON Fields (response): id, state, number, createdAt
		`),
		Example: heredoc.Doc(`
			# Trigger a pipeline on main:
			$ circleci pipeline trigger --project github/myorg/myrepo --branch main

			# Trigger with pipeline parameters:
			$ circleci pipeline trigger --project github/myorg/myrepo --branch main \
			    --parameters '{"deploy_env":"staging","run_tests":true}'

			# Trigger on a tag:
			$ circleci pipeline trigger --project github/myorg/myrepo --tag v1.2.3

			# Preview without triggering:
			$ circleci pipeline trigger --project github/myorg/myrepo --branch main --dry-run
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--project is required",
					"Provide the project slug to trigger a pipeline for.",
					cierrors.ExitBadArguments,
				)
			}
			if branch == "" && tag == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--branch or --tag is required",
					"Provide either --branch or --tag to identify which ref to trigger.",
					cierrors.ExitBadArguments,
				)
			}

			params, err := apiclient.ParseParameters(parameters)
			if err != nil {
				return cierrors.New(
					"INVALID_JSON",
					"Invalid pipeline parameters",
					fmt.Sprintf("Could not parse --parameters as JSON: %v", err),
					cierrors.ExitBadArguments,
				)
			}

			if dryRun {
				ref := branch
				if ref == "" {
					ref = "tag:" + tag
				}
				fmt.Fprintf(f.IOStreams.Out, "Would trigger pipeline:\n")
				fmt.Fprintf(f.IOStreams.Out, "  project:    %s\n", project)
				fmt.Fprintf(f.IOStreams.Out, "  ref:        %s\n", ref)
				if len(params) > 0 {
					fmt.Fprintf(f.IOStreams.Out, "  parameters: %v\n", parameters)
				}
				return nil
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Triggering pipeline...")
			resp, err := client.TriggerPipeline(project, branch, tag, params)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, resp); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Pipeline #%d triggered (ID: %s, state: %s)\n",
					resp.Number, resp.ID, resp.State)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project slug (e.g. github/myorg/myrepo)")
	cmd.Flags().StringVar(&branch, "branch", "", "Branch to trigger against")
	cmd.Flags().StringVar(&tag, "tag", "", "Tag to trigger against")
	cmd.Flags().StringVar(&parameters, "parameters", "", "Pipeline parameters as a JSON `object`")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Print what would be triggered without making API call")
	output.AddFlags(cmd, &opts, &apiclient.TriggerPipelineResponse{})
	return cmd
}
