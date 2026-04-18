# /add-command

Scaffold a new subcommand inside an existing command group.

**Usage:** `/add-command <group> <verb>`
**Example:** `/add-command pipeline cancel`

---

## Steps

Given $ARGUMENTS in the format `<group> <verb>`:

### 1. Create the command file

Create `pkg/cmd/$GROUP/$VERB.go` with this structure:

```go
package $GROUP

import (
    "github.com/MakeNowJust/heredoc"
    "github.com/spf13/cobra"
    "github.com/circleci/circleci-cli-v2/pkg/cmdutil"
)

func NewCmd$Verb(f *cmdutil.Factory) *cobra.Command {
    var opts struct {
        // flags go here
    }

    cmd := &cobra.Command{
        Use:   "$verb <arg>",
        Short: "One-line description (no period, lowercase first word)",
        Long: heredoc.Doc(`
            Longer description explaining what this command does.
            Can be multiple sentences. Use present tense.
        `),
        Example: heredoc.Doc(`
            # Simple case:
            $ circleci $group $verb <arg>

            # With options:
            $ circleci $group $verb <arg> --flag value

            # Pipe the output:
            $ circleci $group $verb <arg> --json | jq '.field'
        `),
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // implementation
            return run(f, opts, args[0])
        },
    }

    // Add flags here:
    // cmd.Flags().StringVarP(&opts.Field, "flag-name", "f", "", "description")

    return cmd
}

func run(f *cmdutil.Factory, opts interface{}, arg string) error {
    // implementation
    return nil
}
```

### 2. If the command returns data — add JSON output

Declare a typed struct for the JSON shape:

```go
type $verbJSON struct {
    ID        string    `json:"id"`
    // ... fields
}
```

Add the JSON fields to the `Long` description:

```
JSON Fields:
  id, name, state, createdAt, ...
```

Wire the output flags via `pkg/output`:

```go
var jsonFlags output.JSONFlags
output.AddJSONFlags(cmd, &jsonFlags, []string{"id", "name", "state", "createdAt"})
```

Use `output.Write(f.IOStreams, data, jsonFlags)` in the run function.

### 3. If the command mutates state

Add `--force` for destructive operations (so confirmation can be bypassed in scripts):

```go
cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "skip confirmation prompt")
```

Add `--dry-run` if preview is useful:

```go
cmd.Flags().BoolVarP(&opts.DryRun, "dry-run", "n", false, "show what would happen without executing")
```

### 4. All errors must use pkg/errors

```go
import "github.com/circleci/circleci-cli-v2/pkg/errors"

// Never:
return fmt.Errorf("not found")

// Always:
return &errors.CLIError{
    Code:        "NOT_FOUND",
    Title:       "Resource not found",
    Message:     fmt.Sprintf("Pipeline %q does not exist", id),
    Suggestions: []string{"Run: circleci pipeline list to see available pipelines"},
    Ref:         "https://circleci.com/docs/api/v2/",
}
```

Use exit code constants — never raw integers:

```go
// Never: os.Exit(1)
// Always: the error type carries the exit code via ExitCode() method
```

### 5. Wire into the group

Open `pkg/cmd/$group/$group.go` and add:

```go
cmd.AddCommand(NewCmd$Verb(f))
```

### 6. Write a golden test

Create `pkg/cmd/$group/${verb}_test.go`:

```go
func TestCmd$Verb_help(t *testing.T) {
    ios, _, stdout, _ := iostreams.Test()
    f := &cmdutil.Factory{IOStreams: ios}
    cmd := NewCmd$Verb(f)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    testutil.AssertGolden(t, stdout.String(), "testdata/$verb.golden")
}
```

Run `UPDATE_GOLDEN=1 go test ./pkg/cmd/$group/...` to generate the initial golden file.

### 7. Check standards before opening a PR

Run `/check-standards` and address any failures.

---

## Reference

- Flag naming conventions: `agents/06-arguments-and-flags.md`
- Output design: `agents/04-output.md`
- Error format: `agents/05-errors.md`
- Full checklist: `agents/checklist.md`
