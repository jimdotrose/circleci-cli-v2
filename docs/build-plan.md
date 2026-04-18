# CircleCI CLI v2 — Full Build Plan

**Goal:** Build a new CircleCI CLI from scratch in Go + Cobra that scores 57+/60 against the CLI design guidelines in this folder, targeting functional and experiential parity with the GitHub CLI (`gh`).  
**Replaces:** `circleci-cli` (CircleCI-Public/circleci-cli)  
**Binary name:** `circleci` (replaces existing; `cci` as short alias)  
**Language/framework:** Go + Cobra + Viper  
**Build date:** April 2026

---

## 1. North Star

`gh` earns its 57.5/60 score by treating the CLI as a **scripting client that also works interactively**, not the other way around. The design decisions flow from that choice: every data command has `--json`, every flag has an env var equivalent, exit codes are documented, CI mode works with one variable.

The goal for this CLI is identical: a tool that a CircleCI engineer can use at their terminal and a platform team can embed in production CI pipelines with equal confidence. The measure of success is not feature count — it is whether the scripting contract is airtight from day one.

The existing `circleci` CLI is assessed at **38.5/60 (64%)**. A well-designed replacement should open at **55+/60** on first stable release and reach exemplary scores within two release cycles.

---

## 2. Design Principles

These eight principles from the guidelines in this folder govern every design decision. They are listed here as a team contract, not as background reading.

**Human-first.** Format output for humans by default. Composability (JSON, piping) is opt-in, not the baseline.

**Simple parts that work together.** Use standard streams, exit codes, signals. Each command does one thing and composes with others.

**Consistency.** Follow established conventions. Flags mean what users expect from `gh`, `git`, and `docker`. Deviations require explicit justification.

**Saying just enough.** Confirm what happened. Surface what the user needs to act on. Stay out of the way in scripts.

**Ease of discovery.** A user should be able to learn the CLI entirely from within the terminal. No external documentation required for common operations.

**Conversation as the norm.** CLIs are iterative. Suggest corrections, clarify states, confirm before dangerous actions, show what happened.

**Robustness.** Graceful error handling, predictable edge-case behavior, idempotency where possible. The tool should feel solid.

**Transparency.** Every action the CLI takes must be visible to the user. No silent file writes, network calls, or state changes.

---

## 3. Technical Architecture

### Repository structure

```
circleci-cli/
├── cmd/
│   └── circleci/
│       └── main.go              ← entry point, wires factory → root command
├── pkg/
│   ├── cmd/                     ← one package per top-level command group
│   │   ├── root/                ← root command, global flags, help topics
│   │   ├── auth/
│   │   ├── config/              ← pipeline YAML config commands
│   │   ├── context/
│   │   ├── orb/
│   │   ├── pipeline/
│   │   ├── project/
│   │   ├── runner/
│   │   ├── policy/
│   │   ├── trigger/
│   │   ├── settings/            ← CLI tool settings (not pipeline config)
│   │   ├── api/                 ← raw API escape hatch
│   │   └── ...
│   ├── iostreams/               ← TTY detection, color, spinner, stdout/stderr wiring
│   ├── output/                  ← JSON output, jq filtering, Go template output
│   ├── apiclient/               ← CircleCI REST + GraphQL client
│   ├── cmdutil/                 ← shared factory, helpers for building commands
│   ├── text/                    ← formatting utilities, table printing
│   └── errors/                  ← structured error types, exit code constants
├── internal/
│   └── testutil/                ← shared test helpers, golden file framework
├── docs/
│   └── manual/                  ← source for online manual (generated from Cobra)
├── .goreleaser.yml
├── go.mod
└── go.sum
```

### Key dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | Command framework: subcommands, flags, help generation, completion |
| `github.com/spf13/viper` | Config management: flags → env vars → config file → defaults |
| `github.com/itchyny/gojq` | Pure-Go jq implementation for `--jq` flag filtering |
| `golang.org/x/term` | TTY detection (`term.IsTerminal`) |
| `github.com/mattn/go-isatty` | Cross-platform isatty for color control |
| `github.com/fatih/color` | ANSI color, respects `NO_COLOR` natively |
| `github.com/briandowns/spinner` | Progress spinners, always on stderr |
| `github.com/olekukonko/tablewriter` | Table output for list commands |
| `github.com/AlecAivazis/survey/v2` | Interactive prompts (masked input, select lists) |
| `github.com/goreleaser/goreleaser` | Build and release automation |
| `github.com/sigstore/cosign` | Binary signing for SLSA provenance |
| `github.com/MakeNowJust/heredoc` | Clean multi-line help text in Go source |

