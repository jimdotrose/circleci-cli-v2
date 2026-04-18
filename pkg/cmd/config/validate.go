package config

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdValidate returns the `circleci config validate` command.
func NewCmdValidate(f *cmdutil.Factory) *cobra.Command {
	var orgID string

	cmd := &cobra.Command{
		Use:   "validate [<config-file>]",
		Short: "Validate a CircleCI config file",
		Long: heredoc.Doc(`
			Validate a CircleCI configuration file using the CircleCI compile API.

			By default reads .circleci/config.yml from the current directory.
			Exits 0 on success, 7 if the configuration is invalid.

			Pass --org-id to include organization-level settings (pipeline
			parameters, orb allowlists) in the validation.
		`),
		Example: heredoc.Doc(`
			# Validate the default config file:
			$ circleci config validate

			# Validate a specific file:
			$ circleci config validate path/to/config.yml

			# Validate with organization-level settings:
			$ circleci config validate --org-id 00000000-0000-0000-0000-000000000000
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
				).WithSuggestions("Run: circleci config generate")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Validating config...")
			resp, err := client.CompileConfig(string(data), "", orgID)
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

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Config file at %s is valid.\n", path)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&orgID, "org-id", "", "Organization ID for validation context")
	return cmd
}
