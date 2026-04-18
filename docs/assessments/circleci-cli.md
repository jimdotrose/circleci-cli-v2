# CircleCI CLI — Design Assessment

**Source:** https://github.com/CircleCI-Public/circleci-cli  
**Framework:** Go + Cobra + Viper  
**Evaluated against:** clig.dev, Heroku, oclif, and Thoughtworks CLI design guidelines (this folder)  
**Assessment date:** April 2026

---

## Overview

The CircleCI CLI (`circleci`) is a Go-based command-line tool built on the Cobra framework. It covers a wide surface area — config validation, orb management, context management, pipeline operations, runner administration, and security policy — making it a moderately complex multi-domain CLI. It is the primary developer interface to the CircleCI platform from the terminal.

Overall the CLI is well-structured with solid fundamentals (consistent use of Cobra, Viper config, global flags with env var equivalents, shell completion, JSON output on key commands). However, it has several meaningful gaps against current design guidelines, most notably in color control, output format consistency, error message quality, and command nesting depth.

---

## Command Map

### Global Flags (inherited by all commands)

| Flag | Short | Env Var | Default | Description |
|------|-------|---------|---------|-------------|
| `--token` | | `CIRCLECI_CLI_TOKEN` | | CircleCI API token |
| `--host` | | `CIRCLECI_CLI_HOST` | `https://circleci.com` | CircleCI host URL |
| `--endpoint` | | | `graphql-unstable` | GraphQL API endpoint |
| `--debug` | | | `false` | Enable debug logging |
| `--skip-update-check` | | | `false` | Skip version check *(hidden)* |
| `--github-api` | | | | GitHub API endpoint for updates *(hidden)* |

### Complete Command Tree

```
circleci
├── setup                          Configure the tool for the first time
│
├── config
│   ├── validate                   Validate config YAML
│   ├── process                    Validate and expand config (resolves orbs)
│   ├── pack                       Pack multiple YAML files into one
│   ├── migrate                    Migrate config to newer format (delegates to agent)
│   └── generate                   Generate a starter config
│
├── context
│   ├── list      --org-id  --json  List contexts for an org
│   ├── show                        Show a context
│   ├── create    --org-id          Create a new context
│   ├── delete    --force           Delete a context
│   ├── store-secret               Store an env var in a context
│   └── remove-secret              Remove an env var from a context
│
├── orb
│   ├── list      --json --details  List orbs in registry (or namespace)
│   │             --private --sort
│   ├── validate                    Validate an orb YAML
│   ├── process                     Validate and expand an orb
│   ├── publish                     Publish an orb version
│   ├── promote                     Promote a dev orb to semantic version
│   ├── search    --json            Search the orb registry
│   └── info      --json            Show orb details
│
├── pipeline
│   ├── list      --json            List pipelines for a project
│   ├── create                      Create a pipeline
│   └── run                         Run a pipeline
│
├── trigger
│   └── create                      Create a project trigger
│
├── project
│   ├── environment-variable
│   │   ├── list                    List project env vars
│   │   ├── get                     Get a project env var
│   │   ├── create                  Create a project env var
│   │   └── delete                  Delete a project env var
│   └── dlc
│       ├── list                    List DLC (Docker Layer Cache) volumes
│       ├── rename                  Rename a DLC volume
│       └── delete                  Delete a DLC volume
│
├── runner
│   ├── resource-class
│   │   ├── list                    List runner resource classes
│   │   ├── create                  Create a resource class
│   │   └── delete  --force         Delete a resource class
│   ├── token
│   │   ├── list                    List tokens for a resource class
│   │   ├── create                  Create a runner token
│   │   └── delete  --force         Delete a runner token
│   └── instance
│       └── list                    List runner instances
│
├── policy
│   ├── push                        Push policy bundle to org
│   ├── diff                        Diff policy bundle
│   ├── fetch                       Fetch policy bundle
│   ├── logs                        Get policy decision logs
│   ├── decide                      Evaluate a policy decision
│   ├── eval                        Evaluate Rego policies
│   ├── settings                    Manage policy settings
│   └── test                        Run policy tests
│
├── namespace
│   └── create                      Create a namespace
│
├── admin  *(hidden)*
│   ├── import-orb
│   ├── rename-namespace
│   ├── delete-namespace-alias
│   └── delete-namespace
│
├── init                            One-command project setup
├── run      [name]                 Plugin: delegates to circleci-<name> in PATH
├── local
│   └── execute                     Run a job locally using Docker
├── open                            Open CircleCI dashboard in browser
├── follow                          Follow a project
├── update                          Update the CLI
├── version                         Print version
├── diagnostic                      Check config and connectivity
├── telemetry                       Manage telemetry settings
└── completion                      Generate shell completion scripts
```

