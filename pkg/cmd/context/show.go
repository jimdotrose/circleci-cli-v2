package context

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/apiclient"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/output"
)

// NewCmdShow returns the `circleci context show` command.
func NewCmdShow(f *cmdutil.Factory) *cobra.Command {
	var opts output.Options

	cmd := &cobra.Command{
		Use:   "show <context-id>",
		Short: "Show a context and its environment variables",
		Long: heredoc.Doc(`
			Display details for a context, including all environment variable
			names (values are never returned by the API).

			Pass the context UUID obtained from 'circleci context list'.
		`),
		Example: heredoc.Doc(`
			# Show a context by ID:
			$ circleci context show 00000000-0000-0000-0000-000000000000

			# Show as JSON:
			$ circleci context show <id> --json

			# Show variable names only:
			$ circleci context show <id> --jq '.variables[].variable'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Fetching context...")
			ctx, err := client.GetContext(id)
			stop()
			if err != nil {
				return err
			}

			var vars []apiclient.ContextVariable
			pageToken := ""
			for {
				items, next, err := client.ListContextVariables(id, pageToken)
				if err != nil {
					return err
				}
				vars = append(vars, items...)
				if next == "" {
					break
				}
				pageToken = next
			}

			type contextDetail struct {
				*apiclient.Context
				Variables []apiclient.ContextVariable `json:"variables"`
			}

			detail := &contextDetail{Context: ctx, Variables: vars}

			if err := opts.Write(f.IOStreams.Out, detail); err != nil {
				return err
			}
			if opts.IsJSON() {
				return nil
			}

			fmt.Fprintf(f.IOStreams.Out, "ID:      %s\n", ctx.ID)
			fmt.Fprintf(f.IOStreams.Out, "Name:    %s\n", ctx.Name)
			fmt.Fprintf(f.IOStreams.Out, "Created: %s\n\n", ctx.CreatedAt.Format("2006-01-02 15:04:05"))

			if len(vars) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No environment variables set.")
			} else {
				fmt.Fprintln(f.IOStreams.Out, "Environment Variables:")
				for _, v := range vars {
					fmt.Fprintf(f.IOStreams.Out, "  %s\n", v.Variable)
				}
			}
			return nil
		},
	}

	output.AddFlags(cmd, &opts, &apiclient.Context{})
	return cmd
}
