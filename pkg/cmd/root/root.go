package root

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/auth"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/diagnostic"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/settings"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/version"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdRoot builds the root cobra.Command with all global flags, command
// groups, help topics, and subcommands wired to the provided Factory.
func NewCmdRoot(f *cmdutil.Factory, buildVersion string) *cobra.Command {
	// applyGlobalFlags reads persistent flag values and propagates them to
	// IOStreams. Called from both PersistentPreRunE (normal execution) and the
	// custom HelpFunc — Cobra short-circuits PersistentPreRunE when --help is
	// passed, so the help path needs its own application.
	applyGlobalFlags := func(c *cobra.Command) {
		if nc, err := c.Root().PersistentFlags().GetBool("no-color"); err == nil && nc {
			f.IOStreams.SetColorEnabled(false)
		}
		if np, err := c.Root().PersistentFlags().GetBool("no-prompt"); err == nil && np {
			f.IOStreams.SetInteractive(false)
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
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			applyGlobalFlags(cmd)
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

	settingsCmd := settings.NewCmdSettings(f)
	settingsCmd.GroupID = "developer"
	cmd.AddCommand(settingsCmd)

	diagCmd := diagnostic.NewCmdDiagnostic(f)
	diagCmd.GroupID = "developer"
	cmd.AddCommand(diagCmd)

	versionCmd := version.NewCmdVersion(f, buildVersion)
	versionCmd.GroupID = "developer"
	cmd.AddCommand(versionCmd)

	// --version / -v flag at root level.
	cmd.Version = buildVersion
	cmd.InitDefaultVersionFlag()

	// Shell completion (bash/zsh/fish/powershell).
	cmd.InitDefaultCompletionCmd()

	// ── Help topics ───────────────────────────────────────────────────────────
	cmd.AddCommand(newHelpTopicCmd("environment", environmentHelpTitle, environmentHelpBody))
	cmd.AddCommand(newHelpTopicCmd("exit-codes", exitCodesHelpTitle, exitCodesHelpBody))
	cmd.AddCommand(newHelpTopicCmd("formatting", formattingHelpTitle, formattingHelpBody))

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
