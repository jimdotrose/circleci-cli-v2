package orb

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdValidate returns the `circleci orb validate` command.
func NewCmdValidate(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [<file>]",
		Short: "Validate an orb YAML file",
		Long: heredoc.Doc(`
			Validate an orb YAML file against the CircleCI orb schema.

			By default reads orb.yml from the current directory.
			Exits 0 when the orb is valid, 7 when validation fails.

			Validation checks the YAML structure and schema without publishing.
			Use 'circleci orb publish' after validation to publish to the registry.
		`),
		Example: heredoc.Doc(`
			# Validate orb.yml in the current directory:
			$ circleci orb validate

			# Validate a specific file:
			$ circleci orb validate src/orb.yml

			# Use in a CI pipeline (non-zero exit on failure):
			$ circleci orb validate && circleci orb publish orb.yml myorg/myorb@dev:ci
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "orb.yml"
			if len(args) == 1 {
				path = args[0]
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return cierrors.New(
					"FILE_NOT_FOUND",
					"Orb file not found",
					fmt.Sprintf("Could not read %q: %v", path, err),
					cierrors.ExitNotFound,
				).WithSuggestions("Check the file path and try again")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Validating orb...")
			resp, err := client.ValidateOrb(string(data))
			stop()
			if err != nil {
				return err
			}

			if !resp.Valid || len(resp.Errors) > 0 {
				for _, e := range resp.Errors {
					fmt.Fprintf(f.IOStreams.ErrOut, "  error: %s\n", e)
				}
				return cierrors.New(
					"ORB_INVALID",
					"Orb is invalid",
					fmt.Sprintf("Found %d error(s) in %s", len(resp.Errors), path),
					cierrors.ExitValidationFail,
				)
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Orb at %s is valid.\n", path)
			}
			return nil
		},
	}

	return cmd
}
