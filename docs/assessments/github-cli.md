# GitHub CLI (`gh`) — Design Assessment

**Repository:** https://github.com/cli/cli  
**Version assessed:** v2.87.3 (latest at time of assessment, April 2026)  
**Language/framework:** Go + Cobra command framework  
**Distribution:** brew, winget, apt/deb, rpm, precompiled binaries (.msi, .zip, .tar.gz) for macOS/Windows/Linux  
**Guidelines reference:** CLI Design Guidelines in this folder (clig.dev + Heroku + oclif + Thoughtworks synthesis)  
**Special focus:** GitHub Actions command surface (`gh run`, `gh workflow`, `gh secret`, `gh variable`, `gh cache`)

---

## Overview

`gh` is GitHub's official CLI, providing command-line access to the entire GitHub platform: repositories, pull requests, issues, releases, GitHub Actions, Codespaces, Projects, and more. With 43k stars and 599 contributors, it is one of the most widely deployed platform CLIs in the developer toolchain and serves as a reference implementation for many of the design patterns codified in the guidelines in this folder.

The CLI targets working software engineers across a wide skill range — from developers running a single `gh pr create` to platform engineers scripting full CI/CD automation with `gh run list --json` and `gh api`. It must serve both use cases simultaneously.

---

## Command Structure Map

### Full hierarchy (from the official manual)