### The command factory pattern

Every command package receives a `Factory` struct (identical to `gh`'s pattern) rather than using globals. This enables proper isolation in tests.

```go
// pkg/cmdutil/factory.go
type Factory struct {
    IOStreams   *iostreams.IOStreams  // stdout/stderr/stdin wiring
    Config      func() (config.Config, error)
    APIClient   func() (*apiclient.Client, error)
    BaseURL     func() string
}
```

Every command constructor takes `*Factory` and returns `*cobra.Command`. Tests instantiate a factory with mock streams and clients — no global state to reset.

---

## 4. The Scripting Contract

This is the first thing built, before any business-logic commands. Everything else depends on it.

### 4.1 Exit codes

Defined as constants in `pkg/errors/exitcodes.go` and documented as a help topic (`circleci help exit-codes`):

```go
const (
    ExitSuccess        = 0  // Command succeeded
    ExitGeneralError   = 1  // General / unclassified error
    ExitBadArguments   = 2  // Invalid arguments or flags (misuse)
    ExitAuthError      = 3  // Missing or invalid API token
    ExitAPIError       = 4  // CircleCI API returned an error (4xx/5xx)
    ExitNotFound       = 5  // Requested resource does not exist
    ExitCancelled      = 6  // Operation cancelled by user (Ctrl+C)
    ExitValidationFail = 7  // Config or policy validation failed
    ExitTimeout        = 8  // Operation timed out
)
```

The `ExitAuthError` (3) and `ExitNotFound` (5) codes are the most valuable for scripting — they let callers distinguish "need to run `circleci auth login`" from "resource was deleted" from "command logic failed." The `ExitValidationFail` (7) code is specific to CircleCI's use case: pipeline config validation is a primary CLI workflow, and CI pipelines need to gate on it.

Callers can check codes:

```sh
circleci config validate .circleci/config.yml
case $? in
  0) echo "Config valid" ;;
  7) echo "Config has errors — check output above" ;;
  3) echo "Not authenticated — run: circleci auth login" ;;
esac
```

### 4.2 Environment variables

All environment variables documented at `circleci help environment`. Variables follow `CIRCLECI_` prefix convention:

| Variable | Purpose |
|---|---|
| `CIRCLECI_TOKEN` | API token (preferred over `CIRCLECI_CLI_TOKEN` for shorter form) |
| `CIRCLECI_CLI_TOKEN` | API token (legacy alias — both accepted) |
| `CIRCLECI_HOST` | CircleCI server host (default: `https://circleci.com`) |
| `CIRCLECI_NO_INTERACTIVE` | Suppresses all prompts (equivalent to `--no-prompt` on all commands) |
| `CIRCLECI_NO_COLOR` | Disables ANSI color output |
| `CIRCLECI_SPINNER_DISABLED` | Replaces animated spinner with plain text progress |
| `CIRCLECI_NO_UPDATE_NOTIFIER` | Suppresses version update nag messages |
| `CIRCLECI_DEBUG` | Enables HTTP request logging to stderr |
| `CIRCLECI_NO_TELEMETRY` | Disables telemetry collection |
| `CI` | When set (any value), implies `CIRCLECI_NO_INTERACTIVE=1` and disables spinner and update notifications |
| `NO_COLOR` | Respects the no-color.org standard; disables ANSI color |
| `CLICOLOR=0` | Heroku/BSD convention; disables ANSI color |
| `TERM=dumb` | Disables color |

The `CI=true` detection is the most important: any CI system (GitHub Actions, Jenkins, Travis, Buildkite, and CircleCI itself) sets this, so the CLI automatically behaves correctly in automated contexts without any additional configuration.

### 4.3 CI mode behavior

When `CI=true` OR stdout is not a TTY OR `CIRCLECI_NO_INTERACTIVE` is set:
- All interactive prompts are suppressed; missing required input fails immediately with a clear error
- Animated spinners are replaced with plain-text status lines (e.g., `Validating config...`)
- Update notifications are suppressed
- Color is disabled unless explicitly forced via `CLICOLOR_FORCE=1`

### 4.4 Structured errors

All error output follows a consistent format. In terminal mode:

```
Error [AUTH_REQUIRED]: Authentication required
Your API token is missing or invalid.

Suggestions:
  → Run: circleci auth login
  → Or set CIRCLECI_TOKEN environment variable

Documentation: https://circleci.com/docs/local-cli/
```

In `--json` mode, errors are also JSON on stderr:

```json
{
  "error": true,
  "code": "AUTH_REQUIRED",
  "title": "Authentication required",
  "message": "Your API token is missing or invalid.",
  "suggestions": [
    "Run: circleci auth login",
    "Or set CIRCLECI_TOKEN environment variable"
  ],
  "ref": "https://circleci.com/docs/local-cli/"
}
```

This `errors` package is built first. Every command uses it. No raw `fmt.Errorf` strings in command handlers.

---

## 5. The IOStreams Package

`pkg/iostreams/iostreams.go` is the second thing built. It wires stdout, stderr, and stdin, and provides all output-mode decisions to every command via the factory.

```go
type IOStreams struct {
    In     io.ReadCloser
    Out    io.Writer
    ErrOut io.Writer

    // Computed from env + TTY state at construction time
    IsInteractive     bool
    ColorEnabled      bool
    SpinnerEnabled    bool
    UpdatesEnabled    bool
}

func (s *IOStreams) StartSpinner(msg string)
func (s *IOStreams) StopSpinner()
func (s *IOStreams) StartProgressBar(label string, total int)
func (s *IOStreams) Table(headers []string) *tablewriter.Table
```

Color decisions are resolved once at startup (TTY detection + env var checks + `--no-color` flag) and stored in `IOStreams`. No command ever checks `os.Getenv("NO_COLOR")` directly — they ask `IOStreams.ColorEnabled`. This makes testing trivial: inject a `IOStreams` with `ColorEnabled: false` and all output is plain text.

SIGPIPE is handled in `main.go`:

```go
// Silence broken pipe errors when output is piped to head, grep -m 1, etc.
signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
```

---

## 6. The Output Package

`pkg/output/output.go` provides the three-mode output system used by every data-returning command.

### JSON output with field enumeration

Every list and view command declares its JSON fields as a typed struct:

```go
// pkg/cmd/pipeline/list.go
type pipelineJSON struct {
    ID          string    `json:"id"`
    ProjectSlug string    `json:"projectSlug"`
    State       string    `json:"state"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    Branch      string    `json:"branch,omitempty"`
    Tag         string    `json:"tag,omitempty"`
    Number      int       `json:"number"`
    VCSRevision string    `json:"vcsRevision"`
}
```

The field names are listed in the command's `--help` output automatically via reflection. Users see:

```
JSON Fields:
  id, projectSlug, state, createdAt, updatedAt, branch, tag, number, vcsRevision
