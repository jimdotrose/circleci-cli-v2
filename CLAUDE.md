# CircleCI CLI v2

A new CircleCI CLI built from scratch in Go + Cobra, targeting exemplary CLI design
(57+/60 against the design guidelines in `agents/`). Reference CLI: GitHub CLI (`gh`).

Full architecture, command surface, and phased roadmap: `docs/build-plan.md`

---

## Critical rules — read before writing any command

These are the six design decisions that must not be violated. They exist because the
current circleci CLI got all six wrong, and this project exists to fix them.

**1. Every data-returning command gets `--json` with field enumeration in `--help`.**
No exceptions. Consistent JSON coverage is the #1 differentiator between a scripting
tool and an interactive-only tool. Use the output helper in `pkg/output`.

**2. Use the structured error type in `pkg/errors`. Never `fmt.Errorf` in handlers.**
Every error must have: `code`, `title`, `message`, `suggestions[]`, `ref` (doc URL).
Exit code constants live in `pkg/errors/exitcodes.go` — always use those, never raw integers.

**3. `circleci config` = pipeline YAML. `circleci settings` = CLI tool config.**
This naming is non-negotiable. `circleci config validate` validates pipeline YAML.
`circleci settings set token <value>` manages the API token. Never mix these.

**4. Maximum 2 levels of command nesting. If you go to 3, add an alias.**
`circleci context secret set` = fine (2 levels under root).
`circleci runner resource-class list` = 3 levels → must have `circleci resource-class list` alias.
Four levels (`project environment-variable list`) must never recur — use `circleci project env list`.

**5. Every command needs `Use`, `Short`, `Long` (heredoc), and `Example` (heredoc, 3+ examples).**
Examples are "by far the most-read section of help text." Use `github.com/MakeNowJust/heredoc`
for all multi-line strings. No blank `Long` descriptions.

**6. Telemetry must be disclosed, opt-out, and auto-disabled in CI.**
On first run, print a one-time notice. Respect `CIRCLECI_NO_TELEMETRY`, `NO_ANALYTICS`,
`DO_NOT_TRACK`. When `CI=true` is set, skip the notice and disable telemetry automatically.

---

## Design guidelines

Full guidelines are in `agents/`. Start with the checklist:

```
agents/checklist.md          ← run through this before any PR
agents/01-philosophy.md      ← the 9 core principles
agents/04-output.md          ← --json, color, TTY detection
agents/05-errors.md          ← error format, exit codes
agents/06-arguments-and-flags.md ← flag naming, short forms, env vars
```

Use the `/check-standards` slash command to score the CLI against the checklist at any time.

---

## Package structure

```
pkg/
├── cmd/                  One package per top-level command group.
│   ├── root/             Root command, help topics, global flags.
│   ├── auth/             circleci auth login/logout/status/token
│   ├── config/           circleci config validate/process/pack/generate
│   ├── context/          circleci context + circleci context secret
│   ├── pipeline/         circleci pipeline list/get/trigger
│   ├── workflow/         circleci workflow list/get/cancel/rerun
│   ├── job/              circleci job list/get/cancel/artifacts
│   ├── orb/              circleci orb list/info/validate/publish/...
│   ├── project/          circleci project list/follow + project env
│   ├── runner/           circleci runner resource-class/token/instance
│   ├── policy/           circleci policy push/diff/fetch/...
│   ├── settings/         circleci settings list/get/set
│   └── api/              circleci api <endpoint> (raw API escape hatch)
│
├── iostreams/            TTY detection, color, spinner, stdout/stderr wiring.
│                         NEVER call os.Getenv("NO_COLOR") in a command — ask IOStreams.
│
├── output/               --json with field enumeration, --jq (gojq), --template, --plain.
│                         All data commands use this. Never hand-roll JSON marshalling.
│
├── errors/               Structured error type + exit code constants.
│                         exitcodes.go: ExitSuccess=0, ExitAuthError=3, ExitAPIError=4,
│                         ExitNotFound=5, ExitValidationFail=7, ExitTimeout=8
│
├── apiclient/            CircleCI REST + GraphQL client. Injected via Factory.
│
├── cmdutil/              Factory struct — wires IOStreams + Config + APIClient into commands.
│                         Every command constructor takes *cmdutil.Factory, returns *cobra.Command.
│
└── text/                 Table printing, time formatting, string helpers.
```