```
gh [--help] [--version]
│
├── agent-task [preview]            # Copilot agent task management
│   ├── create
│   ├── list
│   └── view
│
├── alias                           # Custom command shortcuts
│   ├── delete
│   ├── import
│   ├── list
│   └── set
│
├── api <endpoint>                  # Raw REST/GraphQL API access
│
├── attestation                     # Supply-chain/build provenance
│   ├── download
│   ├── trusted-root
│   └── verify
│
├── auth                            # Authentication management
│   ├── login [--hostname] [--git-protocol] [--scopes] [--with-token] [--insecure-storage]
│   ├── logout [--hostname] [--user]
│   ├── refresh [--hostname] [--insecure-storage] [--reset-scopes] [--scopes]
│   ├── setup-git [--hostname] [--force]
│   ├── status [--hostname] [--show-token] [--active]
│   ├── switch [--hostname] [--user]
│   └── token [--hostname] [--user] [--secure-storage]
│
├── browse [<number>|<path>|<commit>]  # Open GitHub in browser
│
├── cache                           # ★ ACTIONS: Runner cache management
│   ├── delete [<cache-id>|<cache-key>] [-a/--all] [-b/--branch]
│   └── list [-b/--branch] [-L/--limit] [--json] [-q/--jq] [-t/--template]
│
├── codespace                       # GitHub Codespaces management
│   ├── code
│   ├── cp
│   ├── create
│   ├── delete
│   ├── edit
│   ├── jupyter
│   ├── list
│   ├── logs
│   ├── ports / ports forward / ports visibility
│   ├── rebuild
│   ├── ssh
│   ├── stop
│   └── view
│
├── completion [-s shell]           # Shell completion scripts (bash/zsh/fish/PowerShell)
│
├── config                          # Local config file management
│   ├── clear-cache
│   ├── get <key>
│   ├── list
│   └── set <key> <value>
│
├── copilot                         # GitHub Copilot in CLI (chat)
│
├── extension                       # Third-party CLI extensions
│   ├── browse
│   ├── create
│   ├── exec
│   ├── install
│   ├── list
│   ├── remove
│   ├── search
│   └── upgrade
│
├── gist                            # GitHub Gist management
│   ├── clone / create / delete / edit / list / rename / view
│
├── gpg-key                         # GPG key management
│   ├── add / delete / list
│
├── help                            # Help topics (not just command help)
│   ├── environment                 # All supported env vars
│   ├── exit-codes                  # Documented exit codes
│   ├── formatting                  # --json / --jq / --template usage
│   ├── mintty                      # Windows terminal notes
│   └── reference                   # Full command reference
│
├── issue                           # GitHub Issues
│   ├── close / comment / create / delete / develop
│   ├── edit / list / lock / pin / reopen
│   ├── status / transfer / unlock / unpin / view
│
├── label                           # Repository label management
│   ├── clone / create / delete / edit / list
│
├── licenses                        # List/view OSS licenses
│
├── org                             # Organization management
│   └── list
│
├── pr                              # Pull Requests
│   ├── checkout / checks / close / comment / create / diff
│   ├── edit / list / lock / merge / ready / reopen / revert
│   ├── review / status / unlock / update-branch / view
│
├── preview                         # Preview/beta features
│   └── prompter
│
├── project                         # GitHub Projects (v2)
│   ├── close / copy / create / delete / edit
│   ├── field-create / field-delete / field-list
│   ├── item-add / item-archive / item-create / item-delete / item-edit / item-list
│   ├── link / list / mark-template / unlink / view
│
├── release                         # GitHub Releases
│   ├── create / delete / delete-asset / download
│   ├── edit / list / upload / verify / verify-asset / view
│
├── repo                            # Repository management
│   ├── archive / unarchive
│   ├── autolink / autolink create / autolink delete / autolink list / autolink view
│   ├── clone / create / delete / edit / fork
│   ├── deploy-key / deploy-key add / deploy-key delete / deploy-key list
│   ├── gitignore / gitignore list / gitignore view
│   ├── license / license list / license view
│   ├── list / rename / set-default / sync / view
│
├── ruleset                         # Branch/tag rulesets
│   ├── check / list / view
│
├── run                             # ★ ACTIONS: Workflow run management
│   ├── cancel [<run-id>]
│   ├── delete [<run-id>]
│   ├── download [<run-id>] [-D/--dir] [-n/--name] [--pattern]
│   ├── list [-a/--all] [-b/--branch] [-c/--commit] [--created] [-e/--event]
│   │        [-L/--limit] [-s/--status] [-u/--user] [-w/--workflow]
│   │        [--json <fields>] [-q/--jq] [-t/--template]
│   ├── rerun [<run-id>] [--failed] [-j/--job]
│   ├── view [<run-id>] [-a/--attempt] [--exit-status] [-j/--job]
│   │        [--log] [--log-failed] [-v/--verbose] [-w/--web]
│   │        [--json <fields>] [-q/--jq] [-t/--template]
│   └── watch [<run-id>] [--exit-status] [-i/--interval]
│
├── search                          # Cross-repo GitHub search
│   ├── code / commits / issues / prs / repos
│
├── secret                          # ★ ACTIONS: Secrets management
│   ├── delete <secret-name> [-a/--app] [-e/--env] [-o/--org] [-u/--user]
│   ├── list [-a/--app] [-e/--env] [-o/--org] [-u/--user] [--json] [-q/--jq]
│   └── set <secret-name> [-a/--app] [-b/--body] [-e/--env] [-f/--env-file]
│                          [-o/--org] [-r/--repos] [-u/--user] [-v/--visibility]
│
├── skill [preview]                 # GitHub Skills (Copilot extension skills)
│   ├── install / preview / publish / search / update
│
├── ssh-key                         # SSH key management
│   ├── add / delete / list
│
├── status                          # Cross-repo activity dashboard
│
├── variable                        # ★ ACTIONS: Variables management
│   ├── delete <variable-name> [-e/--env] [-o/--org]
│   ├── get <variable-name> [-e/--env] [-o/--org]
│   ├── list [-e/--env] [-o/--org] [--json] [-q/--jq] [-t/--template]
│   └── set <variable-name> [-b/--body] [-e/--env] [-f/--env-file]
│                            [-o/--org] [-r/--repos] [-v/--visibility]
│
└── workflow                        # ★ ACTIONS: Workflow file management
    ├── disable [<workflow-id>|<workflow-name>]
    ├── enable [<workflow-id>|<workflow-name>]
    ├── list [-a/--all] [-L/--limit] [--json] [-q/--jq] [-t/--template]
    ├── run [<workflow-id>|<workflow-name>] [-F/--field] [--json] [-f/--raw-field] [-r/--ref]
    └── view [<workflow-id>|<workflow-name>] [-r/--ref] [-w/--web] [-y/--yaml]
```

### GitHub Actions surface in detail

The Actions-specific command groups (`★` above) together form a coherent CI/CD operations toolkit:

| Capability | Commands |
|---|---|
| Trigger a workflow | `gh workflow run` |
| Monitor run status | `gh run list`, `gh run view`, `gh run watch` |
| Read logs | `gh run view --log`, `gh run view --log-failed` |
| Re-run failures | `gh run rerun --failed`, `gh run rerun --job` |
| Download artifacts | `gh run download` |
| Cancel / delete | `gh run cancel`, `gh run delete` |
| Manage workflow files | `gh workflow enable/disable/view` |
| Secrets (Actions/Dependabot/Codespaces) | `gh secret set/list/delete` with `--app` scoping |
| Non-secret variables | `gh variable set/get/list/delete` |
| Runner cache | `gh cache list/delete` |