---

## Assessment by Guideline Category

---

### 1. Foundations and Basics

**Exit Codes**
✅ **Pass** — Follows standard UNIX conventions: exit `0` on success, non-zero on failure. Uses Cobra's error propagation which correctly returns non-zero on errors.

**Standard Streams**
✅ **Pass** — Primary command output goes to stdout; error messages route to stderr via Cobra's built-in error handling.

**Argument Parsing Library**
✅ **Pass** — Uses Cobra, a well-maintained, idiomatic Go CLI framework. Handles flags, subcommands, help generation, and error messages correctly.

**stdin handling**
✅ **Pass** — Many commands accept `-` as a path argument to read from stdin (e.g., config file paths). Non-TTY stdin is detected; telemetry is automatically disabled in non-interactive environments.

---

### 2. Command Structure and Naming

**Command structure model**
✅ **Pass** — Uses space-separated subcommands consistently. Grouping by noun is largely followed: `config validate`, `context list`, `orb publish`, `pipeline create`. This aligns with the noun-first pattern recommended by Thoughtworks.

**Naming consistency**
⚠️ **Partial** — Most commands use clean noun-verb ordering (`context create`, `orb publish`, `policy push`). However, there are some inconsistencies:

- `context store-secret` and `context remove-secret` use verb-object within the noun group — more consistent would be `context secret store` / `context secret remove`, or flags on a `context secret` command.
- `runner resource-class` is a compound noun used as a sub-namespace rather than a command, leading to three levels of nesting (see below).

**Command nesting depth** 
❌ **Fail** — The guidelines recommend a maximum of two levels of nesting, with one level strongly preferred. The CircleCI CLI has three levels in multiple command trees:

```
circleci runner resource-class list     ← 3 levels deep
circleci runner resource-class create   ← 3 levels deep
circleci runner resource-class delete   ← 3 levels deep
circleci runner token list              ← 3 levels deep
circleci runner token create            ← 3 levels deep
circleci runner token delete            ← 3 levels deep
circleci runner instance list           ← 3 levels deep
circleci project environment-variable list   ← 4 levels deep
circleci project environment-variable create ← 4 levels deep
circleci project environment-variable delete ← 4 levels deep
```

`project environment-variable` is four levels deep, which is the deepest in the CLI. The guidelines flag three levels as a design smell requiring justification.

**Most-used commands listed first**
⚠️ **Partial** — No evidence that help output is ordered by frequency of use. Cobra defaults to alphabetical ordering. Frequently-used commands like `config validate` and `setup` should appear before administrative commands.

**Root command description**
✅ **Pass** — The root command provides a description and lists available subcommands. Help is accessible.

**Aliases**
⚠️ **Not observed** — No common short aliases found (e.g., no `circleci ctx` for `circleci context`, no `circleci cfg` for `circleci config`). Power users benefit from shorter forms for frequently-typed commands.

---

### 3. Help and Documentation

**`-h` / `--help` access**
✅ **Pass** — Standard Cobra behavior: both `-h` and `--help` work at every level of the command tree.

**Concise help on missing args**
⚠️ **Partial** — Cobra shows usage/help when required arguments are missing, but the quality of the automatic help depends on how well descriptions and usage strings are defined per command. Inconsistency in description depth was observed across commands.

**Two-level documentation (summary vs. description)**
⚠️ **Partial** — Cobra's `Short` and `Long` fields map to the summary/description two-level pattern the guidelines recommend. However, not all commands appear to use both fields — some have only a short description with no extended `Long` text, missing the opportunity for detailed guidance.

**Examples in help text**
❌ **Gap** — No systematic use of `cobra.Command.Example` was found across commands. The guidelines identify the examples section as "by far the most read and revisited" part of help text. Most CircleCI CLI commands lack command examples in their help output.

**Support path (URL for issues/feedback)**
⚠️ **Not observed** — No GitHub issues URL or support path found in root-level help text. The guidelines recommend including a link to the issue tracker.

**Flag defaults documented**
⚠️ **Partial** — Cobra auto-documents defaults when set, but not all flags explicitly show their defaults in help output. Missing `[default: value]` annotations for flags like `--host`.