```

### --jq flag

A `--jq <expression>` flag is available alongside `--json` on every JSON-capable command. The flag invokes `gojq` inline, eliminating the external `jq` dependency for common filtering:

```sh
# All failing pipelines on main:
circleci pipeline list --json --jq '.[] | select(.state=="failed" and .branch=="main") | .id'

# Count queued pipelines:
circleci pipeline list --json --jq '[.[] | select(.state=="queued")] | length'
```

When `--jq` is used, `--json` is implied. The jq expression is compiled at parse time so syntax errors fail early with a clear message, before any API calls are made.

### --template flag

A `--template <go-template-string>` flag provides Go template formatting for custom human-readable output:

```sh
circleci pipeline list --template '{{range .}}{{.id}}\t{{.state}}\n{{end}}'
```

### Output mode precedence

```
--json → suppress all human text, emit JSON
--jq   → implies --json, then filter
--template → implies --json, then template
--plain → no color, no table formatting, tab-separated columns
(default) → human-formatted tables with color
```

When `--json` is active, **all** human-readable output is suppressed from stdout. Progress messages continue on stderr in non-CI mode. This makes JSON output reliably parseable.

---

## 7. Command Surface

The complete command taxonomy. This is the design-approved surface — implementations are phased but this is the final shape.

### Global flags (all commands)

| Flag | Short | Env var | Default |
|---|---|---|---|
| `--token` | `-T` | `CIRCLECI_TOKEN` | `~/.circleci/cli.yml` value |
| `--host` | | `CIRCLECI_HOST` | `https://circleci.com` |
| `--debug` | `-d` | `CIRCLECI_DEBUG` | `false` |
| `--no-color` | | `CIRCLECI_NO_COLOR`, `NO_COLOR` | auto (TTY detection) |
| `--quiet` | `-q` | | `false` |
| `--no-prompt` | | `CIRCLECI_NO_INTERACTIVE`, `CI` | auto |

### Complete command tree