**Global flags** (inherited by all commands):
- `-R, --repo <[HOST/]OWNER/REPO>` — target a different repository
- Inferred from local git remote when inside a cloned repo (no flag required)

**Global environment variables relevant to Actions scripting:**

| Variable | Purpose |
|---|---|
| `GH_TOKEN` / `GITHUB_TOKEN` | Auth token (GITHUB_TOKEN is pre-injected in runners) |
| `GH_REPO` | Override inferred repository |
| `GH_HOST` | Override default GitHub host |
| `GH_PROMPT_DISABLED` | Suppress all interactive prompts (CI mode) |
| `GH_SPINNER_DISABLED` | Replace spinner with text progress |
| `NO_COLOR` | Disable ANSI color output |
| `GH_DEBUG=api` | Log all HTTP traffic to stderr |

---

## Category-by-Category Evaluation

### 1. Foundations / Philosophy

**Rating: Exemplary (5/5)**

`gh` has one of the clearest philosophical statements in any platform CLI: it brings GitHub to the terminal *next to where developers already work* — with git and their code. This manifests in several concrete design decisions. The CLI infers the target repository from the local git remote by default, eliminating the need for `--repo` flags in the most common case. It treats interactive use and scripted use as equally valid and first-class: every command works from a shell prompt and from a CI pipeline.

The most distinctive philosophical choice is the explicit refusal of telemetry. GitHub CLI collects no usage data whatsoever — a deliberate decision documented in its contributing guidelines. This makes `gh` unusual among platform CLIs (including the other tools assessed in this folder, which use posthog-go) and is a significant trust advantage for enterprise users who may not be able to justify egress to third-party analytics endpoints.

The `gh api` escape hatch reflects the Heroku principle of "access at every level of the stack." Rather than claiming to wrap every GitHub API endpoint, `gh` acknowledges the limits of its abstraction layer and provides a principled path out — with support for `--jq` filtering, paging, and input from stdin.

---

### 2. Command Structure and Naming

**Rating: Excellent (4.5/5)**

The top-level command vocabulary maps directly to GitHub's own object model, which makes the CLI immediately learnable for GitHub users. Every developer who knows that GitHub has "pull requests," "issues," "releases," and "secrets" can guess the right `gh` command group without consulting documentation.

The `gh <noun> <verb>` pattern is consistent across virtually all core commands. Verbs are drawn from a small, stable set: `list`, `view`, `create`, `edit`, `delete`, `close`, `reopen`, `comment`, `merge`, `clone` — all standard CRUD vocabulary. Nothing is invented.

**Specific strengths:**
- `gh run watch` is a perfectly named blocking command for CI polling — distinct from `gh run view` (inspect) and not a flag
- `gh workflow run` vs `gh run list` is a clean design: "workflow" is about the workflow *file*, "run" is about a *workflow execution*
- `gh browse` as a top-level command (no subcommand required) is the right design for a high-frequency shortcut
- `gh status` gives a cross-repo dashboard without a noun group — appropriate for its scope
- `gh auth switch` for multi-account support is the right verb — not `select` or `change`
- `gh completion` is a first-class top-level command, not buried under `help`

**Issues:**

- **`gh project` uses compound-word verbs** (`field-create`, `item-add`, `item-list`, `mark-template`). This deviates from the `<noun> <verb>` pattern used everywhere else. The proper design would be `gh project field create`, `gh project field list`, `gh project item add`, etc. — two-level noun groups nested under the project group. As written, the `gh project` command surface feels like a different tool.

- **Three-level nesting under `gh repo`** — commands like `gh repo autolink create`, `gh repo deploy-key add`, `gh repo gitignore list`, and `gh repo license view` violate the two-level maximum recommended by the guidelines. These would be better served as `gh autolink`, `gh deploy-key`, etc. — or by deferring to `gh api` for infrequent operations.

- **`gh pr update-branch`** uses a multi-word hyphenated verb. `gh pr sync` or `gh pr rebase` would be cleaner.

- **`gh release delete-asset` and `gh release verify-asset`** are verb-noun compound commands under a noun group — a minor inversion. A better surface would be `gh release asset delete` and `gh release asset verify`.

- **`gh agent-task`** (the new Copilot preview group) sits alongside production commands with no visual or structural differentiation. Users encountering it have no signal that it's experimental.

---

### 3. Help and Documentation

**Rating: Exemplary (5/5)**

`gh` sets the standard for in-binary help. Every subcommand has a usage synopsis, description, and examples inline. The examples are concrete, runnable, and cover both the simplest case and the most complex cases — exactly what the guidelines recommend.