**Correction suggestions for typos**
✅ **Pass** — Cobra provides built-in "did you mean?" suggestions for mistyped commands, e.g.:
```
Error: unknown command "validae" for "circleci config"
Did you mean this?
        validate
```

---

### 4. Output Design

**TTY detection**
✅ **Pass** — Telemetry and interactive prompts are automatically disabled when stdin is not a TTY. Setup command detects TTY correctly.

**`--json` flag for machine-readable output**
⚠️ **Partial** — JSON output is available on several commands (`orb list`, `orb search`, `orb info`, `context list`, `pipeline list`) but is **not consistently available across all data-returning commands**. Notably missing or unconfirmed:

- `runner resource-class list` — JSON support unclear
- `runner token list` — JSON support unclear
- `project environment-variable list` — JSON support unclear
- `policy logs` — JSON support unclear
- `context show` — JSON support unclear

The guidelines require `--json` on all commands that return structured data.

**Complete log suppression with `--json`**
❌ **Not confirmed** — The guidelines (from the oclif analysis) require that when `--json` is active, ALL human-readable output is suppressed and only the JSON object appears on stdout. It is unclear whether the CircleCI CLI fully suppresses incidental output when `--json` is used.

**`--plain` / `--terse` flag**
❌ **Gap** — No `--plain` or `--terse` flag found. For users who want to pipe tabular output to `grep` or `awk` without full JSON, no plain-text option is available.

**`-q` / `--quiet` flag**
❌ **Gap** — No `--quiet` flag found. There is no way to suppress non-essential output for scripting without redirecting stderr.

**Color control: `--no-color` flag**
❌ **Fail** — No `--no-color` flag found in the CircleCI CLI. The guidelines require this as a first-class flag alongside `NO_COLOR` env var support.

**Color control: `NO_COLOR` environment variable**
⚠️ **Uncertain** — Support for the `NO_COLOR` standard env var is not confirmed in the CLI source. Given the absence of `--no-color`, full `NO_COLOR` compliance is doubtful.

**Color control: `COLOR=false`**
⚠️ **Uncertain** — No evidence of `COLOR=false` support (Heroku ecosystem convention).

**Progress indicators / long-running operations**
⚠️ **Partial** — Some operations (like `local execute`) show progress. However, there is no consistent progress indicator pattern across long-running operations. The guidelines flag "silence creates user anxiety" as a key DX concern — long API calls without feedback are likely to occur.

**Animations disabled outside TTY**
✅ **Pass** — Telemetry and interactive elements are disabled in non-TTY environments, suggesting awareness of this concern.

**Implicit actions disclosed to user**
⚠️ **Partial** — The `init` command sets up a project, creates a pipeline, and creates a trigger in a single operation. The guidelines require that multi-step implicit actions be surfaced clearly to the user before execution. Whether `init` clearly communicates all the steps it will take is not confirmed.

---

### 5. Error Handling

**Errors to stderr**
✅ **Pass** — Cobra routes error messages to stderr by default.

**Non-zero exit on failure**
✅ **Pass** — Cobra correctly propagates errors and returns non-zero exit codes.

**Stack traces hidden by default**
✅ **Pass** — Stack traces are suppressed in normal mode. The `--debug` flag enables verbose error output.

**Error message quality: what/why/what-next**
⚠️ **Partial** — Basic Cobra errors are clear for argument validation issues. API errors from the CircleCI platform (authentication failures, API rate limits, resource not found) likely surface raw API error messages without the structured format (code, title, message, suggestions, ref) the guidelines recommend.

**Structured error fields (code, title, suggestions, ref)**
❌ **Gap** — No evidence of a consistent structured error format. Errors appear to be plain string messages without: machine-readable error codes, a short title, an array of actionable suggestions, or a documentation URL. For example, an authentication failure likely produces:

```
Error: unauthorized
```

rather than:

```
Error [AUTH_FAILED]: Authentication required
Your API token is missing or invalid.

Suggestions:
  → Run: circleci setup
  → Or set CIRCLECI_CLI_TOKEN environment variable

Documentation: https://circleci.com/docs/local-cli/#configuring-the-cli
```

**Fail fast (input validation before work)**
✅ **Pass** — Cobra validates required arguments and flags before command execution begins.

**Typo correction for flags**
✅ **Pass** — Cobra provides built-in flag suggestions.

