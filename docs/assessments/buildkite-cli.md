# Buildkite CLI (`bk`) — Design Assessment

**Repository:** https://github.com/buildkite/cli  
**Version assessed:** v3.35.2 (latest at time of assessment, April 2026)  
**Language/framework:** Go + Kong arg parser (`github.com/alecthomas/kong`)  
**Distribution:** goreleaser binary releases (.deb, .rpm, .apk, .tar.gz/.zip, brew tap)  
**Guidelines reference:** CLI Design Guidelines in this folder (clig.dev + Heroku + oclif + Thoughtworks synthesis)

---

## Overview

`bk` is the official Buildkite platform CLI, providing command-line access to Buildkite's CI/CD platform. It covers pipeline management, build triggering and monitoring, artifact handling, agent management, cluster configuration, and direct API access. The tool targets CI/CD engineers and DevOps practitioners who work with Buildkite at scale.

Unlike many Go CLIs (which default to Cobra), `bk` uses Kong — a struct-tag–driven command framework that enforces a strongly typed, hierarchical command model at compile time rather than at runtime registration. It is distributed as native binaries via goreleaser and brew, with no runtime dependency.

---

## Command Structure Map

```
bk [--help] [--version]
│
├── auth
│   ├── login [--org <slug>] [--token <token>]
│   └── logout [--all]
│
├── build
│   ├── view [<pipeline-slug>] [<build-number>]
│   │       [--job-states <states>]
│   ├── create [--pipeline <slug>] [--branch <branch>]
│   │          [--message <msg>] [--commit <sha>]
│   │          [--metadata <key=value>...]
│   ├── cancel [<pipeline-slug>] [<build-number>]
│   ├── list [--pipeline <slug>] [--branch <branch>]
│   │        [--message <text>] [--limit <n>]
│   │        [-o json]
│   └── rebuild [<pipeline-slug>] [<build-number>]
│
├── artifact
│   └── list [<build-number>] [--job <job-id>]
│            [-p <pipeline>]
│
├── pipeline
│   ├── list [--name <name>] [--repo <url>]
│   │        [--limit <n>] [-o json]
│   ├── create [--description <text>] [--repository <url>]
│   │          [--cluster <id>] [-o <file>] [--dry-run]
│   ├── copy [--target <slug>] [--cluster <id>] [--dry-run]
│   └── convert [--file <path>] [--vendor <name>] [-o <file>]
│
├── cluster
│   ├── list
│   ├── create
│   ├── update
│   └── delete
│
├── agent
│   ├── list [--limit <n>]
│   ├── install
│   ├── run
│   └── stop
│
├── job
│   └── unblock [<job-id>] [--field <key=value>...]
│
├── configure                              # Legacy alias for auth login
│   └── add [--org <slug>] [--token <token>]
│
├── config                                 # View/manage local bk.yaml config
│
└── api <endpoint> [--method GET|POST|PUT]
         [--data <json>] [--analytics]
```

**Global environment variables:**

| Variable | Description |
|---|---|
| `BUILDKITE_API_TOKEN` | REST API token; overrides keychain value |
| `BUILDKITE_ORG` | Default organization slug |

**Configuration file:** `~/.config/buildkite/bk.yaml` (XDG base directory); tokens stored in OS keychain by default since v3.32.0.

---

## Category-by-Category Evaluation

### 1. Foundations / Philosophy

**Rating: Strong (4/5)**

The CLI reflects a coherent "operator-first" philosophy: its primary users are engineers automating CI workflows, and the tool is designed for both interactive and scripted use. The command vocabulary (`build create`, `build view`, `pipeline list`) maps naturally to the Buildkite platform's own object model, making discoverability intuitive for users who already know the platform.

The API escape hatch (`bk api`) is a particularly strong design choice — it acknowledges that the CLI cannot wrap every endpoint and provides a principled fallback rather than forcing users to reach for `curl`. This follows the Heroku principle of providing "access at every level of the stack."