**Specific strengths:**

- `gh help exit-codes` — exit codes are documented as a dedicated help topic. This is the right pattern: codes are `0` (success), `1` (failure), `2` (cancelled by user), `4` (auth required). All four cases are specific and actionable.
- `gh help environment` — all supported environment variables are documented in one place with precise descriptions of precedence order. Particularly thorough: `GH_TOKEN` vs `GH_ENTERPRISE_TOKEN`, `GH_EDITOR` vs `GIT_EDITOR` vs `VISUAL` vs `EDITOR` are all listed with explicit precedence ordering.
- `gh help formatting` — a dedicated topic for the `--json`/`--jq`/`--template` output system, explaining the Go templating syntax and jq syntax with examples.
- `gh help reference` — links to the full online manual.
- Every JSON-capable command lists its exact `JSON Fields` in the `--help` output (e.g., `gh run list --help` shows `attempt`, `conclusion`, `createdAt`, `databaseId`, `displayTitle`, `event`, `headBranch`, `headSha`, `name`, `number`, `startedAt`, `status`, `updatedAt`, `url`, `workflowDatabaseId`, `workflowName`). This is exceptional — users can write `--json` expressions without consulting online docs.

**Minor issue:** `gh agent-task` and `gh skill` (preview commands) appear in the full help and manual without any indication they are experimental. A `[PREVIEW]` or `(beta)` tag in the help listing would set appropriate expectations.

---

### 4. Output

**Rating: Exemplary (5/5)**

Output design is `gh`'s most consistently excellent characteristic. The following properties hold across every command that returns data:

**Machine-readable output system:**
- `--json <field1,field2,...>` on every list and view command, with JSON fields explicitly enumerated in help
- `--jq <expression>` for inline filtering (equivalent to piping to `jq`)
- `--template <string>` for Go template formatting
- These three flags compose together: `gh run list --json status,name --jq '.[] | select(.status=="failure") | .name'`

**Terminal output controls:**
- `NO_COLOR` environment variable respected (follows the `no-color.org` convention)
- `CLICOLOR=0` as an alternative
- `CLICOLOR_FORCE` to keep color even when piped
- `GH_FORCE_TTY` to force terminal-style output in scripts
- `GH_SPINNER_DISABLED` to replace the animated spinner with text progress — essential for CI logs

**Browser integration:**
- `--web` flag available on many commands (`gh run view --web`, `gh workflow view --web`, `gh pr view --web`, etc.) to open the resource in the browser when the terminal display is insufficient

**Stdout/stderr hygiene:** All interactive elements (spinners, prompts, update notices) go to stderr; all data goes to stdout. `gh run view --log` streams log output to stdout cleanly.

**Update notifications:** `GH_NO_UPDATE_NOTIFIER` and `GH_NO_EXTENSION_UPDATE_NOTIFIER` allow silencing update nags in CI environments, and the update check happens at most once every 24 hours.

The only minor note is that `gh api` does not automatically paginate when returning JSON — users must add `--paginate` manually. The documentation covers this, but it is easy to miss in scripts.

---

### 5. Errors

**Rating: Excellent (4.5/5)**

Error messages across `gh` are consistently human-readable, specific, and actionable. Authentication errors always suggest `gh auth login` or `gh auth status`. Repository-not-found errors include the repository slug that was attempted. Rate limit errors include the reset time.

The exit code design is particularly notable: the dedicated `gh help exit-codes` page documents `0`, `1`, `2`, and `4` as the standard set. Exit code `4` (authentication required) is especially valuable for scripting — it allows callers to distinguish "the command failed" from "the command needs credentials," which most CLIs collapse into a single `1`.

The `gh run view --exit-status` flag deserves special mention: it makes `gh` directly usable as a CI gate — `gh run view <id> --exit-status && echo "passed"` will fail the outer script if the monitored run failed.

**Issues:**
- **Cobra's default missing-argument errors** are terse. When a required positional argument is omitted, Cobra's standard output is `Error: accepts 1 arg(s), received 0` without contextual guidance about which argument and what values it accepts. Custom error messages have been added for many commands but not all.
- **`gh workflow run` returns only a URL** (and only "if available") when a workflow is dispatched — it does not return the run ID. This makes chaining with `gh run watch $(gh workflow run ...)` impossible without an intermediate `gh run list` call to find the newly created run. This is a known limitation of the GitHub API's `workflow_dispatch` event, but the CLI's help text does not explain the workaround.

---

### 6. Arguments and Flags