---

### 6. Arguments and Flags

**Standard flag names**
✅ **Pass** — Standard flags are respected: `-h/--help`, `--debug`, `--version` (via `version` subcommand). No `-h` overloading observed.

**Both short and long flag forms**
⚠️ **Partial** — The global flags (`--token`, `--host`, `--debug`) have no short forms. Most per-command flags are long-form only. Short flags are sparse, which increases typing burden for frequent operations.

**Flag description style (lowercase, concise, no period)**
⚠️ **Not assessed in detail** — Would require reading all flag descriptions in detail, but Cobra-generated help may show inconsistencies across the large command surface.

**Preference for flags over positional args**
⚠️ **Mixed** — The CLI has evolved from legacy positional-argument patterns. Several commands support **dual input modes**:

```sh
# New style (preferred):
circleci context list --org-id <UUID>

# Legacy style (still accepted):
circleci context list <vcs-type> <org-name>
```

This dual-mode support means some commands accept two positional arguments (`vcs-type`, `org-name`) alongside the preferred `--org-id` flag — violating the guideline of "one positional argument is fine, two is questionable."

**The `--` passthrough convention**
✅ **Pass** — `circleci local execute` passes arguments through to the underlying Docker execution environment.

**Sensitive flags using stdin**
⚠️ **Not confirmed** — `circleci context store-secret` accepts secret values. Whether this supports reading from stdin via `--secret -` is unclear. The guidelines recommend stdin support for sensitive values to keep them out of shell history.

**Flag typing / early validation**
✅ **Pass** — UUID format validation for `--org-id` and other structured inputs appears to be present.

**Flag relationship validation (exclusive, dependsOn)**
⚠️ **Partial** — The `--no-prompt` flag requiring `--host` and `--token` is documented but it's unclear if this is enforced at parse time vs. checked in command logic.

**Env var mapping on flags**
✅ **Pass** — Key flags have env var equivalents declared:
- `--token` → `CIRCLECI_CLI_TOKEN`
- `--host` → `CIRCLECI_CLI_HOST`

However, other flags (e.g., `--org-id`, `--debug`) do not appear to have env var counterparts.

---

### 7. Interactivity

**Confirmation for destructive operations**
✅ **Pass** — Destructive operations (`context delete`, `runner resource-class delete`, `runner token delete`) support a `--force` flag, indicating confirmation prompts exist without `--force`.

**`--force` to bypass confirmation**
✅ **Pass** — `--force` flag present on destructive commands.

**`--dry-run` for previewing operations**
❌ **Gap** — No `--dry-run` flag observed. `config process` shows expanded config without applying it, which serves a similar function for config operations, but no general dry-run mechanism exists across the CLI.

**Prompts bypassable for automation**
✅ **Pass** — `--no-prompt` flag exists on `setup` to bypass the interactive setup flow.

**Prompts as first-time-user affordance**
✅ **Pass** — The `setup` command uses prompts as an onboarding flow, with `--no-prompt` available for scripted use. This matches the Thoughtworks model.

**Sensitive input masking**
⚠️ **Uncertain** — The `setup` command prompts for an API token. Whether this input is masked (not echoed) is not confirmed.

**Non-interactive/CI detection**
✅ **Pass** — Telemetry is automatically disabled in non-TTY environments. The CLI appears to detect CI contexts.

---

### 8. Configuration and Environment Variables

**Configuration file**
✅ **Pass** — Configuration stored at `~/.circleci/cli.yml`. This is a sensible home-directory location using YAML format.

**Config priority stack**
✅ **Pass** — Uses Viper for multi-source config: flags → env vars → config file → defaults. This is the correct priority order per the guidelines.

**`--config` flag**
⚠️ **Not confirmed** — Whether a `--config` flag exists to specify an alternate config file path is not confirmed. The guidelines recommend this for users with non-standard setups.

**Config management subcommands (list/get/set)**
❌ **Gap** — There is no `circleci config get` or `circleci config set` command for managing CLI tool configuration. (Note: `circleci config` refers to CircleCI YAML pipeline config, not CLI tool settings.) All CLI tool configuration is managed via the `setup` command or by editing `~/.circleci/cli.yml` directly. The guidelines recommend a `config` subcommand for managing tool settings.

> This naming creates genuine confusion: `circleci config` manages *pipeline* configuration files, while *tool* configuration is managed elsewhere. A user looking to set their API token via the CLI has no obvious command to run after `setup`.

