package root

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdRoot builds the root cobra.Command with all global flags wired to the
// Factory's IOStreams and a structured help layout.
func NewCmdRoot(f *cmdutil.Factory) *cobra.Command {
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
	}

	cmd := &cobra.Command{
		Use:   "circleci",
		Short: "CircleCI CLI",
		Long: heredoc.Doc(`
			Work with CircleCI from the command line.

			Run 'circleci --help' to see available commands.
			Run 'circleci help <topic>' for detailed help on a topic:

			  circleci help environment    All supported environment variables
			  circleci help exit-codes     Documented exit codes
		`),
		// SilenceUsage suppresses the usage block on errors — Cobra's default
		// behavior is to print the full usage on any error, which is noisy and
		// unhelpful for most errors. Commands that need usage hints can call
		// cmd.Usage() explicitly.
		SilenceUsage: true,
		// SilenceErrors lets main.go control error formatting via IOStreams.
		SilenceErrors: true,
		// PersistentPreRunE propagates global flag side-effects to every subcommand.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			applyGlobalFlags(cmd)
			return nil
		},
	}

	// ── Global flags ──────────────────────────────────────────────────────────
	pf := cmd.PersistentFlags()

	pf.StringP("token", "T", "", "CircleCI API token (env: CIRCLECI_TOKEN)")
	pf.String("host", "https://circleci.com", "CircleCI host (env: CIRCLECI_HOST)")
	pf.BoolP("debug", "d", false, "Enable HTTP debug logging (env: CIRCLECI_DEBUG)")
	pf.Bool("no-color", false, "Disable ANSI color output (env: CIRCLECI_NO_COLOR, NO_COLOR)")
	pf.BoolP("quiet", "q", false, "Suppress progress and informational output")
	pf.Bool("no-prompt", false, "Disable interactive prompts (env: CIRCLECI_NO_INTERACTIVE, CI)")

	// ── Override help to apply global flags before rendering ──────────────────
	// Cobra calls the HelpFunc directly for --help, bypassing PersistentPreRunE.
	origHelp := cmd.HelpFunc()
	cmd.SetHelpFunc(func(helpCmd *cobra.Command, args []string) {
		applyGlobalFlags(helpCmd)
		origHelp(helpCmd, args)
	})

	// ── Help topics ───────────────────────────────────────────────────────────
	cmd.AddCommand(newHelpTopicCmd("environment", environmentHelpTitle, environmentHelpBody))
	cmd.AddCommand(newHelpTopicCmd("exit-codes", exitCodesHelpTitle, exitCodesHelpBody))

	return cmd
}

// newHelpTopicCmd returns a hidden cobra.Command that renders a help topic
// when the user runs `circleci help <topic>`.
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