**Rating: Exemplary (5/5)**

Flag design across `gh` follows all the patterns recommended in the guidelines.

**Consistent flag vocabulary across commands:**
- `-R, --repo` — select repository (global, inherited by all commands)
- `-L, --limit` — max results (list commands)
- `--json` — machine-readable output (data commands)
- `-q, --jq` — JSON filter (data commands)
- `-t, --template` — template output (data commands)
- `-w, --web` — open in browser (view commands)
- `-b, --branch` — filter by branch (where applicable)

**Short forms on high-frequency flags:** `-b` for `--branch`, `-e` for `--event`, `-s` for `--status`, `-u` for `--user`, `-w` for `--workflow` on `gh run list`. The full set of short flags is well-chosen — the most commonly typed flags have single-character forms; rare flags are long-form only.

**`-F/--field` vs `-f/--raw-field` on `gh workflow run`** — the distinction between `-F` (respects `@` file syntax for reading from files) and `-f` (literal string value) follows the same pattern as `gh api`, which means users who know one command's flag semantics can transfer that knowledge.

**Positional arguments** are used appropriately: they are for high-confidence, well-known identifiers (run ID, PR number, workflow file name) that users always know before invoking the command. Filters and output options are always flags, never positional.

The global `--repo` flag, inherited from parent commands, is perhaps the single most important flag in the entire CLI — it enables every command to target a different repository without changing context. Its consistent availability and inheritance is a design achievement worth noting.

---

### 7. Interactivity

**Rating: Exemplary (5/5)**

The interactive / non-interactive mode detection and behavior is a model of correct design.

**In interactive sessions (TTY):**
- Commands prompt for missing required arguments (e.g., `gh workflow run` with no workflow name shows a searchable selection list)
- `gh pr create` opens an editor for the PR body if not provided via flag
- Destructive operations confirm before proceeding
- Spinners, progress bars, and colored status indicators are shown

**In non-interactive / CI mode (no TTY or `GH_PROMPT_DISABLED`):**
- All prompts are suppressed automatically
- Commands fail with a clear error message if required information is missing rather than hanging waiting for input
- `GH_PROMPT_DISABLED` provides an explicit opt-in to non-interactive mode for scripts that run in contexts where TTY detection may be unreliable (e.g., Docker containers, some CI environments)

**`GH_SPINNER_DISABLED`** replaces animated spinners with plain text progress messages — essential for CI log readability, where animated output produces thousands of redundant lines.

**`GITHUB_TOKEN` is pre-injected by GitHub-hosted runners** in GitHub Actions workflows, meaning `gh` works out-of-the-box in Actions without any authentication setup — a first-class integration with the platform `gh` is designed to manage.

---

### 8. Subcommands

**Rating: Excellent (4.5/5)**

The overall subcommand architecture is coherent and well-scoped. Each top-level group maps to a single GitHub concept. There is no overlap between groups — you would never be uncertain whether to use `gh run` or `gh workflow` for a given task once you understand the distinction (workflow file vs. workflow execution).

**Actions-specific subcommand design highlights:**

The split between `gh workflow` and `gh run` is a principled design choice that perfectly mirrors GitHub's own data model. A workflow is a YAML file that defines automation; a run is an instance of that workflow executing. Keeping them in separate command groups prevents a single overloaded `gh actions` group that would be harder to explore.

`gh run watch` as a standalone command (rather than a flag on `gh run view`) is correct — "watch until complete" is a fundamentally different use pattern (blocking, long-running, exit-code-bearing) from "view current state." Its `--exit-status` and `--interval` flags make it suitable for use in CI scripts.

`gh secret set --app {actions|codespaces|dependabot}` scopes secrets to specific apps within the same command rather than creating separate `gh actions-secret` and `gh dependabot-secret` groups. This reduces top-level sprawl while covering all secret types with one command.

**Issues:**

- **`gh project` command surface** is the most significant structural problem in the CLI. Using `field-create`, `item-add`, `item-edit`, `mark-template` as top-level verbs within a noun group deviates from every other command group's design. The correct structure would be nested noun groups: `gh project field create`, `gh project item add`. This inconsistency suggests `gh project` was added before the convention was fully established and has not been refactored.

- **`gh repo` three-level nesting** (`gh repo autolink create`, `gh repo deploy-key add`) creates inconsistency — these are the only commands in the CLI that go three levels deep.

- **`gh agent-task` and `gh skill`** are preview commands that appear alongside production commands without clear differentiation. The guidelines recommend that preview/beta commands should carry a visible experimental indicator.