**Standard env vars respected (NO_COLOR, CI, etc.)**
⚠️ **Partial** — `CI` appears to be respected (non-interactive mode). `NO_COLOR` support is unconfirmed.

**All env vars documented**
⚠️ **Partial** — `CIRCLECI_CLI_TOKEN` and `CIRCLECI_CLI_HOST` are documented in help and README. Other env vars (telemetry control, debug) may not be fully documented.

**Credentials handling**
✅ **Pass** — API token is stored in the config file (not committed to VCS). Missing credentials produce an error directing the user to run `circleci setup` or set `CIRCLECI_CLI_TOKEN`.

---

### 9. Robustness

**Signal handling (SIGINT/Ctrl+C)**
⚠️ **Not confirmed** — `circleci local execute` runs Docker containers, so Ctrl+C should be caught. For API-only commands, Go's default SIGINT handling should be sufficient, but whether partial state (e.g., a half-created context) is cleaned up is not confirmed.

**SIGPIPE handling**
⚠️ **Uncertain** — Go CLIs can exhibit broken pipe errors when piped to early-exiting commands (`head`, `grep -m 1`). Whether this is handled silently is not confirmed.

**Idempotency**
⚠️ **Partial** — `config validate` is idempotent. Create operations (`context create`, `runner resource-class create`) likely fail if the resource already exists — whether the error is clean and informative is unknown.

**Deprecation handling**
✅ **Pass** — The dual-mode argument support (legacy positional args + new `--org-id` flag) demonstrates backward compatibility maintenance. Cobra supports deprecation warnings on flags.

**Edge case handling**
⚠️ **Not assessed** — No explicit evidence of handling for unusual filenames, large config files, or read-only filesystems.

---

### 10. Naming and Distribution

**Tool name**
✅ **Pass** — `circleci` is lowercase, single word, not conflicting with standard UNIX tools.

**`--version` / `version` command**
⚠️ **Partial** — Version is available via `circleci version` subcommand. However, the guidelines recommend `--version` / `-V` as a flag at the root level (not just a subcommand). Many users will try `circleci --version` and find it doesn't work.

**Shell completion**
✅ **Pass** — `circleci completion` generates completion scripts for bash, zsh, and fish.

**Semantic versioning**
✅ **Pass** — Uses semantic versioning with GitHub releases.

**Update mechanism**
✅ **Pass** — `circleci update` command provided. Update check is automatic but can be skipped with `--skip-update-check`.

---

### 11. Analytics / Telemetry

**Telemetry transparency**
✅ **Pass** — `circleci telemetry` subcommand exists for managing telemetry. The CLI discloses that it collects usage data.

**Opt-out mechanism**
✅ **Pass** — Telemetry can be disabled via the `circleci telemetry disable` command.

**Non-interactive auto-disable**
✅ **Pass** — Telemetry is automatically disabled in non-TTY environments (CI, piped scripts), preventing hangs.

**Async/non-blocking**
⚠️ **Not confirmed** — Whether telemetry events are sent asynchronously (non-blocking) is not confirmed.

---

## Scorecard Summary

| Category | Score | Key Issue |
|----------|-------|-----------|
| **Basics** (exit codes, streams, arg parsing) | ✅ Good | Solid foundations |
| **Command structure and naming** | ⚠️ Partial | 3–4 levels deep in runner/project |
| **Help and documentation** | ⚠️ Partial | No examples in help text |
| **Output design** | ❌ Needs work | No `--no-color`, `NO_COLOR` unconfirmed, no `--plain`/`--quiet` |
| **Error handling** | ⚠️ Partial | No structured error format |
| **Arguments and flags** | ⚠️ Partial | Legacy dual-mode args; few short forms |
| **Interactivity** | ✅ Good | `--no-prompt`, TTY detection, `--force` |
| **Configuration** | ⚠️ Partial | No `circleci config set/get`; naming conflict |
| **Robustness** | ⚠️ Partial | Signal handling and idempotency not fully verified |
| **Naming and distribution** | ✅ Good | `circleci version` subcommand only (no `--version` flag) |
| **Telemetry** | ✅ Good | Well-handled, auto-disabled in CI |

---

## Priority Recommendations

Ranked by impact and implementation effort:

### P1 — High Impact, Relatively Straightforward