```
circleci
│
├── auth                                  ← replaces legacy `setup` command
│   ├── login   [--no-prompt] [--host]   Token-based auth setup
│   ├── logout  [--host]                 Remove stored token
│   ├── status  [--show-token]           Show authentication state
│   └── token   [--host]                 Print active token (for scripting)
│
├── config                                ← pipeline YAML management (unchanged)
│   ├── validate   [--org-id] [file]     Validate config YAML
│   ├── process    [--org-id] [file]     Validate + expand (resolves orbs)
│   ├── pack       [dir]                 Pack multiple YAML files into one
│   ├── migrate    [file]                Migrate to current format
│   └── generate                         Generate a starter config
│
├── context
│   ├── list     [--org-id] [--json] [--jq] [--template]
│   ├── create   [--org-id] <name>
│   ├── show     <name>   [--json]
│   ├── delete   <name>   [--force]
│   └── secret
│       ├── set     <context> <name>    (replaces context store-secret)
│       └── remove  <context> <name>   (replaces context remove-secret)
│
├── orb
│   ├── list     [--json] [--jq] [--private] [--sort]
│   ├── info     <orb>   [--json]
│   ├── validate [file]
│   ├── process  [file]
│   ├── publish  <file> <orb@version>
│   ├── promote  <orb@dev:label> <segment>
│   └── search   <query>   [--json] [--jq]
│
├── pipeline
│   ├── list     [--project] [--branch] [--json] [--jq] [--limit]
│   ├── get      <id>         [--json]
│   ├── create   [--project] [--branch] [--parameters]
│   └── trigger  [--project] [--branch] [--parameters]   ← wraps trigger API
│
├── workflow
│   ├── list     <pipeline-id>   [--json] [--jq]
│   ├── get      <id>            [--json]
│   ├── cancel   <id>
│   └── rerun    <id>  [--failed]
│
├── job
│   ├── list     <workflow-id>   [--json] [--jq]
│   ├── get      <id>            [--json]
│   ├── cancel   <id>
│   └── artifacts <id>  [--json]
│
├── project
│   ├── list     [--org-id]      [--json] [--jq]
│   ├── follow   <project-slug>
│   └── env                       ← replaces project environment-variable (too deep)
│       ├── list    [--json] [--jq]
│       ├── get     <name>
│       ├── set     <name> <value>
│       └── delete  <name> [--force]
│
├── runner
│   ├── resource-class            ← kept as namespace but also shortened via aliases
│   │   ├── list    [--json] [--jq]
│   │   ├── create  <name>
│   │   └── delete  <name>  [--force]
│   ├── token
│   │   ├── list    <resource-class>  [--json]
│   │   ├── create  <resource-class>
│   │   └── delete  <token-id>  [--force]
│   └── instance
│       └── list    [--json] [--jq]
│
├── policy
│   ├── push    <bundle-dir>  [--dry-run]
│   ├── diff    <bundle-dir>
│   ├── fetch
│   ├── logs    [--json] [--jq]
│   ├── decide  [--json]
│   ├── eval
│   ├── settings
│   └── test
│
├── settings                       ← NEW: CLI tool configuration (not pipeline config)
│   ├── list                       List all CLI settings
│   ├── get   <key>                Get a setting value
│   └── set   <key> <value>        Set a setting value
│
├── api   <endpoint>               ← NEW: raw API escape hatch (like gh api)
│   [--method GET|POST|PATCH|DELETE]
│   [--field key=value]            request body fields
│   [--header key:value]           additional headers
│   [--paginate]                   follow next-page links
│   [--jq <expression>]            filter output
│   [--json]                       always output JSON
│
├── telemetry
│   ├── status
│   ├── enable
│   └── disable
│
├── open      [<project-slug>]     Open CircleCI dashboard in browser
├── diagnostic                     Check config + connectivity
├── update                         Update the CLI
├── version                        Print version (also: circleci --version / -V)
├── completion  [bash|zsh|fish|powershell]
│
└── help                           ← extended help topics (like gh help)
    ├── environment                All supported environment variables
    ├── exit-codes                 Documented exit codes
    ├── formatting                 --json / --jq / --template usage guide
    └── api                        REST API access guide for circleci api
```

### Notable design decisions in this taxonomy

**`circleci auth`** replaces `circleci setup`. The `setup` command tried to do too much (configure tool AND set up a project). Auth management belongs in `auth`. The project setup (follow + initial pipeline trigger) is a separate `circleci project follow` + `circleci pipeline trigger` — explicit, composable steps.

**`circleci context secret set/remove`** replaces the hyphenated `context store-secret` / `context remove-secret`. The noun-group-then-verb pattern (`context secret set`) is consistent with every other command group and provides a logical home for future context secret listing.

**`circleci project env`** replaces `circleci project environment-variable`. Four-level nesting is eliminated. `env` as the sub-namespace is shorter, idiomatic, and still unambiguous.