- **No `gh environment` command.** Deployment environments — a central GitHub Actions concept (used with `gh secret set --env`) — cannot be managed via the CLI. Users must use the web UI or `gh api` to create, configure, or delete environments. Given that `gh secret set` and `gh variable set` already accept `--env` flags, the absence of `gh environment create/list/delete` is a meaningful gap for platform automation.

---

### 9. Robustness

**Rating: Excellent (4.5/5)**

**Positives:**
- All commands support `--repo` for explicit targeting, eliminating the risk of operating on the wrong repository
- `gh run rerun --failed` (retry only failed jobs) reflects careful thought about what engineers actually need during an incident
- The `--attempt <uint>` flag on `gh run view` allows inspecting previous run attempts — rare but valuable for debugging flaky workflows
- `gh auth status` is idempotent and safe for health checks in scripts
- `gh config clear-cache` provides a manual remedy for corrupted local state without requiring manual file system manipulation
- Build provenance attestation (`gh at verify`) using Sigstore is a supply-chain security feature that most CLIs do not provide at all

**Issues:**
- **`gh run download` does not block until the run completes.** If called on an in-progress run, it returns incomplete artifacts or an error rather than waiting. Users must chain `gh run watch && gh run download`, but this pattern is not described in `gh run download --help`.
- **No `--timeout` flag on `gh run watch`.** Long-running workflows that hang indefinitely will cause scripts using `gh run watch` to hang indefinitely too. The recommended workaround (`timeout` shell command) is external to `gh` and platform-specific.
- **Pagination defaults to 20 items on `gh run list`** with a max of 1000 via `-L`. For repositories with very high run volumes, there is no cursor-based pagination — users may miss runs that fall outside the 1000-run window.

---

### 10. Configuration and Environment

**Rating: Exemplary (5/5)**

`gh`'s configuration and environment design is comprehensive and exceptionally well-documented. The `gh help environment` page documents every supported environment variable with precise descriptions and explicit precedence ordering where multiple variables control the same setting.

**Configuration priority:** flags > environment variables > config file > defaults. This is the correct order and matches the recommendation in the guidelines.

**Config file location:** `$XDG_CONFIG_HOME/gh` (Linux), `$AppData/GitHub CLI` (Windows), `$HOME/.config/gh` (fallback) — full XDG compliance and cross-platform correctness.

**Multi-host support:** `GH_HOST` and `--hostname` flags on every auth command provide clean GitHub Enterprise Server support. `GH_ENTERPRISE_TOKEN` is distinct from `GH_TOKEN`, preventing accidental credential leakage between github.com and enterprise instances.

**Multi-account support:** `gh auth switch` allows developers managing multiple GitHub accounts to change the active account without logging out. `gh auth status` lists all authenticated accounts.

**CI-specific controls:**
- `GITHUB_TOKEN` (auto-injected by GitHub Actions runners) — no setup needed in Actions
- `GH_PROMPT_DISABLED` — suppresses interactive prompts
- `GH_SPINNER_DISABLED` — replaces animated spinner with text
- `GH_NO_UPDATE_NOTIFIER` — suppresses update notices in CI logs
- `GH_DEBUG=api` — logs HTTP requests to stderr for script debugging
- `GH_FORCE_TTY` — forces terminal-style output in piped contexts (useful when capturing `gh` output through a logging layer)

The breadth and precision of the environment variable support is exceptional — it reflects deep awareness of how `gh` is used in automated contexts.

---

### 11. Naming and Distribution

**Rating: Exemplary (5/5)**

**Naming:** `gh` is the gold-standard short-binary name. Two characters, unambiguous, directly derived from the brand, not already occupied by any system command. It is both typable and memorable.

**Distribution:**
- macOS: `brew install gh` (Homebrew core formula — no tap required)
- Windows: `winget install GitHub.cli` (WinGet), `.msi` installer
- Linux: native packages for `apt`/Debian/Ubuntu, `dnf`/Fedora/RHEL, `zypper`/openSUSE, `apk`/Alpine
- All platforms: precompiled binaries on GitHub Releases with checksums and Sigstore attestation
- GitHub-hosted runners: pre-installed on all runner images, updated weekly
- Codespaces devcontainer feature: `ghcr.io/devcontainers/features/github-cli:1`

The distribution story is comprehensive — there is no common platform or package manager missing. The inclusion of `gh` in GitHub's own runner images and as a devcontainer feature reflects the product's position as a first-class part of the GitHub ecosystem.

