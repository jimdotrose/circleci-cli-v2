package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
)

func NewCmdVersion(f *cmdutil.Factory, version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the circleci CLI version",
		Long:  "Print the version of the circleci CLI binary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(f.IOStreams.Out, "circleci version %s\n", version)
			return nil
		},
	}
}