**`circleci workflow` and `circleci job`** are new top-level groups. The existing circleci CLI has no workflow or job-level inspection — you trigger a pipeline and go to the web UI to see what's happening. These commands provide terminal-native visibility into the execution graph, modeled directly on `gh run` and `gh workflow`. Together they form CircleCI's equivalent of GitHub's Actions command surface.

**`circleci api`** is the escape hatch. Rather than trying to wrap every API endpoint, `circleci api /v2/pipeline` with `--jq` and `--paginate` gives power users access to anything not covered by the command surface. This follows the Heroku principle of "access at every level of the stack."

**`circleci settings`** solves the naming collision. CLI tool configuration lives here. `circleci config` unambiguously refers to pipeline YAML. Every user who wants to change their token knows to run `circleci settings set token <value>`.

---

## 8. Help System

### In-binary documentation

Every command uses all three Cobra text fields:

```go
var cmd = &cobra.Command{
    Use:   "list",
    Short: "List pipelines for a project",
    Long: heredoc.Doc(`
        List recent pipelines for a CircleCI project.

        Pipelines are returned in reverse chronological order (newest first).
        Use --branch to filter to a specific branch, or --json to pipe
        the output to other tools.
    `),
    Example: heredoc.Doc(`
        # List pipelines for the current project (inferred from git remote):
        $ circleci pipeline list

        # Filter to main branch:
        $ circleci pipeline list --branch main

        # Get IDs of all failed pipelines as JSON:
        $ circleci pipeline list --json --jq '.[] | select(.state=="failed") | .id'

        # Count pipelines in each state:
        $ circleci pipeline list --json --jq 'group_by(.state) | map({state: .[0].state, count: length})'
    `),
}
```

The `heredoc` package keeps indentation clean in Go source while preserving the formatting users see in the terminal.

### Help topics

`circleci help environment` — documents all 14+ supported environment variables with descriptions, precedence rules, and CI-specific guidance. Modeled on `gh help environment`.

`circleci help exit-codes` — documents the 9 exit codes with their names, values, and the situations that trigger each one. Includes examples of how to branch on exit codes in shell scripts.

`circleci help formatting` — explains the `--json` / `--jq` / `--template` system with jq syntax examples. Covers how to enumerate available JSON fields for any command.

`circleci help api` — explains how to use `circleci api` for direct REST access, with examples for common operations not covered by the main command surface.

### Root help improvements

The root `circleci --help` output groups commands visually (Core, Platform, Infrastructure, Developer Tools) so users can orient themselves. The most frequently used commands (`config validate`, `pipeline trigger`, `pipeline list`) appear first within their groups. The root help includes a support link: `Open an issue: https://github.com/CircleCI-Public/circleci-cli/issues`.

---

## 9. Auth System

Auth is redesigned from `circleci setup` (a monolithic wizard) to `circleci auth` (a focused command group).

```sh
# Interactive login — prompts for token with masked input:
circleci auth login

# Non-interactive / CI login:
circleci auth login --no-prompt
# (reads token from CIRCLECI_TOKEN env var)

# For CircleCI Server (self-hosted):
circleci auth login --host https://circleci.mycompany.com

# Check auth state:
circleci auth status
# ✓ Logged in to circleci.com as jim@circleci.com (token: ••••••7f3a)

# Get the active token for use in other tools:
export TOKEN=$(circleci auth token)
```

Tokens are stored in `~/.circleci/cli.yml`. Multiple host profiles are supported — `circleci auth login --host <server-url>` writes a separate stanza. `circleci auth status` lists all authenticated hosts.

Token input in `circleci auth login` is masked (not echoed). The `--with-token` flag reads from stdin for scripted use:

```sh
echo "$MY_TOKEN" | circleci auth login --with-token
```

---

## 10. Configuration System

### Settings file

Tool configuration is stored at `~/.circleci/cli.yml`. Contents:

```yaml
# CircleCI CLI configuration
host: https://circleci.com
token: <redacted>         # stored here, never in VCS
update_check: true
telemetry: true
```

### Settings commands

```sh
circleci settings list                     # Show all settings
circleci settings get host                 # Get a single value
circleci settings set host https://...     # Set a value
circleci settings set telemetry false      # Disable telemetry via settings
```

### Priority stack (Viper)

```
CLI flags  >  CIRCLECI_* env vars  >  project .circleci/cli.yml  >  ~/.circleci/cli.yml  >  defaults
```

Project-level config (`.circleci/cli.yml` in the working directory) is useful for repository-specific defaults like `host` for repos pointed at a CircleCI Server instance. It never contains credentials.

---

## 11. Telemetry Design

Telemetry should be **opt-out, auto-disclosed, and non-blocking**.