**Binary attestation:** Since v2.50.0, release binaries ship with SLSA build provenance attestations signed via Sigstore, verifiable with `gh at verify -R cli/cli <binary>`. This is security best practice that very few CLIs implement.

---

### 12. Analytics / Telemetry

**Rating: Exemplary (5/5)**

GitHub CLI collects **no telemetry**. This is an explicit, documented product decision. There are no analytics calls in the codebase, no third-party analytics dependencies in `go.mod`, and no consent prompts or opt-out mechanisms — because there is nothing to opt out of.

This makes `gh` notable among all three CLIs assessed in this folder. Both `bk` (Buildkite CLI) and `chunk` include posthog-go telemetry without disclosure or opt-out. The GitHub CLI sets the bar these tools should aspire to: if a CLI is part of an organization's production infrastructure, the decision to include or exclude telemetry should be explicit, documented, and respectable.

---

## Actions-Specific Deep Dive

The Actions command surface is examined here in greater depth per the assessment brief.

### What the CLI covers well

**Triggering workflows** is handled via `gh workflow run`, which supports `workflow_dispatch` events with inputs provided interactively, as `-f key=value` flags, or as JSON via stdin. The three input modes cover every use case: human invocation, scripted invocation with known values, and scripted invocation with complex payloads.

**Monitoring runs** via `gh run list` with its rich filter set (--branch, --event, --status, --workflow, --user, --created, --commit) is well-designed. The full status vocabulary (`queued`, `in_progress`, `completed`, `waiting`, `action_required`, `cancelled`, `failure`, `neutral`, `skipped`, `success`, `timed_out`, etc.) is documented in the flag description. The `--json` output makes `gh run list` a first-class data source for dashboards and scripts.

**Log access** via `gh run view --log` and `gh run view --log-failed` covers the two most common debugging patterns: read all output, or read only what failed. The `--log-failed` flag is a particularly good design because failure triage is the primary reason engineers read CI logs.

**Secrets management** via `gh secret set` is comprehensive. The `--app {actions|codespaces|dependabot}` scoping flag, the `--env` flag for deployment environment secrets, the `--org` and `--visibility` flags for organization-level secrets, and the `-f .env` bulk import pattern together cover every secret management scenario that arises in practice.

### Gaps and weaknesses

**No run ID returned by `gh workflow run`.** When `gh workflow run triage.yml` fires a `workflow_dispatch` event, the CLI prints the run URL *only if the run has already appeared in the API* by the time the response arrives. In practice this is often empty. The correct scripted workflow — trigger a run, then watch it — requires a `gh run list --workflow triage.yml --limit 1` call to find the run that was just created. This is a well-known pain point and a significant ergonomic gap for automation authors. The `--help` output acknowledges the problem ("The created workflow run URL will be returned if available") but offers no workaround.

**No `gh environment` command.** Deployment environments are a core Actions concept — they gate deployments with required reviewers, environment-specific secrets, and protection rules. Despite `gh secret set --env` and `gh variable set --env` both accepting environment names, there is no `gh environment list`, `gh environment create`, or `gh environment delete`. Platform engineers managing many repositories must use `gh api` or the web UI to manage environments. This is the most significant functional gap in the Actions surface.

**`gh run download` does not wait for completion.** Calling `gh run download` on an in-progress run returns whatever artifacts are available at that moment (often nothing). The user must manually orchestrate `gh run watch <id> && gh run download <id>`. A `--wait` flag on `gh run download` would remove this friction.

**`gh workflow run` cannot trigger non-`workflow_dispatch` events.** There is no mechanism to trigger a push, pull_request, or schedule event via `gh`. This is a GitHub platform limitation (the API only exposes `workflow_dispatch` for manual triggering), not a CLI design flaw — but it means that testing workflows that only fire on push events still requires an actual git push or `gh api` with a repository dispatch event.

**No `gh run job` subcommand.** Job-level inspection is only available as `gh run view --job <id>` and `gh run view --log --job <id>`. A developer debugging a multi-job workflow who wants to list all jobs with their statuses must use `--json jobs` to get the JSON and parse it. A `gh run job list <run-id>` command — returning a table of job names, statuses, and durations — would match the pattern of `gh run list` and be highly useful.

---

## Scorecard Summary

