package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdPack returns the `circleci config pack` command.
func NewCmdPack(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pack <dir>",
		Short: "Pack multiple YAML files into a single config",
		Long: heredoc.Doc(`
			Combine multiple YAML files from a directory into a single
			CircleCI config.yml.

			Files are merged by their top-level keys. Useful for splitting a
			large config into separate files for maintainability, then packing
			them before pushing to CircleCI.

			The directory must contain at least one .yml or .yaml file.
			Output is written to stdout.
		`),
		Example: heredoc.Doc(`
			# Pack all YAML files in .circleci/src/:
			$ circleci config pack .circleci/src

			# Pack and write to config.yml:
			$ circleci config pack .circleci/src > .circleci/config.yml

			# Pack then validate:
			$ circleci config pack .circleci/src | circleci config validate -
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]

			entries, err := os.ReadDir(dir)
			if err != nil {
				return cierrors.New(
					"DIR_NOT_FOUND",
					"Directory not found",
					fmt.Sprintf("Could not read directory %q: %v", dir, err),
					cierrors.ExitNotFound,
				)
			}

			merged := map[string]interface{}{}

			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
					continue
				}

				data, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					return cierrors.New(
						"READ_ERROR",
						"Could not read file",
						fmt.Sprintf("Error reading %s: %v", name, err),
						cierrors.ExitGeneralError,
					)
				}

				var doc map[string]interface{}
				if err := yaml.Unmarshal(data, &doc); err != nil {
					return cierrors.New(
						"YAML_PARSE_ERROR",
						"YAML parse error",
						fmt.Sprintf("Error parsing %s: %v", name, err),
						cierrors.ExitValidationFail,
					)
				}

				for k, v := range doc {
					merged[k] = v
				}
			}

			if len(merged) == 0 {
				return cierrors.New(
					"NO_FILES",
					"No YAML files found",
					fmt.Sprintf("No .yml or .yaml files found in %q", dir),
					cierrors.ExitBadArguments,
				)
			}

			out, err := yaml.Marshal(merged)
			if err != nil {
				return cierrors.New(
					"YAML_ENCODE_ERROR",
					"YAML encode error",
					fmt.Sprintf("Error encoding merged config: %v", err),
					cierrors.ExitGeneralError,
				)
			}

			fmt.Fprint(f.IOStreams.Out, string(out))
			return nil
		},
	}

	return cmd
}