The **Factory pattern** is mandatory for all commands. Tests inject mock clients and
`iostreams.Test()` streams — no global state, no `os.Stdout` writes in production code.

---

## Environment variables

All documented at `circleci help environment`. Variables use `CIRCLECI_` prefix:

| Variable | Purpose |
|---|---|
| `CIRCLECI_TOKEN` | API token (also: `CIRCLECI_CLI_TOKEN` legacy alias) |
| `CIRCLECI_HOST` | CircleCI server host (default: `https://circleci.com`) |
| `CIRCLECI_NO_INTERACTIVE` | Suppress all prompts |
| `CIRCLECI_NO_COLOR` | Disable ANSI color |
| `CIRCLECI_SPINNER_DISABLED` | Replace animated spinner with plain text |
| `CIRCLECI_NO_UPDATE_NOTIFIER` | Suppress version update messages |
| `CIRCLECI_DEBUG` | Log HTTP requests to stderr |
| `CIRCLECI_NO_TELEMETRY` | Disable telemetry |
| `CI` | When set, implies NO_INTERACTIVE + disables spinner + update notifications |
| `NO_COLOR` | no-color.org standard — always respected |

---

## Exit codes

Defined in `pkg/errors/exitcodes.go`. Document new codes there before using them.

| Code | Constant | Meaning |
|---|---|---|
| 0 | `ExitSuccess` | Command succeeded |
| 1 | `ExitGeneralError` | Unclassified error |
| 2 | `ExitBadArguments` | Invalid arguments or flags |
| 3 | `ExitAuthError` | Missing or invalid API token |
| 4 | `ExitAPIError` | CircleCI API returned 4xx/5xx |
| 5 | `ExitNotFound` | Requested resource does not exist |
| 6 | `ExitCancelled` | Operation cancelled by user (Ctrl+C) |
| 7 | `ExitValidationFail` | Config or policy validation failed |
| 8 | `ExitTimeout` | Operation timed out |

---

## Common commands

```sh
go build ./cmd/circleci/...            # build binary
go test ./...                          # unit tests (no network)
UPDATE_GOLDEN=1 go test ./...          # regenerate golden help-text files
go test -tags integration ./...        # integration tests (needs CIRCLECI_TEST_TOKEN)
goreleaser build --snapshot --clean    # test multi-platform release builds
circleci --help                        # smoke test
NO_COLOR=1 circleci --help             # verify color is disabled
CI=true circleci --help                # verify CI mode
```

---

## When adding a new command

Use `/add-command <group> <verb>` to scaffold it, or follow these steps manually:

1. Create `pkg/cmd/<group>/<verb>.go`
2. Add `Use`, `Short`, `Long` (heredoc), `Example` (heredoc, 3+ examples covering simple
   and complex cases)
3. If the command returns data: declare a typed JSON struct, enumerate fields in `Long`,
   wire `--json`/`--jq`/`--template` via `pkg/output`
4. If the command mutates state: add `--force` for destructive ops; add `--dry-run` where
   preview is useful
5. All errors via `pkg/errors` — never raw strings
6. Wire the command into `pkg/cmd/<group>/<group>.go`
7. Add `<verb>_test.go` with at least one golden test for `--help` output
8. Run `/check-standards` before opening a PR

---

## When adding a new command group

Use `/new-command-group <name>` or follow the steps in that slash command's documentation.

---

## Assessments and references

```
docs/build-plan.md                   Full architecture + phased roadmap
docs/assessments/github-cli.md       GitHub CLI design assessment (the reference target)
docs/assessments/circleci-cli.md     Existing circleci CLI assessment (what we're replacing)
docs/assessments/buildkite-cli.md    Buildkite CLI assessment (peer comparison)
docs/assessments/circleci-vs-gh.md   Gap analysis + prioritized recommendations
```