**1. Add `--no-color` flag and `NO_COLOR` env var support**
The absence of color control is the single biggest standards gap. Every CI system, log aggregator, and terminal that doesn't support ANSI colors will display escape codes as literal characters. This should be a global persistent flag.

```sh
circleci --no-color orb list
NO_COLOR=1 circleci config validate
```

**2. Add usage examples to all command help text**
The guidelines identify examples as "by far the most read and revisited" section of help. Cobra's `Example` field should be populated for every command that isn't trivially self-explanatory. Especially valuable for:

- `circleci config process` (common source of confusion)
- `circleci orb publish` (multiple steps, versioning)
- `circleci context store-secret` (piping from stdin)
- `circleci policy push`

**3. Add `--quiet` / `-q` flag globally**
Scripts running the CLI in automation contexts need a way to suppress informational output without redirecting stderr to `/dev/null`. This is a one-line addition as a persistent global flag.

---

### P2 — Medium Impact, Some Design Work Required

**4. Extend `--json` to all data-returning commands**
JSON output should be available on every command that returns structured data:
- `runner resource-class list`
- `runner token list`
- `runner instance list`
- `project environment-variable list`
- `policy logs`
- `context show`

When `--json` is active, all human-readable output should be suppressed and only the JSON object should appear on stdout.

**5. Improve error message structure**
API errors should include actionable guidance, not just the raw API error string. At minimum, authentication errors should suggest `circleci setup` or setting `CIRCLECI_CLI_TOKEN`. A structured error pattern (code, title, suggestions, ref) would significantly improve the "conversation as the norm" principle.

**6. Flatten deep command nesting**
`runner resource-class` and `project environment-variable` are too deep. Options:

- Flatten `project environment-variable` → `circleci env-var list|create|delete --project <id>`
- Flatten `runner resource-class` → `circleci resource-class list|create|delete`
- Or introduce shortcut aliases: `circleci rc list` as an alias for `circleci runner resource-class list`

**7. Add `--version` / `-V` at the root level**
A dedicated `circleci version` subcommand is fine but should be complemented by `circleci --version` at the root level, which is what most users will try first. Cobra makes this trivial to add.

---

### P3 — Lower Impact or Larger Scope

**8. Rename tool configuration namespace to avoid collision with pipeline config**
`circleci config` currently refers to pipeline YAML configuration. This creates confusion for users who want to manage *tool* configuration (API token, host). Consider:

- `circleci settings` for tool configuration management (list/get/set)
- Or document the distinction clearly in help text

**9. Add `--plain` / `--terse` flag for grep-friendly output**
For users who want to pipe tabular output without full JSON, a `--plain` flag stripping ANSI codes and formatting would improve composability.

**10. Add `--dry-run` to mutating commands**
For operations like `policy push`, `trigger create`, and `namespace create`, a dry-run mode that shows what would happen without executing is valuable for CI validation workflows.

**11. Document sensitive flag stdin support**
`circleci context store-secret` should document (and confirm support for) reading secrets from stdin via `--secret -`. This prevents secrets from appearing in shell history:
```sh
echo "$MY_SECRET" | circleci context store-secret mycontext MY_VAR -
```

**12. Improve short flag coverage for common global flags**
Adding short forms for the most frequently-typed flags reduces friction for power users:
- `--token` → `-t` or `-T`
- `--debug` → `-d`

---

## Notable Design Strengths

Before closing, it's worth recognising what the CircleCI CLI does well:

- **Cobra + Viper combination** is the correct choice for Go CLIs — mature, well-tested, extensible
- **Telemetry design** is exemplary: transparent, opt-out capable, auto-disabled in CI
- **`--no-prompt` flag** on `setup` is a good pattern for interactive/non-interactive duality
- **Shell completion** via `circleci completion` supports bash, zsh, and fish
- **Legacy compatibility** — maintaining `vcs-type org-name` alongside `--org-id` shows care for existing users
- **`--force` on destructive operations** throughout the CLI is consistent
- **Plugin system** (`circleci run <name>` → `circleci-<name>` in PATH) is a clean extension model
- **`--skip-update-check`** prevents the update check from slowing down scripted use cases
- **Hidden admin commands** — using Cobra's hidden flag to keep operational commands out of standard help is the correct pattern

---

*Sources: [CircleCI CLI GitHub](https://github.com/CircleCI-Public/circleci-cli) · [CircleCI Local CLI Docs](https://circleci.com/docs/local-cli/) · CLI design guidelines in this folder*