On first run (when no `~/.circleci/cli.yml` exists), the CLI prints a single disclosure notice before executing the command:

```
CircleCI CLI collects anonymous usage data to improve the tool.
To disable: circleci settings set telemetry false
             or set CIRCLECI_NO_TELEMETRY=1
```

The notice appears once and is suppressed on subsequent runs. It is suppressed entirely when `CI=true` is set (non-interactive contexts should never have first-run UX).

Telemetry is disabled when any of the following are set: `CIRCLECI_NO_TELEMETRY`, `NO_ANALYTICS`, `DO_NOT_TRACK`.

No PII is collected: no file paths, no flag values, no API tokens, no IP addresses. Events contain only: CLI version, command name (not arguments), OS/arch, execution duration, exit code.

Telemetry events are fired asynchronously in a goroutine with a 500ms timeout. A failure to send telemetry is silently ignored — it never causes the command to fail or slow down.

---

## 12. Testing Strategy

### Unit tests

Every package has a `_test.go` file alongside it. The `Factory` pattern enables dependency injection — tests pass mock API clients and `IOStreams` instances backed by `bytes.Buffer`. No network calls in unit tests.

```go
func TestPipelineList(t *testing.T) {
    ios, _, stdout, _ := iostreams.Test()
    factory := &cmdutil.Factory{
        IOStreams:  ios,
        APIClient:  func() (*apiclient.Client, error) { return mockClient, nil },
    }
    cmd := NewCmdList(factory)
    cmd.SetArgs([]string{"--json"})
    err := cmd.Execute()
    assert.NoError(t, err)
    // Assert JSON output shape
    var result []pipelineJSON
    json.Unmarshal(stdout.Bytes(), &result)
    assert.Len(t, result, 3)
}
```

### Golden file tests

Help text is tested with golden files: run the command with `--help`, capture stdout, compare against `testdata/golden/<command>.txt`. This ensures help text regressions are visible in PRs and that example commands don't silently become wrong.

Run `UPDATE_GOLDEN=1 go test ./...` to regenerate all golden files when help text is intentionally changed.

### Contract tests for exit codes

A test suite in `pkg/errors/exitcodes_test.go` verifies that specific error conditions produce specific exit codes:

```go
func TestAuthErrorExitCode(t *testing.T) {
    // Run command with invalid token
    // Assert exit code == ExitAuthError (3)
}
```

### Integration tests

A `test/integration/` directory contains end-to-end tests that run against a real CircleCI API token (provided via `CIRCLECI_TEST_TOKEN` env var). These are marked with `//go:build integration` and skipped in normal `go test` runs. They run in CI on a dedicated schedule.

### CI self-test

The CLI includes a `circleci diagnostic` command that verifies:
- Config file is valid and readable
- Token is set and authenticates correctly
- Can reach the configured host
- Outputs a clear pass/fail summary with actionable error messages

This is used in onboarding documentation and CI pipeline health checks.

---

## 13. Distribution Pipeline

### Goreleaser

The `.goreleaser.yml` produces:
- Linux: `amd64`, `arm64`, `armv6` (`.tar.gz`, `.deb`, `.rpm`)
- macOS: `amd64`, `arm64` (`.tar.gz`, `.pkg`)
- Windows: `amd64`, `arm64` (`.zip`, `.msi`)

Checksums are computed for every artifact. Archives include a `CHANGELOG.md` extract for the release version.

### Binary signing (SLSA / Sigstore)

Release binaries are signed using Sigstore's `cosign` in keyless mode, producing SLSA build provenance attestations verifiable with:

```sh
circleci attestation verify <binary>
```

This matches what `gh` has done since v2.50.0 and is table-stakes for a security-conscious enterprise CI platform.

### Package registries

- **Homebrew**: `brew install circleci/cli/circleci` — formula in `circleci/homebrew-circleci`, auto-updated by Goreleaser on release
- **apt** (Debian/Ubuntu): signed `.deb` in a CircleCI apt repository, installable via `apt install circleci`
- **rpm** (RHEL/Fedora): signed `.rpm` in a CircleCI rpm repository
- **WinGet**: `winget install CircleCI.CLI`
- **Scoop** (Windows): manifest in `circleci/scoop-circleci`
- **Docker image**: `cimg/circleci-cli:<version>` for pipeline use — based on `alpine`, single binary
- **GitHub Release binaries**: precompiled binaries for all platforms on every release with checksums

### Shell completion

`circleci completion` generates completion scripts for bash, zsh, fish, and PowerShell. Cobra generates these automatically. The install instructions are included in `circleci completion --help`.

---

## 14. Migration Strategy

### Phase 1: Parallel availability (Months 1–3)

