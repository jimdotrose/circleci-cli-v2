# CircleCI CLI vs GitHub CLI — Comparison and Recommendations

**CLIs compared:** `circleci` (CircleCI-Public/circleci-cli) vs `gh` (cli/cli v2.87.3)  
**Guidelines reference:** CLI Design Guidelines in this folder (clig.dev + Heroku + oclif + Thoughtworks)  
**Assessment date:** April 2026

---

## Executive Summary

The CircleCI CLI is a mature, well-structured tool with a solid foundation — Cobra + Viper, shell completion, good telemetry management, and a broad command surface covering config, context, orbs, pipelines, runners, and policy. It scores **38.5/60 (64%)** against the design guidelines used in this folder.

GitHub CLI scores **57.5/60 (96%)** across the same criteria.

The 19-point gap is not a fundamental architecture problem — it is a polish and composability gap. The circleci CLI is structurally sound but lacks the output layer, in-binary documentation, and CI automation ergonomics that lift `gh` to exemplary status. Unlike the gap between `gh` and `chunk` (CircleCI's newer internal CLI, which lacks even the scripting basics), the circleci CLI gap is recoverable without breaking changes and without restructuring the core command model.

The areas where the gap is widest, in order:

1. **Output** (−3.0): No `--no-color`, no `--quiet`, no `--jq`, JSON coverage incomplete
2. **Help and documentation** (−2.5): No examples in help text, no `circleci help environment` or `circleci help exit-codes`
3. **Arguments and flags** (−2.0): Few short forms, legacy dual-mode positional args
4. **Configuration and environment** (−2.0): Naming collision, no `circleci settings get/set`

---

## Scored Comparison

| Category | `circleci` | `gh` | Gap |
|---|---|---|---|
| 1. Foundations / Philosophy | 3.5 / 5 | 5.0 / 5 | −1.5 |
| 2. Command Structure and Naming | 3.0 / 5 | 4.5 / 5 | −1.5 |
| 3. Help and Documentation | 2.5 / 5 | 5.0 / 5 | **−2.5** |
| 4. Output | 2.0 / 5 | 5.0 / 5 | **−3.0** |
| 5. Errors | 3.0 / 5 | 4.5 / 5 | −1.5 |
| 6. Arguments and Flags | 3.0 / 5 | 5.0 / 5 | **−2.0** |
| 7. Interactivity | 4.0 / 5 | 5.0 / 5 | −1.0 |
| 8. Subcommands | 3.0 / 5 | 4.5 / 5 | −1.5 |
| 9. Robustness | 3.0 / 5 | 4.5 / 5 | −1.5 |
| 10. Configuration and Environment | 3.0 / 5 | 5.0 / 5 | **−2.0** |
| 11. Naming and Distribution | 4.0 / 5 | 5.0 / 5 | −1.0 |
| 12. Analytics / Telemetry | 4.5 / 5 | 5.0 / 5 | −0.5 |
| **Overall** | **38.5 / 60 (64%)** | **57.5 / 60 (96%)** | **−19.0** |

---

## Where the circleci CLI Already Does Well

Before examining the gaps, it is worth being precise about where the circleci CLI is genuinely strong relative to the guidelines — not just "not bad" but correct.

**Telemetry design** is the closest to exemplary status. The `circleci telemetry` subcommand makes data collection visible, `circleci telemetry disable` provides a clear opt-out, and telemetry is automatically suppressed in non-TTY environments. This is the right model: explicit, reversible, and CI-safe. It contrasts favorably with `bk` and `chunk`, which both ship posthog-go without disclosure.

**`--no-prompt` on `setup`** correctly implements the interactive/non-interactive duality recommended by Thoughtworks guidelines: prompts are used as a first-time-user onboarding affordance, with a clean escape hatch for scripted use.

**`--force` on destructive operations** is applied consistently across `context delete`, `runner resource-class delete`, and `runner token delete`. This is the right pattern.

**Shell completion via `circleci completion`** supports bash, zsh, and fish. This is table-stakes in 2026 but many CLIs still don't do it — including `chunk`.

**The `circleci run <name>` plugin system** (delegates to `circleci-<name>` in PATH) is a clean extension model. It follows the same convention as Git (`git-<name>`), which users already understand.

**Cobra + Viper combination** with the correct priority stack (flags → env vars → config file → defaults) is the right architecture choice and gives the CLI a solid base to build on.

**Legacy compatibility** — maintaining `vcs-type org-name` positional args alongside `--org-id` UUID flags — shows care for existing users during a migration.

---

## The Output Composability Gap

The single largest gap between `circleci` and `gh` is the **output composability layer**. This is where `gh` most visibly separates from every other platform CLI.

`gh`'s output system has three dimensions that work together:

```
gh run list --json status,name,headBranch \
            --jq '.[] | select(.status=="failure") | .name'
```

1. `--json <fields>` — structured output with an explicit field list (enumerated in `--help`)
2. `--jq <expression>` — inline jq filtering, eliminating a pipe to `jq` in most scripts
3. `--template <string>` — Go template formatting for custom human-readable output

The result is that `gh` commands are both inspection tools (for humans) and data sources (for scripts) without needing a separate API client. The three flags compose: `--jq` implies `--json`; `--template` implies `--json`.

The circleci CLI has `--json` on some commands (`orb list`, `orb search`, `orb info`, `context list`, `pipeline list`) but:

- **JSON coverage is incomplete**: `runner resource-class list`, `runner token list`, `runner instance list`, `project environment-variable list`, `context show`, and `policy logs` all lack confirmed `--json` support
- **No `--jq` flag**: users must pipe to `jq` externally
- **No field enumeration in `--help`**: when `--json` is available, users don't know what fields to expect without reading documentation or running the command
- **No `--quiet` / `--plain`**: no mechanism to suppress non-essential output in scripts short of redirecting stderr

This matters in practice. A CircleCI engineer trying to script "list all pipelines that failed in the last 24 hours" has to:

```sh
# With circleci (current):
circleci pipeline list --json | jq '.[] | select(.errors != null)'

# But many other relevant commands don't have --json at all, requiring gh api style workarounds or
# the CircleCI API directly via curl.
```

Whereas with `gh`, equivalent patterns work uniformly across every data-returning command.

---

## The In-Binary Documentation Gap

`gh` treats `--help` as a first-class product surface. The contrast with `circleci` is stark.

### What `gh help` provides

- **Every command has examples** in its help output. Not boilerplate — runnable, specific examples covering common cases and complex flags.
- **`gh help exit-codes`** documents `0`, `1`, `2`, and `4` as named, actionable states. Exit code `4` (auth required) is particularly useful for scripting — it lets callers distinguish "the command failed" from "the command is unauthenticated."
- **`gh help environment`** lists every supported environment variable with descriptions, precedence rules, and interaction notes. 25+ variables, all documented in one place.
- **`gh help formatting`** explains the `--json`/`--jq`/`--template` system with syntax examples.

### What `circleci --help` provides

- Standard Cobra-generated help with command listings and flag descriptions
- No `Example` sections on any command (or very sparse — none were found in the assessment)
- No `circleci help environment` page
- No `circleci help exit-codes` page
- No help topic for the `--json` output format or field listings

The guidelines (citing multiple sources) identify **examples as "by far the most read and revisited section of help text."** This is not a minor gap. For complex commands like `circleci config process`, `circleci orb publish`, `circleci context store-secret`, and `circleci policy push`, the absence of examples means users who never leave the terminal cannot learn correct usage from the tool itself.

---

## The Configuration Naming Collision

This is the circleci CLI's most unique and most painful design problem. It does not exist in `gh` or any other CLI assessed in this folder.

`circleci config` refers to **pipeline configuration** — the `.circleci/config.yml` file and its validation, processing, and packing.

But **tool configuration** — the API token, host URL, CLI preferences — is managed via `circleci setup` and stored at `~/.circleci/cli.yml`. There is no `circleci config get`, `circleci config set`, or `circleci config list` that refers to tool settings, because `circleci config` is already taken.

Compare with `gh`:

```sh
# gh tool config management — clear and unambiguous:
gh config get editor
gh config set editor vim
gh config list
```

There is no `gh` command group that could be confused with repository-level configuration — GitHub's equivalent ("repo settings") is managed via the web UI and `gh api`, not via `gh config`.

The circleci CLI needs a separate namespace for tool configuration. The assessment recommends `circleci settings` as the new group name, which would avoid the collision while being semantically accurate.

---

## The CI Automation Ergonomics Gap

`gh` is explicitly designed as a CI automation tool. The evidence is in every design decision that touches CI contexts:

- `GITHUB_TOKEN` is auto-injected by GitHub Actions runners — `gh` works in any Actions workflow with zero setup
- `GH_PROMPT_DISABLED` provides explicit non-interactive mode, independent of TTY detection
- `GH_SPINNER_DISABLED` replaces animated spinners with plain text — essential for readable CI logs
- `GH_NO_UPDATE_NOTIFIER` silences version nag messages in CI output
- `GH_DEBUG=api` logs all HTTP traffic to stderr for script debugging
- `gh run view --exit-status` makes `gh` directly usable as a CI gate

The circleci CLI has some of this (`--no-prompt` on `setup`, automatic telemetry disable in non-TTY) but the pattern is incomplete:

- No env var for non-interactive mode (`CI=true` detection is unconfirmed)
- No equivalent of `GH_SPINNER_DISABLED` for CI log cleanliness
- No equivalent of `--exit-status` for using the CLI as a pipeline gate
- No documented exit code vocabulary for callers to branch on

A CircleCI pipeline using `circleci` commands for validation or policy decisions cannot reliably distinguish auth failures from API errors from not-found errors — all produce non-zero exit codes with no documented semantics.

---

## Prioritized Recommendations

### P0 — Fix before declaring the CLI CI-ready

**P0.1: Add `--no-color` flag and `NO_COLOR` env var support**

This is the single biggest standards gap. Every CI system, log aggregator, and terminal that doesn't support ANSI codes will display escape sequences as literal characters. Should be a persistent global flag applied before any output is rendered.

```sh
circleci --no-color orb list
NO_COLOR=1 circleci config validate
```

**P0.2: Add `--quiet` / `-q` as a global persistent flag**

Scripts running in CI need to suppress informational output without redirecting stderr. This is a one-line Cobra addition. When `--quiet` is active, only errors and explicit data output appear.

**P0.3: Document exit codes as a help topic**

Add `circleci help exit-codes` documenting at minimum:

```
0  Success
1  General error
2  Authentication error (no token, invalid token)
3  API error (server-side failure, rate limit)
4  Validation error (config invalid, bad input)
5  Not found (resource does not exist)
```

This gives callers the ability to handle errors specifically. Auth errors (code 2) should always suggest `circleci setup` or `CIRCLECI_CLI_TOKEN`. Without documented codes, every non-zero exit is opaque.

**P0.4: Extend `--json` to all data-returning commands with field enumeration in `--help`**

Every command that returns structured data should support `--json`. The missing commands include:
- `runner resource-class list`
- `runner token list`
- `runner instance list`
- `project environment-variable list`
- `context show`
- `policy logs`

When `--json` is active, ALL human-readable output must be suppressed — only the JSON object on stdout. And for each JSON-capable command, the `--help` output should enumerate the available JSON fields (as `gh` does).

---

### P1 — Substantial impact, no breaking changes

**P1.1: Add examples to every command's `--help` output**

Populate Cobra's `Example` field for every non-trivial command. The guidelines identify examples as the most-used section of help text. Prioritize the commands users struggle with most:

```
# circleci config process examples:
  $ circleci config process .circleci/config.yml
  $ circleci config process - < .circleci/config.yml
  $ circleci config process --org-id <uuid> .circleci/config.yml

# circleci context store-secret examples:
  $ circleci context store-secret mycontext AWS_ACCESS_KEY_ID
  $ echo "$MY_SECRET" | circleci context store-secret mycontext MY_VAR -

# circleci orb publish examples:
  $ circleci orb publish src/orb.yml myorg/myorb@dev:first
  $ circleci orb publish increment src/orb.yml myorg/myorb patch --token $TOKEN
```

**P1.2: Add `circleci help environment` and `circleci help exit-codes` topics**

A dedicated environment variable page (modelled on `gh help environment`) lists all supported variables with descriptions, precedence rules, and CI-specific guidance. The missing variables to document include `CI` detection behavior, debug mode (`CIRCLECI_CLI_DEBUG` or equivalent), and any telemetry control env vars.

**P1.3: Improve error message quality for authentication and API failures**

Raw API error strings like `unauthorized` or `not found` should be wrapped with actionable guidance:

```
# Current:
Error: unauthorized

# Target:
Error [AUTH_REQUIRED]: Authentication required
Your API token is missing or invalid.

Suggestions:
  → Run: circleci setup
  → Or set CIRCLECI_CLI_TOKEN environment variable

Documentation: https://circleci.com/docs/local-cli/#configuring-the-cli
```

Cobra's error wrapper makes this straightforward to add per-command. Start with auth errors, which are the most common failure mode for new users.

**P1.4: Add `circleci --version` root flag alongside `circleci version` subcommand**

`circleci version` works, but users who run `circleci --version` (a universal UNIX convention) get an error. Cobra supports this in two lines. Both forms should work.

**P1.5: Add short forms for the highest-frequency global flags**

The global flags have no short forms, which is unusual for a mature CLI:

| Flag | Suggested short |
|---|---|
| `--debug` | `-d` |
| `--token` | `-T` (uppercase to avoid conflicts with per-command `-t`) |

Per-command flags that are typed constantly should also get short forms:
- `--org-id` → `-o` on context, orb, and runner commands
- `--format` / `--json` → `-j` where `--json` is a toggle flag

**P1.6: Introduce a `CI` environment variable mode**

When `CI=true` is set (standard in most CI environments), the CLI should automatically: suppress prompts, disable spinners, disable color, disable update checks. This matches what `gh` achieves through a combination of TTY detection and `GH_PROMPT_DISABLED`. The `--no-prompt` flag already exists on `setup` — the same behavior should apply globally when `CI=true` is detected.

---

### P2 — Design improvements, some breaking risk

**P2.1: Introduce `circleci settings` for tool configuration management**

The naming collision between `circleci config` (pipeline YAML management) and CLI tool settings is the most distinctive UX problem in this CLI. A user who wants to change their API token has no obvious command after initial `setup`. The fix is a new namespace:

```sh
circleci settings list           # Show all CLI tool settings
circleci settings get host       # Get a specific setting
circleci settings set host https://my-cci.example.com
circleci settings set token $TOKEN
```

This parallels `gh config get/set/list` exactly — a well-understood pattern. The `setup` command becomes a guided wrapper around `settings set` calls.

**P2.2: Flatten deep command nesting with aliases or restructuring**

The `circleci project environment-variable` path (4 levels) and `circleci runner resource-class` path (3 levels) violate the 2-level maximum the guidelines recommend. Two approaches:

Option A — Aliases: `circleci env-var list|create|delete --project <id>` as aliases for the full paths, while keeping the deep forms for backwards compatibility.

Option B — Promote to top-level: `circleci resource-class list|create|delete` as a top-level group, since resource classes are first-class objects in CircleCI's platform model.

Aliases are lower risk and can be shipped without deprecating the existing paths.

**P2.3: Add `--jq` filtering on all JSON-capable commands**

Once `--json` coverage is complete (P0.4), adding `--jq` provides inline filtering without requiring external `jq` installation. This is particularly valuable in environments where installing additional tools is constrained. The implementation is straightforward: the `--jq` flag takes a jq expression string and filters the JSON output before printing.

**P2.4: Add `--dry-run` to mutating commands**

For operations like `circleci policy push`, `circleci trigger create`, `circleci pipeline create`, and `circleci namespace create`, a dry-run mode that shows what would be sent without executing is valuable for CI validation workflows. `circleci config process` already serves this role for config — the same pattern should extend to other mutating commands.

---

## Implementation Sequencing

Given that the circleci CLI has a mature codebase and existing users, sequencing matters.

**Sprint 1 (non-breaking, high visibility):** P0.1 (`--no-color` / `NO_COLOR`), P0.2 (`--quiet`), P1.1 (examples in help), P1.4 (`--version` root flag). None of these break existing behavior or change output format.

**Sprint 2 (output layer):** P0.3 (exit code documentation), P0.4 (`--json` completion + field enumeration), P1.2 (`circleci help environment`). These require work across the full command surface but no breaking changes.

**Sprint 3 (CI ergonomics):** P1.3 (error message quality), P1.5 (short flags), P1.6 (`CI=true` mode). Adds convenience for power users and CI consumers.

**Sprint 4 (structural):** P2.1 (`circleci settings`), P2.2 (alias-based nesting relief), P2.3 (`--jq`), P2.4 (`--dry-run`). These require more design work and coordination, but can be introduced without removing existing commands.

---

## How the Two CLIs Compare on Dimension-by-Dimension Specifics

| Dimension | `gh` | `circleci` |
|---|---|---|
| JSON output | Every data command, fields in `--help` | Selected commands, fields not documented |
| jq filtering | `--jq` on every JSON command | Not available |
| `--no-color` / `NO_COLOR` | Yes (+ `CLICOLOR`, `CLICOLOR_FORCE`) | **No** |
| `--quiet` | Yes | **No** |
| Exit code vocabulary | Documented (`gh help exit-codes`) | **Undocumented** |
| Help examples | Every subcommand | **None / sparse** |
| Env var reference | `gh help environment` (25+ vars) | **Not consolidated** |
| Config management | `gh config get/set/list` | `circleci setup` only; naming collision |
| Non-interactive mode | `GH_PROMPT_DISABLED` (env var) | `--no-prompt` flag on `setup` only |
| CI spinner control | `GH_SPINNER_DISABLED` | **Not available** |
| API escape hatch | `gh api` (excellent, full REST + GraphQL) | **No equivalent** |
| Shell completion | `gh completion` (bash/zsh/fish/PowerShell) | `circleci completion` (bash/zsh/fish) |
| Error message quality | Structured, actionable | Raw API string |
| `--version` root flag | `gh --version` works | `circleci --version` **does not work** |
| Telemetry | **None** | Opt-out, auto-disabled in CI |
| Command depth (max) | 3 (in `repo` group) | 4 (`project environment-variable`) |
| Distribution | Exemplary (all platforms + runners) | Good |
| Binary attestation | Sigstore + SLSA | Not documented |

---

## The Underlying Pattern

Looking across both CLIs, the gap resolves to a single observation: `gh` was designed as a **scripting client that also works interactively**. The circleci CLI was designed as an **interactive tool that also works in scripts**. Those are not the same design philosophy, and the difference shows in every detail of the output layer, documentation, and CI ergonomics.

The circleci CLI's foundations are strong enough to change this — the Cobra + Viper architecture supports everything needed. The work is mostly additive: more `--json` coverage, more examples, more documentation, better output controls. The one structural change (the `circleci settings` namespace) is the only thing that requires new API surface and potential user education.

Unlike `chunk`, which needs its fundamental scripting contract built from scratch, the circleci CLI needs polish and consistency applied to a structure that is already largely correct.

---

*Sources: [circleci-cli-assessment.md](./circleci-cli-assessment.md) · [github-cli-assessment.md](./github-cli-assessment.md) · CLI design guidelines in this folder · [CircleCI CLI GitHub](https://github.com/CircleCI-Public/circleci-cli) · [GitHub CLI](https://github.com/cli/cli)*
