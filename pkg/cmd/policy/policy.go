package policy

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdPolicy returns the `circleci policy` command group.
func NewCmdPolicy(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy <command>",
		Short: "Manage CircleCI config policies",
		Long: heredoc.Doc(`
			Commands for managing CircleCI config policies.

			Config policies use Open Policy Agent (OPA) Rego rules to enforce
			organizational standards on pipeline configuration. Push a policy
			bundle, view decision logs, and evaluate configs locally.
		`),
	}

	cmd.AddCommand(NewCmdPush(f))
	cmd.AddCommand(NewCmdDiff(f))
	cmd.AddCommand(NewCmdFetch(f))
	cmd.AddCommand(NewCmdLogs(f))
	cmd.AddCommand(NewCmdDecide(f))
	cmd.AddCommand(NewCmdEval(f))
	cmd.AddCommand(NewCmdSettings(f))
	cmd.AddCommand(NewCmdTest(f))
	return cmd
}

// NewCmdPush returns `circleci policy push`.
func NewCmdPush(f *cmdutil.Factory) *cobra.Command {
	var ownerID string
	var noPrompt bool

	cmd := &cobra.Command{
		Use:   "push <bundle-dir>",
		Short: "Push a policy bundle to CircleCI",
		Long: heredoc.Doc(`
			Upload a directory of OPA Rego policy files to CircleCI as the
			active policy bundle for an organization.

			All .rego files in <bundle-dir> are uploaded. The policy bundle
			immediately becomes active for all new pipeline evaluations.
			Use --dry-run (via 'policy diff') to preview changes first.
		`),
		Example: heredoc.Doc(`
			# Push a policy bundle:
			$ circleci policy push ./policies --owner-id <org-id>

			# Push from a CI script:
			$ circleci policy push ./policies --owner-id $ORG_ID --no-prompt

			# Preview changes before pushing:
			$ circleci policy diff ./policies --owner-id $ORG_ID
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization ID to push the policy to.", cierrors.ExitBadArguments)
			}

			if !noPrompt && f.IOStreams.IsInteractive {
				fmt.Fprintf(f.IOStreams.ErrOut, "This will replace the active policy bundle for org %s.\n", ownerID)
				fmt.Fprint(f.IOStreams.ErrOut, "Continue? [y/N] ")
				var answer string
				fmt.Fscan(f.IOStreams.In, &answer)
				if answer != "y" && answer != "Y" {
					return cierrors.New("CANCELLED", "Push cancelled", "No changes made.", cierrors.ExitCancelled)
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Pushing policy bundle...")
			result, err := client.PolicyBundlePush(ownerID, args[0], false)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				out, _ := json.MarshalIndent(result, "", "  ")
				fmt.Fprintln(f.IOStreams.Out, string(out))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization ID")
	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "Skip confirmation prompt")
	return cmd
}

// NewCmdDiff returns `circleci policy diff`.
func NewCmdDiff(f *cmdutil.Factory) *cobra.Command {
	var ownerID string

	cmd := &cobra.Command{
		Use:   "diff <bundle-dir>",
		Short: "Preview changes before pushing a policy bundle",
		Long: heredoc.Doc(`
			Show what would change by pushing the policy bundle in <bundle-dir>
			without actually applying it (dry run).

			The output shows added, modified, and deleted policy documents
			compared to the currently active bundle.
		`),
		Example: heredoc.Doc(`
			# Preview policy changes:
			$ circleci policy diff ./policies --owner-id <org-id>

			# Use in CI before merging:
			$ circleci policy diff ./policies --owner-id $ORG_ID

			# Combine with push in a pipeline:
			$ circleci policy diff ./policies --owner-id $ORG_ID && \
			    circleci policy push ./policies --owner-id $ORG_ID --no-prompt
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization ID.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Computing diff...")
			result, err := client.PolicyDiff(ownerID, args[0])
			stop()
			if err != nil {
				return err
			}

			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(f.IOStreams.Out, string(out))
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization ID")
	return cmd
}

// NewCmdFetch returns `circleci policy fetch`.
func NewCmdFetch(f *cmdutil.Factory) *cobra.Command {
	var ownerID string
	var policyName string

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch the active policy bundle",
		Long: heredoc.Doc(`
			Fetch the currently active policy bundle for an organization.

			Without --policy-name, returns all policy documents in the bundle.
			With --policy-name, returns only the named document.
		`),
		Example: heredoc.Doc(`
			# Fetch the entire active bundle:
			$ circleci policy fetch --owner-id <org-id>

			# Fetch a specific policy document:
			$ circleci policy fetch --owner-id <org-id> --policy-name my_policy

			# Save the bundle to disk:
			$ circleci policy fetch --owner-id $ORG_ID > bundle.json
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization ID.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching policy bundle...")
			var result map[string]interface{}
			var fetchErr error
			if policyName != "" {
				result, fetchErr = client.GetPolicyDocument(ownerID, policyName)
			} else {
				result, fetchErr = client.GetPolicyBundle(ownerID)
			}
			stop()
			if fetchErr != nil {
				return fetchErr
			}

			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(f.IOStreams.Out, string(out))
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization ID")
	cmd.Flags().StringVar(&policyName, "policy-name", "", "Fetch a specific policy document by name")
	return cmd
}

// NewCmdLogs returns `circleci policy logs`.
func NewCmdLogs(f *cmdutil.Factory) *cobra.Command {
	var ownerID string
	var after, before string
	var limit int
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "List policy decision logs",
		Long: heredoc.Doc(`
			List historical policy decision logs for an organization.

			Each log entry records the config that was evaluated, the decision
			result (pass/fail), and any policy violations. Use --after and
			--before to filter by time range.
		`),
		Example: heredoc.Doc(`
			# List recent decision logs:
			$ circleci policy logs --owner-id <org-id>

			# Filter to a time range:
			$ circleci policy logs --owner-id <org-id> \
			    --after 2024-01-01 --before 2024-02-01

			# List as JSON:
			$ circleci policy logs --owner-id <org-id> --json
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization ID.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching decision logs...")
			logs, _, err := client.ListPolicyLogs(ownerID, after, before, limit, "")
			stop()
			if err != nil {
				return err
			}

			if err := opts.Write(f.IOStreams.Out, logs); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			if len(logs) == 0 {
				if !f.IOStreams.Quiet {
					fmt.Fprintln(f.IOStreams.Out, "No decision logs found.")
				}
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "%-36s  %-10s  %s\n", "ID", "STATUS", "CREATED")
			for _, l := range logs {
				fmt.Fprintf(f.IOStreams.Out, "%-36s  %-10s  %s\n",
					l.ID, l.Decision.Status, l.CreatedAt.Format("2006-01-02 15:04"))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization ID")
	cmd.Flags().StringVar(&after, "after", "", "Filter logs after this timestamp (RFC3339)")
	cmd.Flags().StringVar(&before, "before", "", "Filter logs before this timestamp (RFC3339)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of logs to return")
	output.AddFlags(cmd, &opts, &apiclient.PolicyLog{})
	return cmd
}

// NewCmdDecide returns `circleci policy decide`.
func NewCmdDecide(f *cmdutil.Factory) *cobra.Command {
	var ownerID string
	var pipelineParams string
	var metaProjectID string

	cmd := &cobra.Command{
		Use:   "decide <config-file>",
		Short: "Evaluate a config against the active policy bundle",
		Long: heredoc.Doc(`
			Evaluate a CircleCI config file against the organization's active
			policy bundle and show the decision result.

			Exits 0 when the config passes all policies. Exits 7 when any
			hard-failure policy is violated. Soft failures are reported but
			do not change the exit code.
		`),
		Example: heredoc.Doc(`
			# Evaluate the default config:
			$ circleci policy decide .circleci/config.yml --owner-id <org-id>

			# Evaluate with pipeline parameters:
			$ circleci policy decide .circleci/config.yml --owner-id <org-id> \
			    --pipeline-params '{"deploy_env":"staging"}'

			# Use in CI to block merges:
			$ circleci policy decide .circleci/config.yml --owner-id $ORG_ID || exit 1
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization ID.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Evaluating policy...")
			decision, err := client.PolicyDecide(ownerID, args[0], pipelineParams, metaProjectID)
			stop()
			if err != nil {
				return err
			}

			printDecision(f, decision)

			if decision.Status == "HARD_FAIL" || len(decision.HardFailures) > 0 {
				return cierrors.New("POLICY_VIOLATION", "Policy hard failure",
					"The config violates one or more hard-failure policies.",
					cierrors.ExitValidationFail)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization ID")
	cmd.Flags().StringVar(&pipelineParams, "pipeline-params", "", "Pipeline parameters as JSON")
	cmd.Flags().StringVar(&metaProjectID, "meta-project-id", "", "Project ID for metadata context")
	return cmd
}

// NewCmdEval returns `circleci policy eval`.
func NewCmdEval(f *cmdutil.Factory) *cobra.Command {
	var ownerID string
	var bundlePath string

	cmd := &cobra.Command{
		Use:   "eval <config-file>",
		Short: "Evaluate a config against a local policy bundle",
		Long: heredoc.Doc(`
			Evaluate a CircleCI config file against a local OPA bundle file
			without making an API call. Useful for testing policy changes
			before pushing.
		`),
		Example: heredoc.Doc(`
			# Evaluate against a local bundle:
			$ circleci policy eval .circleci/config.yml \
			    --owner-id <org-id> --bundle ./policies/bundle.tar.gz

			# Use in a pre-commit hook:
			$ circleci policy eval .circleci/config.yml \
			    --owner-id $ORG_ID --bundle ./policies/bundle.tar.gz

			# Combine with policy test in CI:
			$ circleci policy test --owner-id $ORG_ID ./policies && \
			    circleci policy eval .circleci/config.yml --owner-id $ORG_ID \
			    --bundle ./policies/bundle.tar.gz
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization ID.", cierrors.ExitBadArguments)
			}
			if bundlePath == "" {
				return cierrors.New("MISSING_ARG", "--bundle is required",
					"Provide a path to the local policy bundle.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Evaluating policy locally...")
			decision, err := client.PolicyEval(ownerID, bundlePath, args[0])
			stop()
			if err != nil {
				return err
			}

			printDecision(f, decision)

			if decision.Status == "HARD_FAIL" || len(decision.HardFailures) > 0 {
				return cierrors.New("POLICY_VIOLATION", "Policy hard failure",
					"The config violates one or more hard-failure policies.",
					cierrors.ExitValidationFail)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization ID")
	cmd.Flags().StringVar(&bundlePath, "bundle", "", "Path to local policy bundle file")
	return cmd
}

// NewCmdSettings returns `circleci policy settings`.
func NewCmdSettings(f *cmdutil.Factory) *cobra.Command {
	var ownerID string
	var enabled string

	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Get or set policy enforcement settings",
		Long: heredoc.Doc(`
			Get or update the policy enforcement settings for an organization.

			Without --enabled, shows the current settings.
			With --enabled true|false, updates the enforcement state.

			When enabled=true, all pipelines must pass the active policy bundle
			or they will be blocked from running.
		`),
		Example: heredoc.Doc(`
			# Show current settings:
			$ circleci policy settings --owner-id <org-id>

			# Enable policy enforcement:
			$ circleci policy settings --owner-id <org-id> --enabled true

			# Disable policy enforcement:
			$ circleci policy settings --owner-id <org-id> --enabled false
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization ID.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if enabled == "" {
				// GET
				stop := f.IOStreams.StartSpinner("Fetching settings...")
				s, err := client.GetPolicySettings(ownerID)
				stop()
				if err != nil {
					return err
				}
				fmt.Fprintf(f.IOStreams.Out, "enabled: %v\n", s.Enabled)
				return nil
			}

			// SET
			val := enabled == "true"
			stop := f.IOStreams.StartSpinner("Updating settings...")
			s, err := client.SetPolicySettings(ownerID, val)
			stop()
			if err != nil {
				return err
			}
			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Policy enforcement enabled: %v\n", s.Enabled)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization ID")
	cmd.Flags().StringVar(&enabled, "enabled", "", "Set enforcement: true or false")
	return cmd
}

// NewCmdTest returns `circleci policy test`.
func NewCmdTest(f *cmdutil.Factory) *cobra.Command {
	var ownerID string

	cmd := &cobra.Command{
		Use:   "test <bundle-dir>",
		Short: "Run OPA tests against a policy bundle",
		Long: heredoc.Doc(`
			Run OPA unit tests defined alongside the policy bundle in <bundle-dir>.

			Test files must match the pattern *_test.rego. Exits 0 when all
			tests pass, non-zero otherwise. Use this in CI to gate pushes.
		`),
		Example: heredoc.Doc(`
			# Run tests in a bundle directory:
			$ circleci policy test ./policies --owner-id <org-id>

			# Run tests and block push on failure:
			$ circleci policy test ./policies --owner-id $ORG_ID && \
			    circleci policy push ./policies --owner-id $ORG_ID --no-prompt

			# Run tests in CI:
			$ circleci policy test ./policies --owner-id $ORG_ID || exit 1
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ownerID == "" {
				return cierrors.New("MISSING_ARG", "--owner-id is required",
					"Provide the organization ID.", cierrors.ExitBadArguments)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Running policy tests...")
			result, err := client.PolicyTest(ownerID, args[0])
			stop()
			if err != nil {
				return err
			}

			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(f.IOStreams.Out, string(out))

			// Check for failures in the result.
			if failed, ok := result["failed"].(bool); ok && failed {
				return cierrors.New("TEST_FAILED", "Policy tests failed",
					"One or more OPA tests failed.", cierrors.ExitValidationFail)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Organization ID")
	return cmd
}

// printDecision prints a human-readable summary of a policy decision.
func printDecision(f *cmdutil.Factory, d *apiclient.PolicyDecision) {
	if d == nil {
		return
	}
	fmt.Fprintf(f.IOStreams.Out, "Status: %s\n", d.Status)
	if len(d.HardFailures) > 0 {
		fmt.Fprintln(f.IOStreams.Out, "Hard Failures:")
		for _, v := range d.HardFailures {
			fmt.Fprintf(f.IOStreams.Out, "  [%s] %s\n", v.Rule, v.Reason)
		}
	}
	if len(d.SoftFailures) > 0 {
		fmt.Fprintln(f.IOStreams.Out, "Soft Failures:")
		for _, v := range d.SoftFailures {
			fmt.Fprintf(f.IOStreams.Out, "  [%s] %s\n", v.Rule, v.Reason)
		}
	}
	_ = os.Stderr // keep os import used
}
