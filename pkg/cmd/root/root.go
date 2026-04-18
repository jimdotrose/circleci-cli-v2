package root

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	cmdapi "github.com/CircleCI-Public/circleci-cli/pkg/cmd/api"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/auth"
	cmdorb "github.com/CircleCI-Public/circleci-cli/pkg/cmd/orb"
	cmdconfig "github.com/CircleCI-Public/circleci-cli/pkg/cmd/config"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/context"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/diagnostic"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/job"
	cmdnamespace "github.com/CircleCI-Public/circleci-cli/pkg/cmd/namespace"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/pipeline"
	cmdpolicy "github.com/CircleCI-Public/circleci-cli/pkg/cmd/policy"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/project"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/runner"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/settings"
	cmdtelemetry "github.com/CircleCI-Public/circleci-cli/pkg/cmd/telemetry"
	cmdtrigger "github.com/CircleCI-Public/circleci-cli/pkg/cmd/trigger"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/version"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/workflow"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/telemetry"
)

// NewCmdRoot builds the root cobra.Command with all global flags, command
// groups, help topics, and subcommands wired to the provided Factory.
func NewCmdRoot(f *cmdutil.Factory, buildVersion string) *cobra.Command {
	// bindEnvToFlag sets a persistent flag from an env var only when the flag
	// has not already been set explicitly on the command line. This gives:
	//   CLI flag > env var > config file > built-in default
	bindEnvToFlag := func(c *cobra.Command, flagName, envVar string) {
		pf := c.Root().PersistentFlags()
		if pf.Changed(flagName) {
			return // explicit flag wins
		}
		if v := os.Getenv(envVar); v != "" {
			_ = pf.Set(flagName, v)
		}
	}

	// applyGlobalFlags reads persistent flag values and propagates them to
	// IOStreams. Called from both PersistentPreRunE (normal execution) and the
	// custom HelpFunc — Cobra short-circuits PersistentPreRunE when --help is
	// passed, so the help path needs its own application.
	applyGlobalFlags := func(c *cobra.Command) {
		// Apply env var → flag bindings before reading flag values.
		bindEnvToFlag(c, "token", "CIRCLECI_TOKEN")
		if !c.Root().PersistentFlags().Changed("token") {
			bindEnvToFlag(c, "token", "CIRCLECI_CLI_TOKEN")
		}
		bindEnvToFlag(c, "host", "CIRCLECI_HOST")
		bindEnvToFlag(c, "debug", "CIRCLECI_DEBUG")
		bindEnvToFlag(c, "no-color", "NO_COLOR")
		bindEnvToFlag(c, "no-color", "CIRCLECI_NO_COLOR")
		bindEnvToFlag(c, "no-prompt", "CI")
		bindEnvToFlag(c, "no-prompt", "CIRCLECI_NO_INTERACTIVE")
		bindEnvToFlag(c, "quiet", "CIRCLECI_QUIET")

		if nc, err := c.Root().PersistentFlags().GetBool("no-color"); err == nil && nc {
			f.IOStreams.SetColorEnabled(false)
		}
		if np, err := c.Root().PersistentFlags().GetBool("no-prompt"); err == nil && np {
			f.IOStreams.SetInteractive(false)
		}
		if q, err := c.Root().PersistentFlags().GetBool("quiet"); err == nil && q {
			f.IOStreams.Quiet = true
		}
		if dbg, err := c.Root().PersistentFlags().GetBool("debug"); err == nil && dbg {
			f.Debug = true
		}
		if fl := c.Flags().Lookup("json"); fl != nil && fl.Value.String() == "true" {
			f.JSONOutput = true
		}
		// --host flag overrides config + env var when explicitly set.
		if c.Root().PersistentFlags().Changed("host") {
			host, _ := c.Root().PersistentFlags().GetString("host")
			f.BaseURL = func() string { return host }
		}
	}

	cmd := &cobra.Command{
		Use:   "circleci",
		Short: "CircleCI CLI",
		Long: heredoc.Doc(`
			Work with CircleCI from the command line.

			Run 'circleci <command> --help' for usage of a specific command.
			Run 'circleci help <topic>' for detailed help on a topic:

			  circleci help environment    All supported environment variables
			  circleci help exit-codes     Documented exit codes
			  circleci help formatting     --json, --jq, and --template usage
			  circleci help api            Raw API access with 'circleci api'

			Report bugs and request features:
			  https://github.com/CircleCI-Public/circleci-cli/issues
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			applyGlobalFlags(cmd)
			if home, err := os.UserHomeDir(); err == nil {
				telemetry.ShowNoticeIfNeeded(f.IOStreams.ErrOut, home)
			}
			return nil
		},
	}

	// ── Command groups ────────────────────────────────────────────────────────
	cmd.AddGroup(&cobra.Group{ID: "core", Title: "Core Commands:"})
	cmd.AddGroup(&cobra.Group{ID: "developer", Title: "Developer Tools:"})

	// ── Global flags ──────────────────────────────────────────────────────────
	pf := cmd.PersistentFlags()
	pf.StringP("token", "T", "", "CircleCI API token (env: CIRCLECI_TOKEN)")
	pf.String("host", "https://circleci.com", "CircleCI host (env: CIRCLECI_HOST)")
	pf.BoolP("debug", "d", false, "Enable HTTP debug logging (env: CIRCLECI_DEBUG)")
	pf.Bool("no-color", false, "Disable ANSI color output (env: CIRCLECI_NO_COLOR, NO_COLOR)")
	pf.BoolP("quiet", "q", false, "Suppress progress and informational output")
	pf.Bool("no-prompt", false, "Disable interactive prompts (env: CIRCLECI_NO_INTERACTIVE, CI)")

	// ── Override help to apply global flags before rendering ──────────────────
	origHelp := cmd.HelpFunc()
	cmd.SetHelpFunc(func(helpCmd *cobra.Command, args []string) {
		applyGlobalFlags(helpCmd)
		origHelp(helpCmd, args)
	})

	// ── Subcommands ───────────────────────────────────────────────────────────
	authCmd := auth.NewCmdAuth(f)
	authCmd.GroupID = "core"
	cmd.AddCommand(authCmd)

	configCmd := cmdconfig.NewCmdConfig(f)
	configCmd.GroupID = "core"
	cmd.AddCommand(configCmd)

	contextCmd := context.NewCmdContext(f)
	contextCmd.GroupID = "core"
	cmd.AddCommand(contextCmd)

	pipelineCmd := pipeline.NewCmdPipeline(f)
	pipelineCmd.GroupID = "core"
	cmd.AddCommand(pipelineCmd)

	workflowCmd := workflow.NewCmdWorkflow(f)
	workflowCmd.GroupID = "core"
	cmd.AddCommand(workflowCmd)

	jobCmd := job.NewCmdJob(f)
	jobCmd.GroupID = "core"
	cmd.AddCommand(jobCmd)

	projectCmd := project.NewCmdProject(f)
	projectCmd.GroupID = "core"
	cmd.AddCommand(projectCmd)

	runnerCmd := runner.NewCmdRunner(f)
	runnerCmd.GroupID = "core"
	cmd.AddCommand(runnerCmd)

	orbCmd := cmdorb.NewCmdOrb(f)
	orbCmd.GroupID = "core"
	cmd.AddCommand(orbCmd)

	policyCmd := cmdpolicy.NewCmdPolicy(f)
	policyCmd.GroupID = "developer"
	cmd.AddCommand(policyCmd)

	telemetryCmd := cmdtelemetry.NewCmdTelemetry(f)
	telemetryCmd.GroupID = "developer"
	cmd.AddCommand(telemetryCmd)

	settingsCmd := settings.NewCmdSettings(f)
	settingsCmd.GroupID = "developer"
	cmd.AddCommand(settingsCmd)

	diagCmd := diagnostic.NewCmdDiagnostic(f)
	diagCmd.GroupID = "developer"
	cmd.AddCommand(diagCmd)

	versionCmd := version.NewCmdVersion(f, buildVersion)
	versionCmd.GroupID = "developer"
	cmd.AddCommand(versionCmd)

	apiCmd := cmdapi.NewCmdAPI(f)
	apiCmd.GroupID = "developer"
	cmd.AddCommand(apiCmd)

	triggerCmd := cmdtrigger.NewCmdTrigger(f)
	triggerCmd.GroupID = "core"
	cmd.AddCommand(triggerCmd)

	namespaceCmd := cmdnamespace.NewCmdNamespace(f)
	namespaceCmd.GroupID = "developer"
	cmd.AddCommand(namespaceCmd)

	// --version / -v flag at root level.
	cmd.Version = buildVersion
	cmd.InitDefaultVersionFlag()

	// Shell completion (bash/zsh/fish/powershell).
	cmd.InitDefaultCompletionCmd()

	// ── context set-secret alias: 2-level path for 3-level context secret set ──
	// context secret set is 3 levels deep; expose set-secret at 2 levels.
	contextCmd.AddCommand(&cobra.Command{
		Use:   "set-secret <context-id> <variable-name>",
		Short: "Alias for: circleci context secret set",
		Long: heredoc.Doc(`
			Alias for 'circleci context secret set'.

			Create or update an environment variable in a CircleCI context.
			The value is read from stdin (masked when interactive).
		`),
		Example: heredoc.Doc(`
			# Set a variable interactively:
			$ circleci context set-secret <context-id> MY_VAR

			# Set from stdin:
			$ echo "$SECRET" | circleci context set-secret <context-id> MY_VAR

			# Canonical form (same behavior):
			$ circleci context secret set <context-id> MY_VAR
		`),
		Hidden: true,
		Args:   cobra.ExactArgs(2),
		RunE:   context.NewCmdSecretSet(f).RunE,
	})

	// ── Deprecation shims ─────────────────────────────────────────────────────
	// context store-secret → context secret set
	contextCmd.AddCommand(&cobra.Command{
		Use:        "store-secret",
		Short:      "[deprecated] use: circleci context secret set",
		Hidden:     true,
		Deprecated: "use 'circleci context secret set' instead",
		Run:        func(cmd *cobra.Command, args []string) {},
	})
	// project environment-variable → project env
	projectCmd.AddCommand(&cobra.Command{
		Use:        "environment-variable",
		Short:      "[deprecated] use: circleci project env",
		Hidden:     true,
		Deprecated: "use 'circleci project env' instead",
		Run:        func(cmd *cobra.Command, args []string) {},
	})

	// ── Top-level alias: circleci resource-class → circleci runner resource-class ──
	cmd.AddCommand(newResourceClassAlias(f))

	// ── Help topics ───────────────────────────────────────────────────────────
	cmd.AddCommand(newHelpTopicCmd("environment", environmentHelpTitle, environmentHelpBody))
	cmd.AddCommand(newHelpTopicCmd("exit-codes", exitCodesHelpTitle, exitCodesHelpBody))
	cmd.AddCommand(newHelpTopicCmd("formatting", formattingHelpTitle, formattingHelpBody))
	cmd.AddCommand(newHelpTopicCmd("api", apiHelpTitle, apiHelpBody))

	return cmd
}

// newHelpTopicCmd returns a hidden cobra.Command that prints its Long text when
// the user runs `circleci help <topic>`.
func newHelpTopicCmd(name, title, body string) *cobra.Command {
	return &cobra.Command{
		Use:    name,
		Short:  title,
		Long:   body,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), cmd.Long)
		},
	}
}

// newResourceClassAlias returns a hidden top-level `resource-class` command
// that delegates to `runner resource-class`. Satisfies:
//
//	circleci resource-class list --namespace myorg
func newResourceClassAlias(f *cmdutil.Factory) *cobra.Command {
	runnerCmd := runner.NewCmdRunner(f)

	// Find the resource-class sub-command from the runner group.
	var rcCmd *cobra.Command
	for _, sub := range runnerCmd.Commands() {
		if sub.Use == "resource-class <command>" {
			rcCmd = sub
			break
		}
	}
	if rcCmd == nil {
		// Fallback — shouldn't happen.
		return &cobra.Command{
			Use:    "resource-class",
			Hidden: true,
		}
	}

	rcCmd.Use = "resource-class <command>"
	rcCmd.Short = "Alias for: circleci runner resource-class"
	rcCmd.Long = heredoc.Doc(`
		Alias for 'circleci runner resource-class'.

		Manage self-hosted runner resource classes. This top-level alias exists
		for convenience; the canonical path is 'circleci runner resource-class'.
	`)
	rcCmd.Hidden = true
	return rcCmd
}

