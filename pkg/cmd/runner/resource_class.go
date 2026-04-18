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

// NewCmdResourceClass returns the `circleci runner resource-class` group.
func NewCmdResourceClass(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource-class <command>",
		Short: "Manage runner resource classes",
		Long: heredoc.Doc(`
			Commands for managing self-hosted runner resource classes.

			Resource classes define the type of runner that executes jobs.
			They are namespaced under your organization and referenced in
			your config.yml executor definition.
		`),
	}

	cmd.AddCommand(NewCmdResourceClassList(f))
	cmd.AddCommand(NewCmdResourceClassCreate(f))
	cmd.AddCommand(NewCmdResourceClassDelete(f))
	return cmd
}

// NewCmdResourceClassList returns `circleci runner resource-class list`.
func NewCmdResourceClassList(f *cmdutil.Factory) *cobra.Command {
	var namespace string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List runner resource classes",
		Long: heredoc.Doc(`
			List self-hosted runner resource classes for a namespace.

			Use --namespace to filter to a specific organization namespace.
			Without --namespace, all resource classes visible to the token
			are returned.
		`),
		Example: heredoc.Doc(`
			# List resource classes for a namespace:
			$ circleci runner resource-class list --namespace myorg

			# List as JSON:
			$ circleci runner resource-class list --namespace myorg --json

			# Also available as a top-level alias:
			$ circleci resource-class list --namespace myorg
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching resource classes...")
			classes, err := client.ListRunnerResourceClasses(namespace)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, classes); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(classes) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintln(f.IOStreams.Out, "No resource classes found.")
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-50s  %s\n", "RESOURCE CLASS", "DESCRIPTION")
			for _, rc := range classes {
				fmt.Fprintf(f.IOStreams.Out, "%-50s  %s\n", rc.ResourceClass, rc.Description)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace to filter resource classes")
	output.AddFlags(cmd, &opts, &apiclient.RunnerResourceClass{})
	return cmd
}

// NewCmdResourceClassCreate returns `circleci runner resource-class create`.
func NewCmdResourceClassCreate(f *cmdutil.Factory) *cobra.Command {
	var description string
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "create <resource-class>",
		Short: "Create a runner resource class",
		Long: heredoc.Doc(`
			Create a new self-hosted runner resource class.

			The resource class name must be fully qualified with the namespace:
			  <namespace>/<name>

			For example: myorg/my-runner-class
		`),
		Example: heredoc.Doc(`
			# Create a resource class:
			$ circleci runner resource-class create myorg/my-runner

			# Create with a description:
			$ circleci runner resource-class create myorg/my-runner \
			    --description "Linux amd64 runner"

			# Create and output as JSON:
			$ circleci runner resource-class create myorg/my-runner --json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Creating resource class...")
			rc, err := client.CreateRunnerResourceClass(args[0], description)
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, rc); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Created resource class %s\n", rc.ResourceClass)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Human-readable description")
	output.AddFlags(cmd, &opts, &apiclient.RunnerResourceClass{})
	return cmd
}

// NewCmdResourceClassDelete returns `circleci runner resource-class delete`.
func NewCmdResourceClassDelete(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <resource-class>",
		Short: "Delete a runner resource class",
		Long: heredoc.Doc(`
			Delete a self-hosted runner resource class.

			This action is irreversible. Any runner agents registered under
			this resource class will no longer be able to claim jobs.
			Use --force to skip the confirmation prompt.
		`),
		Example: heredoc.Doc(`
			# Delete a resource class (prompts for confirmation):
			$ circleci runner resource-class delete myorg/old-runner

			# Delete without prompting:
			$ circleci runner resource-class delete myorg/old-runner --force

			# Delete in a CI script:
			$ circleci runner resource-class delete myorg/old-runner --force --no-prompt
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force && !f.IOStreams.IsInteractive {
				return cierrors.New(
					"CONFIRMATION_REQUIRED",
					"Confirmation required",
					"Pass --force to delete a resource class non-interactively.",
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Deleting resource class...")
			err = client.DeleteRunnerResourceClass(args[0])
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Deleted resource class %s\n", args[0])
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
