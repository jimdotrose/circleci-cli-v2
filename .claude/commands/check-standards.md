# /check-standards

Run the CLI design standards checklist against the current state of the codebase.
Produces a scored report showing what passes, what's partial, and what's missing.

---

## Steps

Work through each section of `agents/checklist.md` and evaluate the codebase.
For each item, mark: ✅ pass, ⚠️ partial, or ❌ fail — and for failures, point to the
specific file or command where the gap exists.

### Section 1 — Foundations

Check `pkg/errors/exitcodes.go`:
- [ ] Exit code constants defined (ExitSuccess=0 through ExitTimeout=8)
- [ ] Non-zero exit returned on all error paths in command RunE functions
- [ ] Primary output goes to `f.IOStreams.Out` (stdout), never `os.Stdout`
- [ ] Errors and messages go to `f.IOStreams.ErrOut` (stderr)
- [ ] stdin handled in commands that accept file input (no hanging on empty TTY)

### Section 2 — Help and documentation

Scan all `cobra.Command` structs in `pkg/cmd/`:
- [ ] Every command has a non-empty `Short` description
- [ ] Every command has a non-empty `Long` description (using heredoc)
- [ ] Every command has `Example` text with 3+ runnable examples (using heredoc)
- [ ] All flags document their default values
- [ ] Root-level help includes a support/issues URL
- [ ] `-h` and `--help` both work at every level

### Section 3 — Output

Check `pkg/iostreams/iostreams.go` and all command packages:
- [ ] TTY detection implemented in `IOStreams` (not inline in commands)
- [ ] Every data-returning command has `--json` flag
- [ ] JSON fields enumerated in `Long` description for every `--json` command
- [ ] `--jq` flag available on all `--json` commands (via `pkg/output`)
- [ ] `--template` flag available on all `--json` commands
- [ ] `--plain` or `--terse` flag available on list commands
- [ ] `-q`/`--quiet` global flag suppresses non-essential output
- [ ] `NO_COLOR` env var respected (check `IOStreams` constructor)
- [ ] `CIRCLECI_NO_COLOR` env var respected
- [ ] `CLICOLOR=0` respected
- [ ] `TERM=dumb` disables color
- [ ] `--no-color` flag wired to `IOStreams`
- [ ] Spinners route to stderr, not stdout
- [ ] Spinners disabled when `CI=true` or `CIRCLECI_SPINNER_DISABLED` is set
- [ ] When `--json` is active, ALL human-readable output is suppressed from stdout

### Section 4 — Errors

Check error handling across all command packages:
- [ ] All errors use the structured type from `pkg/errors` (no raw `fmt.Errorf`)
- [ ] Auth errors (exit 3) suggest `circleci auth login` or `CIRCLECI_TOKEN`
- [ ] Not-found errors (exit 5) suggest how to list available resources
- [ ] Validation errors (exit 7) point to the specific file/line that failed
- [ ] `--debug` flag enables HTTP request logging
- [ ] Stack traces hidden by default
- [ ] `--json` produces structured JSON error on stderr when active

### Section 5 — Arguments and flags

Spot-check 5+ commands across different groups:
- [ ] High-frequency flags have short forms (`-j` for `--json`, `-q` for `--quiet`, etc.)
- [ ] `--flag=value` and `--flag value` both work (Cobra default)
- [ ] Positional args: max 1 per command (2 is flagged, 3+ is a redesign)
- [ ] Sensitive flags (tokens) support stdin via `--flag -`
- [ ] All flags have env var mappings declared (check `viper.BindPFlag`)
- [ ] Standard flag names used: `-h/--help`, `-V/--version`, `-q/--quiet`, `-f/--force`, `-n/--dry-run`

### Section 6 — Interactivity

Check interactive commands (`auth login`, `context create`, destructive operations):
- [ ] Destructive operations prompt for confirmation in a TTY
- [ ] Confirmation defaults to No: `[y/N]`
- [ ] `--force` bypasses confirmation
- [ ] `--dry-run` available on commands where preview is useful
- [ ] When `CI=true` or `CIRCLECI_NO_INTERACTIVE` is set, prompts fail with clear error
- [ ] Sensitive inputs (tokens) are masked in prompts

### Section 7 — Subcommands

Check the command tree:
- [ ] All top-level groups use `<noun> <verb>` ordering
- [ ] No command goes beyond 2 levels deep without an alias at 2 levels
- [ ] Consistent verb set: list, get/view, create, edit, delete, cancel — no invented verbs
- [ ] `circleci help environment`, `circleci help exit-codes`, `circleci help formatting` exist

### Section 8 — Configuration

Check `pkg/cmdutil/factory.go` and `pkg/cmd/settings/`:
- [ ] Config priority: flags > env vars > config file > defaults (Viper)
- [ ] Config stored at `~/.circleci/cli.yml`
- [ ] `circleci settings list/get/set` commands exist
- [ ] No credentials stored in project-level config files
- [ ] `CIRCLECI_TOKEN` and `CIRCLECI_CLI_TOKEN` both accepted

### Section 9 — Robustness

Check `cmd/circleci/main.go` and long-running operations:
- [ ] SIGPIPE handled silently (no "broken pipe" error when piped to `head`)
- [ ] SIGINT (Ctrl+C) exits with `ExitCancelled` (6), not a panic
- [ ] Input validated before API calls begin (fail fast)

### Section 10 — Naming and distribution

- [ ] `circleci --version` / `-V` works at root level
- [ ] `circleci version` subcommand also works
- [ ] `circleci completion` generates scripts for bash, zsh, fish, PowerShell
- [ ] `.goreleaser.yml` produces binaries for Linux amd64/arm64, macOS amd64/arm64, Windows amd64

### Section 11 — Telemetry

Check telemetry implementation:
- [ ] First-run disclosure notice implemented
- [ ] Notice suppressed when `CI=true`
- [ ] `CIRCLECI_NO_TELEMETRY`, `NO_ANALYTICS`, `DO_NOT_TRACK` all disable collection
- [ ] No PII in telemetry events (no file paths, flag values, tokens, IPs)
- [ ] Telemetry sent asynchronously (non-blocking)
- [ ] `circleci telemetry status/enable/disable` commands exist

---

## Output format

Produce a report like this:

```
CircleCI CLI — Standards Check
═══════════════════════════════════════

Foundations ..................... ✅  5/5
Help and documentation .......... ⚠️  4/6  (missing Example on 3 commands)
Output .......................... ❌  8/14  (--json missing on runner commands)
Errors .......................... ✅  7/7
Arguments and flags ............. ⚠️  4/6  (no short flags on --org-id, --branch)
Interactivity ................... ✅  6/6
Subcommands ..................... ⚠️  3/4  (runner resource-class at 3 levels, no alias)
Configuration ................... ✅  5/5
Robustness ...................... ✅  3/3
Naming and distribution ......... ✅  4/4
Telemetry ....................... ✅  5/5

───────────────────────────────────────
Overall: 54/65 (83%)

Failures requiring action:
  ❌ pkg/cmd/runner/resource_class.go: --json flag missing on list command
  ❌ pkg/cmd/runner/token.go: --json flag missing on list command
  ⚠️  pkg/cmd/job/list.go: Example field is empty
  ⚠️  pkg/cmd/pipeline/get.go: no short flag for --branch (-b)
```

End the report with the top 3 highest-impact fixes if the score is below 55/65.