The new CLI ships as `circleci-v2` (or under an opt-in flag `CIRCLECI_USE_NEW_CLI=1`). The existing `circleci-cli` repository continues to receive bug fixes. Users can install and test the new CLI alongside the existing one.

### Phase 2: Feature parity (Months 3–6)

The new CLI reaches feature parity with the existing one. All commands from the existing CLI either exist in the new CLI or have documented migration paths:

| Existing command | New equivalent |
|---|---|
| `circleci setup` | `circleci auth login` |
| `circleci context store-secret` | `circleci context secret set` |
| `circleci context remove-secret` | `circleci context secret remove` |
| `circleci project environment-variable list` | `circleci project env list` |
| `circleci local execute` | `circleci local execute` (unchanged) |
| `circleci config validate <vcs> <org>` | `circleci config validate --org-id <uuid>` |

The old `context store-secret` and positional `vcs org` patterns remain supported with deprecation warnings:

```
Warning: `circleci context store-secret` is deprecated. Use `circleci context secret set` instead.
This form will be removed in v3.0. See: https://circleci.com/docs/cli/migration
```

### Phase 3: Default switch (Month 6)

The new CLI becomes the default download at `circleci.com/docs/local-cli/`. The old CLI enters maintenance-only mode. Existing installation scripts continue to work (same binary name: `circleci`).

### Phase 4: Old CLI EOL (Month 12+)

The old CLI repository is archived. A migration guide remains in the documentation for users still on the old version.

---

## 15. Phased Roadmap

### Sprint 1 — Foundation (Weeks 1–3)

Deliverables: The scaffolding that everything else depends on.

- Repository structure, `go.mod`, Cobra root command
- `pkg/iostreams` with TTY detection, color control, `NO_COLOR` / `CLICOLOR` support
- `pkg/errors` with exit code constants and structured error type
- `pkg/cmdutil` factory pattern
- Global flags: `--token`, `--host`, `--debug`, `--no-color`, `--quiet`, `--no-prompt`
- `CIRCLECI_TOKEN`, `CI`, `CIRCLECI_NO_INTERACTIVE`, `CIRCLECI_NO_COLOR` env vars wired
- `circleci help exit-codes` and `circleci help environment` topics (skeleton)
- `circleci --version` / `-V` at root AND `circleci version` subcommand
- `circleci completion` for bash/zsh/fish/PowerShell
- Goreleaser configuration producing binaries for all platforms
- Golden file test infrastructure
- CI pipeline with unit tests + linting

**Success criteria:** `circleci --help`, `circleci --version`, `NO_COLOR=1 circleci --help`, `CI=true circleci --help` all behave correctly. Binary builds and runs on macOS/Linux/Windows.

### Sprint 2 — Auth + Settings + Output Layer (Weeks 4–6)

Deliverables: The auth and configuration foundation, plus the complete output system.

- `circleci auth login/logout/status/token`
- `circleci settings list/get/set`
- `pkg/output` with `--json`, `--jq`, `--template`, `--plain`
- JSON field enumeration in `--help` for all JSON-capable commands
- Spinner with `CIRCLECI_SPINNER_DISABLED` and CI-mode plain-text fallback
- Structured error format with suggestions and doc URLs
- Auth error (exit 3) producing structured error with `circleci auth login` suggestion
- `circleci diagnostic` command
- `circleci help formatting` topic

**Success criteria:** `circleci auth login` works interactively and non-interactively. `circleci settings set host https://...` persists across invocations. `NO_COLOR=1 circleci auth status` has no ANSI codes. Exit code 3 is returned on auth failure.

### Sprint 3 — Core Command Surface (Weeks 7–12)

Deliverables: The commands that cover 90% of user workflows.

- `circleci config validate/process/pack/generate` (port from existing CLI)
- `circleci context list/create/show/delete` + `circleci context secret set/remove`
- `circleci pipeline list/get/trigger` with full `--json/--jq` support
- `circleci workflow list/get/cancel/rerun`
- `circleci job list/get/cancel/artifacts`
- `circleci orb list/info/validate/publish/promote/search`
- All commands with: examples in `--help`, `--json` with field enumeration, `--jq`, `-q`, `--force` on destructive ops

**Success criteria:** A user can validate config, trigger a pipeline, watch its progress via `circleci workflow list`, and inspect job artifacts entirely from the terminal. All commands have examples in `--help`. `circleci pipeline list --json | jq '.'` works.

### Sprint 4 — Infrastructure + API Escape Hatch (Weeks 13–16)

Deliverables: Runner, policy, project env, and the `circleci api` command.