The one notable philosophical gap is that the CLI is *opaque by default on what it is doing*: several commands use SpinWhile spinners that swallow the underlying API calls being made, making it harder to debug or script with confidence. A `--verbose` / `--dry-run` flag on more commands would reinforce the trust-building principle from the guidelines ("humans should be able to understand what the tool is doing").

---

### 2. Command Structure and Naming

**Rating: Good (4/5)**

The command hierarchy is well-structured and follows the `bk <noun> <verb>` pattern consistently across all top-level commands:

- `bk build view`, `bk build create`, `bk build cancel` — correct noun-verb ordering
- `bk pipeline list`, `bk pipeline create`, `bk pipeline convert` — consistent
- `bk agent list`, `bk agent install`, `bk agent run` — consistent
- `bk cluster list/create/update/delete` — consistent CRUD vocabulary

No command inverts to verb-noun. The hierarchy stays at exactly two levels (noun group + verb), matching the guidelines' recommendation to avoid going deeper than two levels unless absolutely necessary.

**Positives:**
- CRUD verbs are used consistently (`list`, `create`, `update`, `delete`) rather than inventing synonyms
- `bk pipeline convert` clearly names its function without overloading a generic verb
- `bk api` is a well-chosen name for an escape-hatch command

**Issues:**
- `bk configure` (legacy) conflicts semantically with `bk auth login` — two commands that do the same thing with no clear deprecation message. The legacy alias is undocumented in the primary reference but still ships with the binary, creating confusion for new users who may encounter it in older blog posts or Stack Overflow answers.
- `bk artifact` (singular) vs. `bk artifacts` — both appear in different docs pages and changelog entries. The singular vs. plural inconsistency is a naming debt that creates friction when tab-completing.
- `bk config` (local config file) and `bk configure` (legacy auth alias) are too similar in name and distinct in purpose; users encountering them side by side will struggle to differentiate.

---

### 3. Help and Documentation

**Rating: Adequate (3/5)**

The CLI provides contextual `--help` output via Kong's built-in help generation, and Buildkite maintains a full reference documentation site at `buildkite.com/docs/platform/cli/reference`. However, several gaps exist.

**Positives:**
- Each subcommand produces focused `--help` output with flag descriptions
- `bk` alone (no subcommand) lists all top-level groups with brief descriptions
- The online docs include usage examples for the key commands

**Issues:**
- **No `--help` examples in the binary itself.** Kong's generated help does not include usage examples; the help text describes flags but never shows a concrete invocation. The guidelines (clig.dev and Heroku) strongly recommend at least one example per command, ideally at the end of the `--help` output. For a CI tool where commands often require multiple flags in combination (e.g., `bk build create --pipeline foo --branch main --message "hotfix"`), examples would dramatically reduce trial-and-error.
- **`bk pipeline convert` help does not document supported `--vendor` values.** A user seeing `--vendor STRING` in the help text has no indication that valid values are `circleci` and `github` without consulting the online docs.
- **No man page.** The distribution format (.deb, .rpm) makes man page inclusion feasible but none is shipped. This is a missed opportunity for users who prefer local documentation.
- **The legacy `configure` command produces no deprecation notice** when invoked, violating the principle that deprecated paths should always surface an actionable upgrade message.

---

### 4. Output

**Rating: Improving (3/5)**

The output story has been actively improving across recent releases, but there are still meaningful gaps.

**Positives:**
- `bk build list` and `bk pipeline list` support `-o json` for machine-readable output — this is the right flag name and follows the convention recommended by the guidelines
- v3.35.2 added explicit stdout protection, ensuring that progress messages and interactive content go to stderr, keeping stdout clean for piped usage — a critical fix that reflects awareness of the guidelines principle "only put output intended for machines on stdout"
- The SpinWhile pattern routes spinner UI to stderr, which is the correct design

