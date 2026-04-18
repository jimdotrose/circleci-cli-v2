package orb

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdList returns the `circleci orb list` command.
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var namespace string
	var private bool
	var sort string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List orbs in the registry",
		Long: heredoc.Doc(`
			List orbs available in the CircleCI Orb Registry.

			Optionally filter by --namespace to list orbs published by a specific
			organization. Use --private to include private orbs your token can
			access. Use --sort to control ordering.
		`),
		Example: heredoc.Doc(`
			# List all public orbs:
			$ circleci orb list

			# List orbs for a namespace:
			$ circleci orb list --namespace circleci

			# List as JSON:
			$ circleci orb list --namespace myorg --json

			# Extract orb names with jq:
			$ circleci orb list --json --jq '.[].name'

			# Sort by most recently published:
			$ circleci orb list --sort latest
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching orbs...")
			orbs, err := client.ListOrbs(namespace, private, sort)
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
					fmt.Fprintln(f.IOStreams.Out, "No orbs found.")
				}
				return nil
			}

			if opts.Plain {
				for _, o := range orbs {
					latest := "-"
					if len(o.Versions) > 0 {
						latest = o.Versions[0].Version
					}
					fmt.Fprintf(f.IOStreams.Out, "%s\t%s\n", o.Name, latest)
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

	cmd.Flags().StringVar(&namespace, "namespace", "", "Filter to a specific namespace")
	cmd.Flags().BoolVar(&private, "private", false, "Include private orbs accessible to the token")
	cmd.Flags().StringVar(&sort, "sort", "popularity", `Sort order: popularity, latest, alphabetical (default: "popularity")`)
	output.AddFlags(cmd, &opts, &apiclient.Orb{})
	return cmd
}
