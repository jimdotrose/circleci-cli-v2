package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

const starterConfig = `version: 2.1

# See https://circleci.com/docs/configuration-reference/
jobs:
  build:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      - run:
          name: Say Hello
          command: echo "Hello, World!"

workflows:
  say-hello:
    jobs:
      - build
`

// NewCmdGenerate returns the `circleci config generate` command.
func NewCmdGenerate(f *cmdutil.Factory) *cobra.Command {
	var force bool
	var outPath string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a starter CircleCI config file",
		Long: heredoc.Doc(`
			Generate a minimal CircleCI configuration file to get started.

			Writes a basic config.yml with a single job and workflow that you
			can use as a starting point. By default writes to
			.circleci/config.yml in the current directory.

			Use --force to overwrite an existing file.
		`),
		Example: heredoc.Doc(`
			# Generate .circleci/config.yml in the current directory:
			$ circleci config generate

			# Generate at a custom path:
			$ circleci config generate --out my-config.yml

			# Overwrite an existing config:
			$ circleci config generate --force
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if outPath == "" {
				outPath = filepath.Join(".circleci", "config.yml")
			}

			if !force {
				if _, err := os.Stat(outPath); err == nil {
					return cierrors.New(
						"FILE_EXISTS",
						"Config file already exists",
						fmt.Sprintf("%q already exists. Use --force to overwrite.", outPath),
						cierrors.ExitBadArguments,
					)
				}
			}

			dir := filepath.Dir(outPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return cierrors.New(
					"DIR_CREATE_ERROR",
					"Could not create directory",
					fmt.Sprintf("Error creating %q: %v", dir, err),
					cierrors.ExitGeneralError,
				)
			}

			if err := os.WriteFile(outPath, []byte(starterConfig), 0644); err != nil {
				return cierrors.New(
					"WRITE_ERROR",
					"Could not write config file",
					fmt.Sprintf("Error writing %q: %v", outPath, err),
					cierrors.ExitGeneralError,
				)
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Created %s\n", outPath)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing config file")
	cmd.Flags().StringVar(&outPath, "out", "", "Output path (default: .circleci/config.yml)")
	return cmd
}
