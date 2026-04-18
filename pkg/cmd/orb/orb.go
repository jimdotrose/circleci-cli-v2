package orb

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

// NewCmdOrb returns the `circleci orb` command group.
func NewCmdOrb(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orb <command>",
		Short: "Manage and explore CircleCI orbs",
		Long: heredoc.Doc(`
			Commands for working with CircleCI orbs.

			Orbs are reusable packages of CircleCI configuration — jobs, commands,
			and executors — that can be shared across projects and published to
			the CircleCI Orb Registry.

			Typical workflow:
			  1. Author an orb YAML file.
			  2. Validate it with 'circleci orb validate'.
			  3. Publish a dev version with 'circleci orb publish ... @dev:label'.
			  4. Promote to a release with 'circleci orb promote'.
		`),
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdInfo(f))
	cmd.AddCommand(NewCmdValidate(f))
	cmd.AddCommand(NewCmdPublish(f))
	cmd.AddCommand(NewCmdPromote(f))
	cmd.AddCommand(NewCmdSearch(f))
	return cmd
}