- `circleci project list/follow` + `circleci project env list/get/set/delete`
- `circleci runner resource-class/token/instance` commands with `--json`
- `circleci policy push/diff/fetch/logs/decide/eval/settings/test`
- `circleci trigger create` / `circleci namespace create`
- `circleci api` escape hatch with `--method`, `--field`, `--paginate`, `--jq`
- `circleci help api` topic
- `circleci telemetry status/enable/disable`
- First-run telemetry disclosure notice
- Deprecation shims for old command names (with warnings)

**Success criteria:** All commands from the existing circleci CLI are available in the new CLI or have shims with deprecation warnings. `circleci api /v2/me` works with `--jq .login`.

### Sprint 5 — Polish and Migration (Weeks 17–20)

Deliverables: Help completeness, migration tooling, release automation.

- All `--help` text reviewed: examples cover simple and complex cases, flags have defaults documented, no blank `Long` descriptions
- `circleci help environment` and `circleci help exit-codes` fully populated
- Binary signing with Sigstore/cosign
- Homebrew formula, apt/rpm repository setup, WinGet manifest
- `circleci migrate` tool that reads old `~/.circleci/cli.yml` and converts to new format
- Migration documentation
- Performance profiling: startup time target < 80ms (same as `gh`)

**Success criteria:** CLI scores 55+/60 against the design guidelines checklist in this folder. Every command in the existing CLI's README works in the new CLI (either directly or via deprecated alias). Binary is available via `brew install`, `apt install`, and direct download.

---

## 16. Success Criteria and Scoring Target

Before each major release, evaluate the CLI against the design checklist (`checklist.md` in this folder). The target scores:

| Category | Target | Rationale |
|---|---|---|
| Foundations / Philosophy | 5/5 | Exit codes, streams, philosophy documented |
| Command Structure | 4.5/5 | 2-level max, noun-verb, minor nesting in runner |
| Help and Documentation | 5/5 | Examples everywhere, three help topics |
| Output | 5/5 | Full --json/--jq/--template, NO_COLOR, --quiet |
| Errors | 4.5/5 | Structured errors; Cobra missing-arg messages are still basic |
| Arguments and Flags | 4.5/5 | Short flags on high-frequency flags; some long-only acceptable |
| Interactivity | 5/5 | CI mode, --no-prompt, --force, masked input |
| Subcommands | 4.5/5 | New workflow/job groups; minor runner nesting remains |
| Robustness | 4.5/5 | SIGPIPE, SIGINT handled; pagination edge cases remain |
| Configuration and Environment | 5/5 | settings namespace, Viper stack, all vars documented |
| Naming and Distribution | 5/5 | --version flag, completion, multi-platform |
| Analytics / Telemetry | 4.5/5 | Opt-out, disclosed, CI-disabled, async |
| **Overall** | **57/60 (95%)** | |

The 3-point gap from perfect (on par with `gh`'s current score) reflects intentional conservatism: exit code coverage won't be complete on day one, runner nesting is partially preserved for backwards compatibility, and Cobra's built-in missing-argument errors are imperfect.

---

## 17. Design Anti-Patterns to Avoid

These are explicitly called out because they exist in the current circleci CLI and must not be carried forward.

**Do not name tool settings `circleci config`.** The collision between pipeline config management and CLI tool settings is the current CLI's worst design problem. `circleci settings` is the new namespace. It is non-negotiable.

**Do not add telemetry without disclosure.** The first-run notice and `CIRCLECI_NO_TELEMETRY` variable are not optional. Both `bk` and `chunk` ship posthog-go without disclosure — the new CLI should not join them.

**Do not omit `--json` from a data-returning command.** If a command returns a list or a resource, it gets `--json` with field enumeration. No exceptions. Consistent JSON coverage is what separates a scripting tool from an interactive-only tool.

**Do not add command levels beyond two without an alias.** `circleci runner resource-class list` is three levels and should have `circleci resource-class list` as an alias. `circleci project env list` is two levels. Four levels (`project environment-variable list`) should never recur.

**Do not let `circleci workflow run` fire and forget.** When a pipeline is triggered, print the pipeline ID and the pipeline URL immediately. This enables chaining (`circleci pipeline trigger ... | circleci workflow list --pipeline`) without requiring a separate list call to find what was just created.

**Do not use raw API error strings.** Every error that reaches the terminal must be wrapped by the structured error type with a suggestion and a documentation URL. Raw `"unauthorized"` strings are not acceptable.

---

*Built from the CLI design guidelines in this folder. Modeled on the GitHub CLI design assessment (`github-cli-assessment.md`). Gap analysis from `circleci-vs-gh-recommendations.md`.*
