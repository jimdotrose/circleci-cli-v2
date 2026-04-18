package version

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

func NewCmdVersion(f *cmdutil.Factory, version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the circleci CLI version",
		Long:  "Print the version of the circleci CLI binary.",
		Example: heredoc.Doc(`
			# Print the current version:
			$ circleci version

			# Via root flag:
			$ circleci --version

			# Check version in a script:
			$ circleci version --json | jq -r .version
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(f.IOStreams.Out, "circleci version %s\n", version)
			return nil
		},
	}
}
