package orb

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdInfo returns the `circleci orb info` command.
func NewCmdInfo(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "info <orb>",
		Short: "Show details for an orb",
		Long: heredoc.Doc(`
			Show detailed information about an orb in the CircleCI Orb Registry.

			The orb argument is the fully qualified name: namespace/name.
			Displays published versions, description, and 30-day usage statistics.

			Use --json to get the full metadata structure including all version
			sources, suitable for scripting and automation.
		`),
		Example: heredoc.Doc(`
			# Show info for a public orb:
			$ circleci orb info circleci/node

			# Show info as JSON:
			$ circleci orb info circleci/python --json

			# Extract the latest version number:
			$ circleci orb info circleci/node --jq '.versions[0].version'

			# Check whether an orb is private:
			$ circleci orb info myorg/internal-orb --jq '.is_private'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching orb info...")
			o, err := client.GetOrb(args[0])
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, o); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "Name:       %s\n", o.Name)
			fmt.Fprintf(f.IOStreams.Out, "Namespace:  %s\n", o.Namespace)
			fmt.Fprintf(f.IOStreams.Out, "Private:    %v\n", o.IsPrivate)
			fmt.Fprintf(f.IOStreams.Out, "Runs (30d): %d\n", o.Statistics.Last30DayRunCount)

			if len(o.Versions) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "\nNo versions published.")
				return nil
			}

			fmt.Fprintln(f.IOStreams.Out, "\nVersions:")
			for _, v := range o.Versions {
				if v.Description != "" {
					fmt.Fprintf(f.IOStreams.Out, "  %-15s  %s\n", v.Version, v.Description)
				} else {
					fmt.Fprintf(f.IOStreams.Out, "  %s\n", v.Version)
				}
			}
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Orb{})
	return cmd
}
