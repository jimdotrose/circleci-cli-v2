package config

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdConfig returns the `circleci config` command group.
func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage CircleCI configuration files",
		Long: heredoc.Doc(`
			Commands for working with CircleCI configuration files.

			Validate, process, pack, and generate .circleci/config.yml files
			locally or using the CircleCI compile API.
		`),
	}

	cmd.AddCommand(NewCmdValidate(f))
	cmd.AddCommand(NewCmdProcess(f))
	cmd.AddCommand(NewCmdPack(f))
	cmd.AddCommand(NewCmdGenerate(f))
	return cmd
}
