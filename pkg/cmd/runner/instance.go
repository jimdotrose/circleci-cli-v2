package runner

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdInstance returns the `circleci runner instance` command group.
func NewCmdInstance(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance <command>",
		Short: "Inspect runner instances",
		Long: heredoc.Doc(`
			Commands for inspecting registered self-hosted runner agent instances.

			Runner instances are the individual agents registered under a
			resource class. Use these commands to monitor connectivity,
			version, and last-used status.
		`),
	}

	cmd.AddCommand(NewCmdInstanceList(f))
	return cmd
}

// NewCmdInstanceList returns `circleci runner instance list`.
func NewCmdInstanceList(f *cmdutil.Factory) *cobra.Command {
	var resourceClass string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List runner instances for a resource class",
		Long: heredoc.Doc(`
			List all registered self-hosted runner agent instances for a
			resource class.

			Shows hostname, version, last connection time, and last-used time
			for each agent. Use --json to get full metadata.
		`),
		Example: heredoc.Doc(`
			# List instances for a resource class:
			$ circleci runner instance list --resource-class myorg/my-runner

			# List as JSON:
			$ circleci runner instance list --resource-class myorg/my-runner --json

			# Find stale instances:
			$ circleci runner instance list --resource-class myorg/my-runner \
			    --jq '[.[] | select(.last_used < "2024-01-01")]'
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resourceClass == "" {
				return cierrors.New("MISSING_ARG", "--resource-class is required",
					"Provide the resource class to list instances for.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching instances...")
			instances, err := client.ListRunnerInstances(resourceClass)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, instances); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(instances) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintln(f.IOStreams.Out, "No runner instances found.")
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-30s  %-10s  %-20s  %s\n", "HOSTNAME", "VERSION", "LAST CONNECTED", "NAME")
			for _, i := range instances {
				fmt.Fprintf(f.IOStreams.Out, "%-30s  %-10s  %-20s  %s\n",
					i.Hostname, i.Version, i.LastConnected.Format("2006-01-02 15:04"), i.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&resourceClass, "resource-class", "", "Resource class to list instances for")
	output.AddFlags(cmd, &opts, &apiclient.RunnerInstance{})
	return cmd
}
