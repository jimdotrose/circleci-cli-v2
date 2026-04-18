package orb

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdPublish returns the `circleci orb publish` command.
func NewCmdPublish(f *cmdutil.Factory) *cobra.Command {
	var ref string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "publish <file>",
		Short: "Publish an orb version",
		Long: heredoc.Doc(`
			Publish a new version of an orb from a YAML source file.

			Provide the orb reference with --ref in the format namespace/name@version,
			where version is a semantic version (1.2.3) or a dev label (dev:my-feature).
			Dev versions can be promoted to stable releases with 'circleci orb promote'.

			Use --dry-run to preview the publish operation without calling the API.
		`),
		Example: heredoc.Doc(`
			# Publish a stable semantic version:
			$ circleci orb publish orb.yml --ref myorg/myorb@1.2.0

			# Publish a dev version for testing:
			$ circleci orb publish orb.yml --ref myorg/myorb@dev:my-feature

			# Preview without publishing:
			$ circleci orb publish orb.yml --ref myorg/myorb@1.2.0 --dry-run
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			if ref == "" {
				return cierrors.New(
					"MISSING_ARG",
					"--ref is required",
					"Provide the orb reference with --ref namespace/name@version.",
					cierrors.ExitBadArguments,
				).WithSuggestions(
					"Example: --ref myorg/myorb@1.0.0",
					"Dev version: --ref myorg/myorb@dev:my-feature",
				)
			}

			atIdx := strings.Index(ref, "@")
			if atIdx < 0 {
				return cierrors.New(
					"INVALID_ORB_REF",
					"Invalid orb reference",
					fmt.Sprintf("Expected namespace/name@version, got %q.", ref),
					cierrors.ExitBadArguments,
				).WithSuggestions(
					"Example: --ref myorg/myorb@1.0.0",
					"Dev version: --ref myorg/myorb@dev:my-feature",
				)
			}
			orbName := ref[:atIdx]
			version := ref[atIdx+1:]

			if !strings.Contains(orbName, "/") {
				return cierrors.New(
					"INVALID_ORB_REF",
					"Invalid orb name",
					fmt.Sprintf("Orb name must be namespace/name, got %q.", orbName),
					cierrors.ExitBadArguments,
				)
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return cierrors.New(
					"FILE_NOT_FOUND",
					"Orb file not found",
					fmt.Sprintf("Could not read %q: %v", filePath, err),
					cierrors.ExitNotFound,
				)
			}

			if dryRun {
				fmt.Fprintf(f.IOStreams.Out, "Would publish %s@%s from %s\n", orbName, version, filePath)
				return nil
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Publishing orb...")
			published, err := client.PublishOrb(orbName, version, string(data))
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Published %s@%s\n", orbName, published.Version)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ref, "ref", "", "Orb reference: namespace/name@version (required)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview without publishing")
	return cmd
}
