package project

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdEnv returns the `circleci project env` command group.
func NewCmdEnv(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env <command>",
		Short: "Manage project environment variables",
		Long: heredoc.Doc(`
			Commands for managing CircleCI project environment variables.

			Environment variable values are never returned by the API — only
			names are readable. Use 'set' to create or update a variable.
		`),
	}

	cmd.AddCommand(NewCmdEnvList(f))
	cmd.AddCommand(NewCmdEnvGet(f))
	cmd.AddCommand(NewCmdEnvSet(f))
	cmd.AddCommand(NewCmdEnvDelete(f))
	return cmd
}

// NewCmdEnvList returns the `circleci project env list` command.
func NewCmdEnvList(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "list <project-slug>",
		Short: "List project environment variables",
		Long: heredoc.Doc(`
			List all environment variables for a CircleCI project.

			Variable values are always redacted (shown as "xxxx") — only
			names are returned by the API. Use 'set' to update a value.
		`),
		Example: heredoc.Doc(`
			# List env vars for a project:
			$ circleci project env list github/myorg/myrepo

			# List as JSON:
			$ circleci project env list github/myorg/myrepo --json

			# Extract names only:
			$ circleci project env list github/myorg/myrepo --jq '.[].name'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching environment variables...")
			vars, err := client.ListEnvVars(args[0])
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, vars); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(vars) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintln(f.IOStreams.Out, "No environment variables set.")
				}
				return nil
			}
			for _, v := range vars {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", v.Name)
			}
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.EnvVar{})
	return cmd
}

// NewCmdEnvGet returns the `circleci project env get` command.
func NewCmdEnvGet(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "get <project-slug> <name>",
		Short: "Get a project environment variable (name only; value is redacted)",
		Long: heredoc.Doc(`
			Get metadata for a project environment variable.

			The API never returns variable values — the value field will always
			be "xxxx". To check whether a variable is set, use this command and
			look for a successful response.
		`),
		Example: heredoc.Doc(`
			# Check whether AWS_KEY is set:
			$ circleci project env get github/myorg/myrepo AWS_KEY

			# Get as JSON:
			$ circleci project env get github/myorg/myrepo AWS_KEY --json

			# Use in a script:
			$ circleci project env get github/myorg/myrepo MY_VAR && echo "is set"
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching variable...")
			ev, err := client.GetEnvVar(args[0], args[1])
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, ev); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%s=%s\n", ev.Name, ev.Value)
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.EnvVar{})
	return cmd
}

// NewCmdEnvSet returns the `circleci project env set` command.
func NewCmdEnvSet(f *cmdutil.Factory) *cobra.Command {
	var value string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "set <project-slug> <name>",
		Short: "Set a project environment variable",
		Long: heredoc.Doc(`
			Create or update a CircleCI project environment variable.

			The variable is available in all subsequent pipeline runs for the
			project. Existing values are overwritten silently.

			Provide the value with --value. Pass --value - to read the value
			from stdin, which avoids shell history exposure for sensitive values.

			Use --dry-run to print what would be set without calling the API.
		`),
		Example: heredoc.Doc(`
			# Set an environment variable:
			$ circleci project env set github/myorg/myrepo AWS_KEY --value AKIAIOSFODNN7

			# Read value from stdin (avoids shell history exposure):
			$ echo "$MY_SECRET" | circleci project env set github/myorg/myrepo MY_SECRET --value -

			# Preview without setting:
			$ circleci project env set github/myorg/myrepo DB_URL --value "$DB_URL" --dry-run
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, name := args[0], args[1]

			if value == "" {
				return cierrors.New("MISSING_ARG", "--value is required",
					"Provide the variable value with --value, or use --value - to read from stdin.",
					cierrors.ExitBadArguments)
			}

			// --value - reads from stdin.
			if value == "-" {
				buf := make([]byte, 65536)
				n, err := f.IOStreams.In.Read(buf)
				if err != nil && n == 0 {
					return cierrors.New("MISSING_VALUE", "No value on stdin",
						"Pipe the value to stdin when using --value -.", cierrors.ExitBadArguments)
				}
				value = string(buf[:n])
			}

			if dryRun {
				fmt.Fprintf(f.IOStreams.Out, "Would set %s on project %s\n", name, slug)
				return nil
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Setting variable...")
			err = client.SetEnvVar(slug, name, value)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Set %s for project %s\n", name, slug)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&value, "value", "", "Variable value (use - to read from stdin)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Print what would be set without making API call")
	return cmd
}

// NewCmdEnvDelete returns the `circleci project env delete` command.
func NewCmdEnvDelete(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <project-slug> <name>",
		Short: "Delete a project environment variable",
		Long: heredoc.Doc(`
			Delete a CircleCI project environment variable.

			This action is irreversible. Use --force to skip the confirmation
			prompt in non-interactive mode.
		`),
		Example: heredoc.Doc(`
			# Delete a variable (prompts for confirmation):
			$ circleci project env delete github/myorg/myrepo OLD_KEY

			# Delete without prompting:
			$ circleci project env delete github/myorg/myrepo OLD_KEY --force

			# Delete in a CI script:
			$ circleci project env delete github/myorg/myrepo OLD_KEY --force --no-prompt
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, name := args[0], args[1]

			if !force && !f.IOStreams.IsInteractive {
				return cierrors.New(
					"CONFIRMATION_REQUIRED",
					"Confirmation required",
					"Pass --force to delete an environment variable non-interactively.",
					cierrors.ExitBadArguments,
				)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Deleting variable...")
			err = client.DeleteEnvVar(slug, name)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Deleted %s from project %s\n", name, slug)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
