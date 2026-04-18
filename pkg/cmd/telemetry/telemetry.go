package telemetry

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
	"github.com/CircleCI-Public/circleci-cli/pkg/telemetry"
)

// NewCmdTelemetry returns the `circleci telemetry` command group.
func NewCmdTelemetry(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "telemetry <command>",
		Short: "Manage anonymous usage telemetry",
		Long: heredoc.Doc(`
			Commands for managing CircleCI CLI telemetry.

			The CLI collects anonymous usage data to help improve the product.
			No personal information, file paths, flag values, or tokens are
			ever collected.

			Telemetry is automatically disabled when any of the following are set:
			  CIRCLECI_NO_TELEMETRY, NO_ANALYTICS, DO_NOT_TRACK, CI=true
		`),
	}

	cmd.AddCommand(NewCmdStatus(f))
	cmd.AddCommand(NewCmdEnable(f))
	cmd.AddCommand(NewCmdDisable(f))
	return cmd
}

// NewCmdStatus returns the `circleci telemetry status` command.
func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether telemetry is enabled",
		Long: heredoc.Doc(`
			Show the current telemetry collection state.

			Reports the effective state after evaluating environment variables
			and the stored config setting. Environment variables always take
			precedence over the stored setting.
		`),
		Example: heredoc.Doc(`
			# Check telemetry state:
			$ circleci telemetry status

			# Check in a script:
			$ circleci telemetry status --quiet && echo "telemetry on"

			# Show the configured setting:
			$ circleci settings get telemetry
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			val, _ := cfg.Get("telemetry")
			enabled := telemetry.Enabled(val)

			if enabled {
				fmt.Fprintln(f.IOStreams.Out, "Telemetry is enabled.")
			} else {
				fmt.Fprintln(f.IOStreams.Out, "Telemetry is disabled.")
			}
			return nil
		},
	}
}

// NewCmdEnable returns the `circleci telemetry enable` command.
func NewCmdEnable(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable anonymous usage telemetry",
		Long: heredoc.Doc(`
			Enable anonymous usage telemetry collection.

			Sets telemetry=true in ~/.circleci/cli.yml. This setting is
			overridden by environment variables — if CIRCLECI_NO_TELEMETRY,
			NO_ANALYTICS, DO_NOT_TRACK, or CI=true are set, telemetry will
			remain disabled regardless of this setting.
		`),
		Example: heredoc.Doc(`
			# Enable telemetry:
			$ circleci telemetry enable

			# Enable and verify:
			$ circleci telemetry enable && circleci telemetry status

			# Equivalent settings command:
			$ circleci settings set telemetry true
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			if err := cfg.Set("telemetry", "true"); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return cierrors.New("SAVE_ERROR", "Could not save config",
					err.Error(), cierrors.ExitGeneralError)
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintln(f.IOStreams.Out, "✓ Telemetry enabled.")
			}
			return nil
		},
	}
}

// NewCmdDisable returns the `circleci telemetry disable` command.
func NewCmdDisable(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable anonymous usage telemetry",
		Long: heredoc.Doc(`
			Disable anonymous usage telemetry collection.

			Sets telemetry=false in ~/.circleci/cli.yml. You can also disable
			telemetry permanently via environment variables:

			  export CIRCLECI_NO_TELEMETRY=1   # or NO_ANALYTICS, DO_NOT_TRACK
		`),
		Example: heredoc.Doc(`
			# Disable telemetry:
			$ circleci telemetry disable

			# Disable and verify:
			$ circleci telemetry disable && circleci telemetry status

			# Equivalent settings command:
			$ circleci settings set telemetry false
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return cierrors.New("CONFIG_ERROR", "Could not load config",
					err.Error(), cierrors.ExitGeneralError)
			}

			if err := cfg.Set("telemetry", "false"); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return cierrors.New("SAVE_ERROR", "Could not save config",
					err.Error(), cierrors.ExitGeneralError)
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintln(f.IOStreams.Out, "✓ Telemetry disabled.")
			}
			return nil
		},
	}
}