**Issues:**
- **JSON output is only available on `list` commands, not on `create` or `view`.** Creating a pipeline or viewing a build produces human-formatted output only; there is no `bk pipeline create -o json` that returns the created object's full JSON for scripting. This significantly limits programmatic use of non-list commands.
- **No `--quiet` flag.** No command supports suppressing non-essential output for use in shell scripts where only the exit code matters. The guidelines explicitly recommend a `--quiet`/`-q` flag as part of a consistent output control vocabulary.
- **No `--no-color` / `NOCOLOR` support documented.** It is unclear from the docs or changelog whether the CLI respects the `NO_COLOR` environment variable convention. Rich terminal output (colors, spinners) without an opt-out is an accessibility and CI-environment concern.
- **`bk build view` renders a TUI-style build status view** — this is impressive for interactive use but creates a problem in CI environments (e.g., another CI system watching Buildkite) where a non-interactive render may misbehave or produce garbled output. There is no documented `--format` flag to force a plain-text mode.

---

### 5. Errors

**Rating: Adequate (3/5)**

Error handling follows Go conventions (errors bubble up as messages) but lacks the consistency and actionability that the guidelines recommend.

**Positives:**
- Authentication errors (`401 Unauthorized`) produce a human-readable message pointing users to `bk auth login` rather than a raw HTTP response
- Build-not-found errors include the pipeline slug and build number that was requested, giving the user enough context to diagnose the problem

**Issues:**
- **Error messages do not consistently suggest a fix.** The guidelines recommend that every error message answer "what should the user do next?" For example, a missing `--pipeline` flag should say "Use `--pipeline <slug>` to specify which pipeline to target, or set `BUILDKITE_ORG` in your environment." Instead, Kong's default missing-argument error is terse: `expected --pipeline <STRING>`.
- **Exit codes are not documented.** The help output and online docs make no mention of what exit codes the CLI returns. For CI integration, predictable exit codes (`0` = success, `1` = usage error, `2` = API error) are essential. Exit code behavior appears to follow Go defaults but is untested and untested behavior in critical paths is a liability.
- **`bk api` errors expose raw GraphQL error payloads** on stderr without normalization. GraphQL errors (which nest under `errors[].message`) are printed as-is, which can be hard to parse when scripting.

---

### 6. Arguments and Flags

**Rating: Good (4/5)**

The flag design is generally clean and idiomatic. Kong enforces typed flags at compile time, which reduces the chance of runtime type confusion.

