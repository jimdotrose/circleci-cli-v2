# /new-command-group

Scaffold a complete new top-level command group with its group file, first command,
factory wiring, and test infrastructure.

**Usage:** `/new-command-group <name>`
**Example:** `/new-command-group environment`

---

## Steps

Given $ARGUMENTS as `<name>` (the new group name, e.g. `environment`):

### 1. Create the group directory and group file

Create `pkg/cmd/$name/$name.go`:

```go
package $name

import (
    "github.com/spf13/cobra"
    "github.com/circleci/circleci-cli-v2/pkg/cmdutil"
)

func NewCmd$Name(f *cmdutil.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "$name <command>",
        Short: "Manage CircleCI $names",
        Long:  "$Name commands for managing CircleCI $names.",
    }

    cmd.AddCommand(
        NewCmdList(f),
        // Add more subcommands here as they are implemented
    )

    return cmd
}
```

### 2. Scaffold the first subcommand

Create `pkg/cmd/$name/list.go` using the `/add-command` template:

```go
package $name

import (
    "github.com/MakeNowJust/heredoc"
    "github.com/spf13/cobra"
    "github.com/circleci/circleci-cli-v2/pkg/cmdutil"
    "github.com/circleci/circleci-cli-v2/pkg/output"
)

type listJSON struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    // Add fields as needed
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
    var jsonFlags output.JSONFlags

    cmd := &cobra.Command{
        Use:   "list",
        Short: "List $names",
        Long: heredoc.Docf(`
            List CircleCI $names.

            JSON Fields:
              %s
        `, output.FieldList(listJSON{})),
        Example: heredoc.Doc(`
            # List all $names:
            $ circleci $name list

            # Output as JSON:
            $ circleci $name list --json

            # Filter with jq:
            $ circleci $name list --json --jq '.[] | select(.name | startswith("prod"))'
        `),
        RunE: func(cmd *cobra.Command, args []string) error {
            return runList(f, jsonFlags)
        },
    }

    output.AddJSONFlags(cmd, &jsonFlags, []string{"id", "name"})
    cmd.Flags().StringP("limit", "L", "30", "maximum number of results")

    return cmd
}

func runList(f *cmdutil.Factory, jsonFlags output.JSONFlags) error {
    // TODO: implement
    return nil
}
```

### 3. Wire the group into the root command

Open `pkg/cmd/root/root.go` and add:

```go
import "$module/pkg/cmd/$name"

// In the function that adds subcommands:
cmd.AddCommand($name.NewCmd$Name(f))
```

### 4. Create the test file

Create `pkg/cmd/$name/list_test.go`:

```go
package $name_test

import (
    "testing"
    "github.com/circleci/circleci-cli-v2/pkg/cmdutil"
    "github.com/circleci/circleci-cli-v2/pkg/iostreams"
    "github.com/circleci/circleci-cli-v2/internal/testutil"
)

func TestCmdList_help(t *testing.T) {
    ios, _, stdout, _ := iostreams.Test()
    f := &cmdutil.Factory{IOStreams: ios}
    cmd := NewCmdList(f)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    testutil.AssertGolden(t, stdout.String(), "testdata/list.golden")
}
```

Create `pkg/cmd/$name/testdata/` directory (leave empty — golden files are generated).

Run `UPDATE_GOLDEN=1 go test ./pkg/cmd/$name/...` to create the initial golden file.

### 5. Design checklist for the new group

Before considering the group complete, verify:

- [ ] Group file has `Use`, `Short`, `Long`
- [ ] All subcommands follow `<noun> <verb>` pattern (`$name list`, `$name create`, etc.)
- [ ] `list` command has `--json`, `--jq`, `--template`, JSON fields in `Long`
- [ ] `create`/mutating commands have `--force` and/or `--dry-run` where appropriate
- [ ] Destructive commands (`delete`) have confirmation prompt + `--force` bypass
- [ ] All errors use `pkg/errors` structured type with suggestions
- [ ] Golden tests exist for `--help` output of every subcommand

### 6. Typical verb set for a new group

Design the full set of subcommands upfront (implement incrementally):

| Verb | Use |
|---|---|
| `list` | List all resources, with `--json/--jq/--template/--limit` |
| `get` / `view` | Show one resource, with `--json` |
| `create` | Create a resource, with `--dry-run` |
| `edit` / `update` | Modify a resource |
| `delete` | Remove a resource, with `--force` confirmation bypass |

Only add the verbs that make sense for the resource. Don't add `edit` if the resource
isn't editable via the API.

---

## Reference

- Command structure: `agents/08-subcommands.md`
- Output design: `agents/04-output.md`
- Full checklist: `agents/checklist.md`
- Existing command group example: `pkg/cmd/pipeline/`
