package orb

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

var validSegments = map[string]bool{"major": true, "minor": true, "patch": true}

// NewCmdPromote returns the `circleci orb promote` command.
func NewCmdPromote(f *cmdutil.Factory) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "promote <orb@dev:label> <segment>",
		Short: "Promote a dev orb to a semantic release",
		Long: heredoc.Doc(`
			Promote a development orb version to a numbered semantic release.

			The first argument is the dev orb reference: namespace/name@dev:label.
			The second argument is the version segment to increment: major, minor,
			or patch. The new version is computed from the orb's latest published
			semantic version.

			Use --dry-run to preview the promotion without executing it.
		`),
		Example: heredoc.Doc(`
			# Promote a dev orb, incrementing the patch version:
			$ circleci orb promote myorg/myorb@dev:my-feature patch

			# Promote and bump the minor version:
			$ circleci orb promote myorg/myorb@dev:my-feature minor

			# Preview the promotion without executing:
			$ circleci orb promote myorg/myorb@dev:my-feature patch --dry-run
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			segment := args[1]

			if !validSegments[segment] {
				return cierrors.New(
					"INVALID_SEGMENT",
					"Invalid version segment",
					fmt.Sprintf("Segment must be major, minor, or patch; got %q.", segment),
					cierrors.ExitBadArguments,
				)
			}

			atIdx := strings.Index(ref, "@")
			if atIdx < 0 || !strings.HasPrefix(ref[atIdx+1:], "dev:") {
				return cierrors.New(
					"INVALID_DEV_REF",
					"Invalid dev orb reference",
					fmt.Sprintf("Expected namespace/name@dev:label, got %q.", ref),
					cierrors.ExitBadArguments,
				).WithSuggestions("Example: myorg/myorb@dev:my-feature")
			}
			orbName := ref[:atIdx]
			devVersion := ref[atIdx+1:]

			if dryRun {
				fmt.Fprintf(f.IOStreams.Out, "Would promote %s (%s) with segment %s\n", orbName, devVersion, segment)
				return nil
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Promoting orb...")
			promoted, err := client.PromoteOrb(orbName, devVersion, segment)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Promoted %s to %s\n", orbName, promoted.Version)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview the promotion without executing")
	return cmd
}