**Positives:**
- Flags consistently use `--long-form` with no ambiguous single-letter flags beyond the well-established conventions (`-o` for output, `-p` for pipeline)
- Required vs. optional arguments are appropriately separated: pipeline slug and build number are positional arguments (you know them or you don't), while optional filters are flags
- `--dry-run` is available on `pipeline create` and `pipeline copy`, which is an excellent affordance for complex, potentially destructive operations

**Issues:**
- **Flag naming is inconsistent across commands.** `bk build list` uses `--pipeline <slug>` while `bk artifact list` uses `-p <pipeline>`. The same concept has two different flag names, which makes scripting harder and muscle memory unreliable.
- **No short forms for common flags.** `--branch`, `--message`, `--commit`, `--limit` have no short forms. Given how frequently these are typed in daily use (`bk build list --branch main --limit 10`), the lack of short aliases (`-b`, `-m`, `-l`) is a real ergonomic cost. The guidelines recommend short forms for any flag used frequently in interactive sessions.
- **`--job-states` on `bk build view` accepts unclear value format.** The help text does not clarify whether values are comma-separated, space-separated, or repeated flags (`--job-states passed --job-states failed`). This ambiguity requires consulting the source or experimenting.
- **`--data` on `bk api` accepts raw JSON string** — this is appropriate but there is no documented alternative for file input (e.g., `--data @file.json`), which is a common curl-compat pattern that users will expect.

---

### 7. Interactivity

**Rating: Good (4/5)**

The CLI correctly detects TTY context for interactive vs. non-interactive mode. The auth login flow prompts for org and token when not provided as flags, but correctly falls back to flag-only mode when stdin is not a TTY.

**Positives:**
- `bk auth login` is interactive when run manually (prompts for org slug and token) but accepts `--org` and `--token` flags for non-interactive / scripted use — the right design
- SpinWhile spinners correctly target stderr and do not appear when stdout is piped
- `bk build view` shows a rich TUI in an interactive terminal

**Issues:**
- **No documented behavior for `CI=true` or `--no-interactive` flag.** Some CI environments set `CI=true`; it is unclear whether the CLI detects this and suppresses all interactive prompts automatically. The guidelines recommend that CLIs never prompt when running in a known CI environment (detected by common env vars).
- **`bk pipeline create` and `bk pipeline copy` may prompt for missing required fields in interactive sessions** — if this is the case, the prompts should be suppressible by a `--no-interactive` flag for use in automation scripts. This behavior is not documented.

---

### 8. Subcommands

**Rating: Strong (4/5)**

The subcommand organization is one of the CLI's clearest strengths. Every resource type has its own group, and the groups are mutually exclusive in their responsibilities.

**Positives:**
- Resources map 1:1 to Buildkite platform concepts: `build`, `pipeline`, `cluster`, `agent`, `artifact`, `job` — no guesswork about where a command lives
- Depth is strictly limited to two levels (`bk <resource> <action>`) across all commands
- `bk api` provides a principled escape hatch for operations not yet wrapped by dedicated subcommands
- `bk pipeline convert` is a well-scoped migration utility that avoids polluting the top-level namespace

**Issues:**
- **`bk configure` (legacy) should be formally deprecated** with a visible notice in the `--help` output and a migration path to `bk auth login`. As long as it ships silently, it creates a split-brain documentation problem — old tutorials and CI configs using `bk configure add` will quietly work but users will not benefit from the new keychain-backed auth.
- **`bk config` (config file management) and `bk configure` (legacy auth) are confusingly close in name.** Running `bk config` is not the same as `bk configure`, but the names look nearly identical in shell history. Renaming the config inspection command to `bk settings` or `bk info` would reduce ambiguity.
- **No `bk completion` command** for generating shell completion scripts (bash, zsh, fish). This is a significant omission for a tool with this many subcommands, flags, and pipeline/org slugs as positional arguments. Kong supports shell completion generation; enabling it would substantially improve the daily-use experience.

---

### 9. Robustness

**Rating: Adequate (3/5)**

**Positives:**
- Kong's compile-time type enforcement means there are fewer classes of runtime parsing errors compared to dynamically registered Cobra commands
- REST migration for `bk artifact` (v3.35.0) replaced a GraphQL-dependent path with the more stable REST API — a good reliability investment
- `goreleaser` + checksums and signed releases (implied by the distribution variety) indicate production-grade release hygiene

**Issues:**
- **Rate limiting and retry behavior is undocumented.** When Buildkite's REST or GraphQL APIs rate-limit a request, it is unclear whether the CLI retries automatically, surfaces a human-readable "rate limited, retrying in Xs" message, or simply fails. For commands run in CI pipelines (e.g., `bk build create` called by a post-deploy hook), silent failure due to rate limiting is a critical gap.
- **`bk agent run` behavior under failure is undocumented.** Does the agent reconnect automatically on connection drop? Does `bk agent run` block indefinitely or time out? These are robustness questions critical to production use that the docs and help output do not address.
- **No `--timeout` flag on any command.** In long-polling scenarios (e.g., waiting for a build to complete), users have no way to set a deadline without wrapping the command with an external timeout tool.

---

### 10. Configuration and Environment

**Rating: Strong (4/5)**

**Positives:**
- System keychain storage for tokens (v3.32.0) is a meaningful security improvement over plaintext config files — aligns with the guidelines' preference for secure credential storage
- `BUILDKITE_API_TOKEN` and `BUILDKITE_ORG` environment variables allow CI pipelines and containerized environments to supply credentials without a config file
- XDG-compliant config file location (`~/.config/buildkite/bk.yaml`)
- The configuration priority is coherent: keychain > env var > config file > flag default

**Issues:**
- **No `--profile` flag for multi-org workflows.** Engineers managing multiple Buildkite organizations (e.g., an agency or a company with separate orgs for prod vs. staging) have no way to switch contexts cleanly from the command line. The auth system stores only one org's token in the keychain by default. `bk auth login` with `--all` flag for logout suggests multiple orgs can be stored, but there is no `--profile` or `--org` global flag to select which org to use for a given invocation.
- **`bk auth token` (added v3.34.0) is underdocumented.** Its purpose — printing the stored token for use in scripts — is a valid and frequently needed operation, but the docs and help output do not clarify whether it prints to stdout (correct) or outputs it in a formatted block (problematic for `$(bk auth token)` shell substitution).
- **Config file schema is not published.** Users who want to manage `bk.yaml` via config management (Ansible, Chef, Terraform) cannot validate their config files without running `bk` itself.

---

### 11. Naming and Distribution

**Rating: Strong (5/5)**

**Positives:**
- `bk` is short, memorable, and unambiguous within the Buildkite ecosystem
- Goreleaser distributes native packages for all major Linux package managers (.deb, .rpm, .apk) and macOS (brew tap `buildkite/buildkite/bk`), macOS .zip, and Linux .tar.gz
- A brew tap (`brew tap buildkite/buildkite && brew install buildkite/buildkite/bk`) is the right canonical install path for macOS users
- The binary name `bk` is short enough for daily interactive use and long enough to be distinct (not a common system command)
- Release artifacts are checksummed and the goreleaser configuration implies standard artifact signing

**Minor issues:**
- The brew package name (`buildkite/buildkite/bk`) requires a two-step `tap + install`; a Homebrew core formula would simplify discovery for users who don't already know the tap name. This is a common trade-off for commercial products that want control over their release cadence.
- A Windows installer (.msi or Scoop manifest) is not documented, which may be a gap for organizations with Windows-heavy developer environments.

---

### 12. Analytics / Telemetry

**Rating: Needs Improvement (2/5)**

The CLI includes `posthog-go` as a dependency, confirming that usage telemetry is collected. However:

**Issues:**
- **Telemetry collection is not mentioned in the `--help` output, README, or documentation.** Users have no in-product notice that their command invocations (and potentially org slugs, pipeline slugs, and subcommand patterns) are sent to PostHog.
- **No opt-out mechanism is documented.** There is no `--no-analytics`, `BUILDKITE_CLI_NO_ANALYTICS=1`, or similar flag/env var mentioned in any official documentation.
- **This is the single most significant compliance gap.** Enterprise users in regulated industries (financial services, healthcare, government) often have network egress controls or legal requirements prohibiting undocumented telemetry to third-party SaaS endpoints. Undocumented telemetry to PostHog (a third party) from a tool used inside CI pipelines could violate data processing agreements.

The chunk CLI has the same issue with its posthog-go dependency and an identical rating was assigned. Both tools would benefit from following GitHub CLI's model: explicit first-run consent prompt, `--no-analytics` flag, and a documentation page describing what is collected and retained.

---

## Scorecard Summary

| Category | Rating | Score |
|---|---|---|
| 1. Foundations / Philosophy | Strong | 4/5 |
| 2. Command Structure and Naming | Good | 4/5 |
| 3. Help and Documentation | Adequate | 3/5 |
| 4. Output | Improving | 3/5 |
| 5. Errors | Adequate | 3/5 |
| 6. Arguments and Flags | Good | 4/5 |
| 7. Interactivity | Good | 4/5 |
| 8. Subcommands | Strong | 4/5 |
| 9. Robustness | Adequate | 3/5 |
| 10. Configuration and Environment | Strong | 4/5 |
| 11. Naming and Distribution | Strong | 5/5 |
| 12. Analytics / Telemetry | Needs Improvement | 2/5 |
| **Overall** | **Good** | **43/60 (72%)** |

---

## Prioritized Recommendations

The following recommendations are ordered by impact and effort. Items 1–4 address compliance or user-trust issues; items 5–10 address design quality.

**1. Document and provide opt-out for telemetry (Critical)**
Before the next major release, add a documented telemetry opt-out (`BUILDKITE_CLI_NO_ANALYTICS=1` or `--no-analytics`), a consent notice on first run, and a documentation page describing what events are collected, how long they are retained, and whether they are shared with third parties. This is a legal and trust requirement for enterprise customers.

**2. Deprecate `bk configure` with a migration notice (High)**
The `bk configure` alias should display a visible deprecation warning when invoked (`DEPRECATED: use 'bk auth login' instead`) and should be scheduled for removal. Until it is removed, it creates split documentation and confuses new users.

**3. Add `--help` usage examples to all subcommands (High)**
Each subcommand's help output should include at least one end-to-end usage example. Kong does not generate these automatically; they must be added as annotations in the command struct comments. Prioritize `bk build create`, `bk pipeline create`, `bk pipeline convert`, and `bk api`.

**4. Add `bk completion` for shell auto-completion (High)**
Kong supports shell completion generation. Enabling `bk completion bash`, `bk completion zsh`, and `bk completion fish` would dramatically improve the daily-use experience, especially for commands where pipeline slugs, cluster IDs, and org names must be typed exactly.

**5. Extend `-o json` to non-list commands (Medium)**
`bk build create`, `bk pipeline create`, `bk pipeline copy`, and `bk build view` should accept `-o json` to return the full created/retrieved resource as JSON. Without this, programmatic workflows must parse human-readable output or make a separate `bk api` call after every create operation.

**6. Standardize flag naming for shared concepts (Medium)**
`--pipeline` and `-p <pipeline>` should be unified across all commands that accept a pipeline slug. Similarly, `--branch`, `--message`, and `--commit` should use consistent names wherever they appear. A naming audit against the guidelines' recommendation for a shared flag vocabulary would catch these inconsistencies systematically.

**7. Add short forms for frequently-typed flags (Medium)**
`-b / --branch`, `-m / --message`, `-l / --limit`, and `-c / --commit` are typed in a large percentage of interactive sessions. Adding short forms reduces typing friction without removing the self-documenting long forms.

**8. Add `--no-color` / respect `NO_COLOR` (Medium)**
Follow the `NO_COLOR` convention (https://no-color.org) and add `--no-color` as a flag alias. This is a low-effort change that matters in CI environments, terminals without color support, and accessibility contexts.

**9. Document exit codes (Medium)**
Publish a table of exit codes in the documentation and, where possible, in the `--help` output. Minimum viable set: `0` success, `1` usage/flag error, `2` API/network error, `3` not found. Predictable exit codes are required for reliable CI scripting.

**10. Add `--profile` / multi-org context switching (Low-medium)**
Introduce a `--org` global flag (or `--profile` for named config profiles) to allow users managing multiple Buildkite organizations to switch context per-invocation without re-authenticating. This follows the pattern established by the GitHub CLI (`gh --hostname`) and AWS CLI (`aws --profile`).

---

## Comparison with chunk CLI

Both tools share the same underlying posthog-go telemetry issue and the same `--json` / `--quiet` gaps for machine-readable output. However, `bk` is substantially more mature:

| Dimension | bk | chunk |
|---|---|---|
| Command depth | 2 levels (strict) | 3 levels (hook subgroup) |
| JSON output | Partial (list only) | None |
| Naming consistency | Good, with `configure`/`config` conflict | Verb-noun inversion on `build-prompt` |
| Distribution quality | Excellent (all package managers + brew) | Good (goreleaser binaries) |
| Shell completion | Missing | Missing |
| Telemetry disclosure | None | None |
| Auth model | Keychain-backed, env var override | Simple API key in config |
| Escape hatch | `bk api` (excellent) | None |

`bk`'s primary design advantage is its consistent resource-verb structure and the `bk api` escape hatch. Its primary deficit compared to other mature platform CLIs (GitHub CLI, Heroku CLI) is the telemetry gap and the absence of shell completion.

---

*Assessment based on: GitHub repository at https://github.com/buildkite/cli (main branch, April 2026), Buildkite CLI documentation at https://buildkite.com/docs/platform/cli, release changelog (v3.31.1–v3.35.2), and the CLI design guidelines in this folder.*
