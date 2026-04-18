# Migration Guide: circleci CLI v1 → v2

This guide maps every command from the legacy `circleci` CLI (v0.x) to its
equivalent in the new CLI (v2). The new CLI is a full rewrite with consistent
`<noun> <verb>` command ordering, machine-readable `--json` output on every
data-returning command, and structured error exit codes.

---

## Quick reference

| Old command | New command | Notes |
|---|---|---|
| `circleci setup` | `circleci auth login` | |
| `circleci config validate` | `circleci config validate` | Same |
| `circleci config process` | `circleci config process` | Same |
| `circleci config pack` | `circleci config pack` | Same |
| `circleci local execute` | _not available_ | Docker-based local execution removed |
| `circleci diagnostic` | `circleci diagnostic` | Same |
| `circleci update` | _not available_ | Use package manager or GitHub releases |
| `circleci namespace create` | `circleci namespace create` | Same |
| `circleci orb create` | _not available_ | Use `circleci api /orb` |
| `circleci orb list` | _not available_ | Use `circleci api /orb` |
| `circleci orb validate` | `circleci config validate` | Config validate covers orb YAML |
| `circleci context list` | `circleci context list --org-slug <vcs/org>` | Requires `--org-slug` |
| `circleci context create` | `circleci context create <name> --org-id <id>` | |
| `circleci context delete` | `circleci context delete <id>` | |
| `circleci context show` | `circleci context show <id>` | |
| `circleci context store-secret` | `circleci context secret set` | Deprecated shim included |
| `circleci context remove-secret` | `circleci context secret remove` | |
| `circleci project list-followed` | `circleci project list` | |
| `circleci project follow` | `circleci project follow <slug>` | |
| `circleci project environment-variable list` | `circleci project env list <slug>` | |
| `circleci project environment-variable create` | `circleci project env set <slug> <name> --value <v>` | |
| `circleci project environment-variable delete` | `circleci project env delete <slug> <name>` | |
| `circleci runner resource-class list` | `circleci runner resource-class list` | Also: `circleci resource-class list` |
| `circleci runner resource-class create` | `circleci runner resource-class create <name>` | |
| `circleci runner resource-class delete` | `circleci runner resource-class delete <name>` | |
| `circleci runner token list` | `circleci runner token list --resource-class <rc>` | |
| `circleci runner token create` | `circleci runner token create --resource-class <rc>` | |
| `circleci runner token delete` | `circleci runner token delete <id>` | |
| `circleci runner instance list` | `circleci runner instance list --resource-class <rc>` | |
| `circleci policy push` | `circleci policy push <dir> --owner-id <id>` | |
| `circleci policy diff` | `circleci policy diff <dir> --owner-id <id>` | |
| `circleci policy fetch` | `circleci policy fetch --owner-id <id>` | |
| `circleci policy logs` | `circleci policy logs --owner-id <id>` | |
| `circleci policy decide` | `circleci policy decide <config> --owner-id <id>` | |
| `circleci policy eval` | `circleci policy eval <config> --owner-id <id> --bundle <path>` | |
| `circleci policy settings` | `circleci policy settings --owner-id <id>` | |
| `circleci policy test` | `circleci policy test <dir> --owner-id <id>` | |

---

## Changed behaviours

### Authentication

**v1:** `circleci setup` prompted for host and token interactively, storing
both in `~/.circleci/cli.yml`.

**v2:**
```sh
# Set host first (skip for circleci.com):
circleci settings set host https://circleci.mycompany.com

# Store token:
circleci auth login
# or non-interactively:
CIRCLECI_TOKEN=mytoken circleci auth login --no-prompt
```

### Context secrets

**v1:** `circleci context store-secret <context-name> <variable-name>`

The old command accepted context **name** and prompted for the value.

**v2:** `circleci context secret set <context-id> <variable-name>`

The new command requires the context **UUID** (from `circleci context list`),
and reads the value from stdin (masked prompt when interactive, plain stdin
when piped).

```sh
# Get context ID:
circleci context list --org-slug github/myorg

# Set the secret:
circleci context secret set 00000000-0000-0000-0000-000000000000 MY_VAR
# or via stdin:
echo "$MY_SECRET" | circleci context secret set <ctx-id> MY_VAR
```

A deprecated shim at `circleci context store-secret` prints a warning and
exits; update your scripts to use the new path.

### Project environment variables

**v1:** `circleci project environment-variable create <slug> <name> <value>`

**v2:** `circleci project env set <slug> <name> --value <value>`

The positional value argument was moved to `--value` to support reading from
stdin safely:

```sh
# Old:
circleci project environment-variable create github/myorg/myrepo KEY VALUE

# New:
circleci project env set github/myorg/myrepo KEY --value VALUE

# With stdin (avoids shell history):
echo "$SECRET" | circleci project env set github/myorg/myrepo KEY --value -
```

A deprecated shim at `circleci project environment-variable` prints a warning.

### Config validation

Both v1 and v2 support `circleci config validate [<file>]`. The new CLI exits
with code **7** (validation failure) on invalid config, making it easier to
script:

```sh
circleci config validate
case $? in
  0) echo "Valid" ;;
  7) echo "Invalid — fix the errors above" ;;
  3) echo "Auth required: circleci auth login" ;;
esac
```

### Machine-readable output

Every data-returning command in v2 supports `--json`, `--jq`, and `--template`:

```sh
# v1: no standard machine-readable output
circleci context list

# v2: full JSON, jq filtering, Go templates
circleci context list --org-slug github/myorg --json
circleci context list --org-slug github/myorg --jq '.[].name'
circleci context list --org-slug github/myorg \
  --template '{{range .}}{{.name}}\n{{end}}'
```

### Exit codes

v2 uses documented, stable exit codes. See `circleci help exit-codes` for the
full table. Key codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Bad arguments |
| 3 | Auth error (run `circleci auth login`) |
| 4 | API error |
| 5 | Not found |
| 6 | Cancelled (Ctrl+C) |
| 7 | Validation failure |

### Commands not migrated

| Command | Reason |
|---------|--------|
| `circleci local execute` | Docker-based local runner was removed; use CircleCI's hosted runners or self-hosted runner agent |
| `circleci update` | Manage the CLI via your package manager (`brew upgrade circleci`) or download from GitHub Releases |
| `circleci orb *` | Orb publishing commands are not included; use `circleci api` for raw API access or the Orb Registry UI |

---

## Environment variables

All v1 env vars are still supported:

| Variable | Description |
|---|---|
| `CIRCLECI_TOKEN` | API authentication token (preferred) |
| `CIRCLECI_CLI_TOKEN` | Legacy alias for `CIRCLECI_TOKEN` |
| `CIRCLECI_HOST` | CircleCI host URL (default: `https://circleci.com`) |

New env vars in v2:

| Variable | Description |
|---|---|
| `CIRCLECI_DEBUG` | Enable HTTP request/response debug logging |
| `CIRCLECI_NO_COLOR` | Disable ANSI colour output |
| `CIRCLECI_QUIET` | Suppress progress and informational messages |
| `CIRCLECI_NO_INTERACTIVE` | Disable interactive prompts (same effect as `CI=true`) |
| `CIRCLECI_NO_TELEMETRY` | Disable anonymous usage telemetry |

See `circleci help environment` for the full list with precedence rules.
