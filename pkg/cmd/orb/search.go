package orb

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdSearch returns the `circleci orb search` command.
func NewCmdSearch(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for orbs in the registry",
		Long: heredoc.Doc(`
			Search the CircleCI Orb Registry for orbs matching a query string.

			Returns orbs whose name or description matches the query.
			Use --json for machine-readable output suitable for scripting.
		`),
		Example: heredoc.Doc(`
			# Search for node-related orbs:
			$ circleci orb search node

			# Search and output as JSON:
			$ circleci orb search aws --json

			# Extract just the orb names:
			$ circleci orb search docker --json --jq '.[].name'

			# Count results:
			$ circleci orb search python --json | jq length
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Searching orbs...")
			orbs, err := client.SearchOrbs(args[0])
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, orbs); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(orbs) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintf(f.IOStreams.Out, "No orbs found matching %q.\n", args[0])
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-45s  %s\n", "ORB", "LATEST VERSION")
			for _, o := range orbs {
				latest := "-"
				if len(o.Versions) > 0 {
					latest = o.Versions[0].Version
				}
				fmt.Fprintf(f.IOStreams.Out, "%-45s  %s\n", o.Name, latest)
			}
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Orb{})
	return cmd
}
