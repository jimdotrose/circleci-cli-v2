package trigger

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdTrigger returns the `circleci trigger` command group.
func NewCmdTrigger(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger <command>",
		Short: "Manage scheduled pipeline triggers",
		Long: heredoc.Doc(`
			Commands for managing scheduled pipeline triggers.

			Scheduled triggers run pipelines on a cron-like schedule. Use
			'trigger create' to set up recurring pipeline runs for a project.
		`),
	}

	cmd.AddCommand(NewCmdTriggerCreate(f))
	return cmd
}

// NewCmdTriggerCreate returns `circleci trigger create`.
func NewCmdTriggerCreate(f *cmdutil.Factory) *cobra.Command {
	var name string
	var description string
	var timetableJSON string
	var parametersJSON string
	var actorID string

	cmd := &cobra.Command{
		Use:   "create <project-slug>",
		Short: "Create a scheduled pipeline trigger",
		Long: heredoc.Doc(`
			Create a new scheduled pipeline trigger for a project.

			The timetable controls when the pipeline runs. It must be a JSON
			object with cron-compatible fields accepted by the CircleCI API.
			Use --parameters to pass pipeline parameters to each triggered run.

			JSON Fields (response):
			  id, name, description, projectSlug, createdAt, updatedAt,
			  timetable, actor.id, actor.login
		`),
		Example: heredoc.Doc(`
			# Create a nightly trigger:
			$ circleci trigger create myorg/myrepo \
			    --name "nightly" \
			    --timetable '{"per-hour":0,"hours-of-day":[0],"days-of-week":["MON","TUE","WED","THU","FRI"]}'

			# Create a trigger with pipeline parameters:
			$ circleci trigger create myorg/myrepo \
			    --name "weekly-deploy" \
			    --timetable '{"per-hour":0,"hours-of-day":[2],"days-of-week":["MON"]}' \
			    --parameters '{"deploy_env":"staging"}'

			# Create using a specific actor:
			$ circleci trigger create myorg/myrepo \
			    --name "daily" \
			    --timetable '{"per-hour":0,"hours-of-day":[6],"days-of-week":["MON","TUE","WED","THU","FRI"]}' \
			    --actor-id $ACTOR_ID
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return cierrors.New("MISSING_ARG", "--name is required",
					"Provide a name for the scheduled trigger.", cierrors.ExitBadArguments)
			}
			if timetableJSON == "" {
				return cierrors.New("MISSING_ARG", "--timetable is required",
					"Provide a timetable JSON object.", cierrors.ExitBadArguments)
			}

			var timetable map[string]interface{}
			if err := json.Unmarshal([]byte(timetableJSON), &timetable); err != nil {
				return cierrors.New("INVALID_ARG", "Invalid --timetable JSON",
					fmt.Sprintf("Parse error: %v", err), cierrors.ExitBadArguments)
			}

			var parameters map[string]interface{}
			if parametersJSON != "" {
				if err := json.Unmarshal([]byte(parametersJSON), &parameters); err != nil {
					return cierrors.New("INVALID_ARG", "Invalid --parameters JSON",
						fmt.Sprintf("Parse error: %v", err), cierrors.ExitBadArguments)
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			stop := f.IOStreams.StartSpinner("Creating trigger...")
			st, err := client.CreateScheduledTrigger(args[0], name, description, actorID, timetable, parameters)
			stop()
			if err != nil {
				return err
			}

			if !f.IOStreams.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "✓ Created trigger %s (ID: %s)\n", st.Name, st.ID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name for the scheduled trigger")
	cmd.Flags().StringVar(&description, "description", "", "Description for the trigger")
	cmd.Flags().StringVar(&timetableJSON, "timetable", "", "Timetable as JSON object")
	cmd.Flags().StringVar(&parametersJSON, "parameters", "", "Pipeline parameters as JSON object")
	cmd.Flags().StringVar(&actorID, "actor-id", "", "Actor ID to attribute pipeline runs to")
	return cmd
}