| Category | Rating | Score |
|---|---|---|
| 1. Foundations / Philosophy | Exemplary | 5/5 |
| 2. Command Structure and Naming | Excellent | 4.5/5 |
| 3. Help and Documentation | Exemplary | 5/5 |
| 4. Output | Exemplary | 5/5 |
| 5. Errors | Excellent | 4.5/5 |
| 6. Arguments and Flags | Exemplary | 5/5 |
| 7. Interactivity | Exemplary | 5/5 |
| 8. Subcommands | Excellent | 4.5/5 |
| 9. Robustness | Excellent | 4.5/5 |
| 10. Configuration and Environment | Exemplary | 5/5 |
| 11. Naming and Distribution | Exemplary | 5/5 |
| 12. Analytics / Telemetry | Exemplary | 5/5 |
| **Overall** | **Exemplary** | **57.5/60 (96%)** |

---

## Prioritized Recommendations

The CLI scores highly across all categories. The following recommendations address the remaining gaps, in order of impact.

**1. Return the run ID from `gh workflow run` (High)**
The most impactful improvement to the Actions surface. When `gh workflow run` dispatches a workflow, it should poll the runs API for up to ~10 seconds to find the newly created run and print its ID (or `--json` the full run object). This enables chaining: `RUN=$(gh workflow run triage.yml --json runId -q .runId) && gh run watch $RUN`. The `--help` should document the current limitation and recommend the `gh run list` workaround until the flag exists.

**2. Add `gh environment` subcommands (High)**
Add `gh environment list`, `gh environment create [--required-reviewers] [--wait-timer]`, `gh environment delete`, and `gh environment view` to complete the Actions management surface. The `--env` flag already exists on `gh secret set` and `gh variable set`; the missing piece is the ability to manage the environment object itself without falling back to `gh api`.

**3. Refactor `gh project` to use nested noun-verb structure (Medium)**
Replace `gh project field-create`, `gh project item-add`, etc. with `gh project field create`, `gh project item add`, `gh project item edit`, etc. This aligns `gh project` with every other command group and removes the only major structural inconsistency in the CLI. The transition could be done with aliases to preserve backwards compatibility during a deprecation period.

**4. Add `--wait` to `gh run download` (Medium)**
Allow `gh run download <run-id> --wait` to block until the run completes before downloading artifacts, eliminating the need to manually chain `gh run watch && gh run download`. Internally this would be equivalent to a `gh run watch` followed by `gh run download`.

**5. Add `gh run job list <run-id>` subcommand (Medium)**
A tabular view of all jobs within a run — name, status, duration, runner — would complete the job/step inspection hierarchy currently only available via `--json jobs`. This is the most commonly needed view when debugging multi-job workflow failures.

**6. Mark preview commands visually in `--help` output (Low)**
Add a `[preview]` or `(beta)` indicator to `gh agent-task` and `gh skill` in the top-level `--help` listing and their own `--help` outputs. This sets appropriate expectations without requiring users to consult documentation.

**7. Add `--timeout` to `gh run watch` (Low)**
Allow `gh run watch <id> --timeout 30m` to fail with a non-zero exit code if the run has not completed within the specified duration. This enables CI scripts to enforce a deadline without wrapping `gh` in an external `timeout` command.

---

## Comparison with `bk` (Buildkite CLI) and `chunk`

`gh` demonstrates what a platform CLI looks like when design principles are applied rigorously over a long period of active development. The contrast is instructive:

| Dimension | gh | bk | chunk |
|---|---|---|---|
| Telemetry | **None** | posthog-go (undisclosed) | posthog-go (undisclosed) |
| JSON output | Every data command | List commands only | None |
| jq filtering | `--jq` everywhere | None | None |
| Exit codes | Documented (4 codes) | Undocumented | Undocumented |
| Shell completion | `gh completion` (4 shells) | Missing | Missing |
| NO_COLOR support | Yes (`NO_COLOR`, `CLICOLOR`) | Unknown | Unknown |
| CI mode (no TTY) | `GH_PROMPT_DISABLED` | Unknown | Inferred from TTY only |
| API escape hatch | `gh api` (excellent) | `bk api` (good) | None |
| Help examples | Every subcommand | None | Sparse |
| Multi-account | `gh auth switch` | Partially (single keychain) | None |
| Command depth | Mostly 2 (3 in repo group) | Strict 2 | 3 in hook group |
| Distribution | Exemplary (all platforms) | Excellent | Good |
| Binary attestation | Sigstore + SLSA | None documented | None documented |

The areas where `bk` and `chunk` fall furthest short of the `gh` standard — telemetry disclosure, JSON output on non-list commands, shell completion, and documented exit codes — are all achievable improvements that would substantially close the gap.

---

*Assessment based on: GitHub repository at https://github.com/cli/cli (trunk branch, April 2026), CLI manual at https://cli.github.com/manual/, and the design guidelines in this folder.*
