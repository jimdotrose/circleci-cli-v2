package config

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdProcess returns the `circleci config process` command.
func NewCmdProcess(f *cmdutil.Factory) *cobra.Command {
	var orgID string
	var pipelineParams string

	cmd := &cobra.Command{
		Use:   "process [<config-file>]",
		Short: "Process a CircleCI config file (expand orbs and pipeline params)",
		Long: heredoc.Doc(`
			Process a CircleCI configuration file via the compile API.

			Expands orb references, resolves pipeline parameters, and outputs
			the fully-expanded YAML to stdout. Useful for debugging and
			inspecting the final config that CircleCI will execute.

			By default reads .circleci/config.yml from the current directory.
		`),
		Example: heredoc.Doc(`
			# Process the default config:
			$ circleci config process

			# Process with pipeline parameters:
			$ circleci config process --pipeline-params '{"deploy_env":"staging"}'

			# Process a specific file:
			$ circleci config process path/to/config.yml

			# Process and write output to a file:
			$ circleci config process > expanded.yml
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ".circleci/config.yml"
			if len(args) == 1 {
				path = args[0]
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return cierrors.New(
					"FILE_NOT_FOUND",
					"Config file not found",
					fmt.Sprintf("Could not read %q: %v", path, err),
					cierrors.ExitNotFound,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Processing config...")
			resp, err := client.CompileConfig(string(data), pipelineParams, orgID)
			stop()
			if err != nil {
				return err
			}

			if !resp.Valid || len(resp.Errors) > 0 {
				for _, e := range resp.Errors {
					fmt.Fprintf(f.IOStreams.ErrOut, "  error: %s\n", e.Message)
				}
				return cierrors.New(
					"CONFIG_INVALID",
					"Configuration is invalid",
					fmt.Sprintf("Found %d error(s) in %s", len(resp.Errors), path),
					cierrors.ExitValidationFail,
				)
			}

			fmt.Fprint(f.IOStreams.Out, resp.OutputYaml)
			return nil
		},
	}

	cmd.Flags().StringVar(&orgID, "org-id", "", "Organization ID for processing context")
	cmd.Flags().StringVar(&pipelineParams, "pipeline-params", "", "Pipeline parameters as a JSON `object`")
	return cmd
}
