package root

import "github.com/MakeNowJust/heredoc"

// ── circleci help environment ─────────────────────────────────────────────────

const environmentHelpTitle = "Environment variables used by circleci"

var environmentHelpBody = heredoc.Doc(`
	ENVIRONMENT VARIABLES

	All environment variables recognised by the circleci CLI. Variables override
	the corresponding flag when both are set. Flags always take final precedence.

	AUTHENTICATION
	  CIRCLECI_TOKEN           API personal access token. Preferred over the
	                           legacy CIRCLECI_CLI_TOKEN.
	  CIRCLECI_CLI_TOKEN       Legacy API token alias; both are accepted.
	  CIRCLECI_HOST            CircleCI host URL. Default: https://circleci.com
	                           Set this for CircleCI Server (self-hosted) instances.

	OUTPUT CONTROL
	  NO_COLOR                 Disable ANSI color output. Follows no-color.org standard.
	  CIRCLECI_NO_COLOR        CircleCI-specific alias for NO_COLOR.
	  CIRCLECI_QUIET           Suppress all decorative output including success
	                           confirmations. Data written to stdout is unaffected.
	                           Equivalent to passing --quiet on every command.
	  CLICOLOR=0               Disable color (Heroku/BSD convention).
	  CLICOLOR_FORCE=1         Force color even when stdout is not a TTY.
	  TERM=dumb                Disables color (inherited from terminal emulator).

	AUTOMATION / CI MODE
	  CI                       When set (any value), enables CI mode:
	                             - All interactive prompts are suppressed.
	                             - Animated spinners are replaced with plain text.
	                             - Update notifications are suppressed.
	                           Set automatically by GitHub Actions, CircleCI,
	                           Jenkins, Buildkite, Travis, and most CI systems.
	  CIRCLECI_NO_INTERACTIVE  Explicitly suppress prompts (same effect as CI=true).

	PROGRESS
	  CIRCLECI_SPINNER_DISABLED  Replace animated spinners with plain-text progress
	                             lines. Useful when capturing stderr.

	UPDATES
	  CIRCLECI_NO_UPDATE_NOTIFIER  Suppress version update nag messages.

	DEBUGGING
	  CIRCLECI_DEBUG           Log HTTP requests and responses to stderr.

	TELEMETRY
	  CIRCLECI_NO_TELEMETRY    Disable anonymous usage telemetry.
	  NO_ANALYTICS             Alias that also disables telemetry.
	  DO_NOT_TRACK             Alias that also disables telemetry.

	PRECEDENCE ORDER
	  CLI flags  >  CIRCLECI_* env vars  >  project .circleci/cli.yml
	           >  ~/.circleci/cli.yml    >  built-in defaults
`)

// ── circleci help exit-codes ──────────────────────────────────────────────────

const exitCodesHelpTitle = "Exit codes returned by circleci"

var exitCodesHelpBody = heredoc.Doc(`
	EXIT CODES

	All exit codes returned by the circleci CLI. Use these in scripts to
	branch on specific error conditions without parsing output text.

	  0   Success                Command completed successfully.
	  1   General error          Unclassified error; see the error message.
	  2   Bad arguments          Invalid flags, unknown arguments, or misuse.
	  3   Auth error             API token is missing or invalid.
	                             Run: circleci auth login
	  4   API error              CircleCI API returned a 4xx or 5xx response.
	  5   Not found              The requested resource does not exist.
	  6   Cancelled              Operation was cancelled (Ctrl+C / SIGINT).
	  7   Validation failed      Config or policy validation produced errors.
	  8   Timeout                The operation exceeded its time limit.

	SCRIPTING EXAMPLES

	  # Gate a pipeline on config validity:
	  circleci config validate .circleci/config.yml
	  case $? in
	    0) echo "Config valid" ;;
	    7) echo "Config has errors — check output above" ;;
	    3) echo "Not authenticated — run: circleci auth login" ;;
	  esac

	  # Distinguish missing resource from auth failure:
	  circleci pipeline get "$ID"
	  STATUS=$?
	  [ $STATUS -eq 5 ] && echo "Pipeline not found"
	  [ $STATUS -eq 3 ] && echo "Authenticate first: circleci auth login"
`)

// ── circleci help formatting ──────────────────────────────────────────────────

const formattingHelpTitle = "Output formatting: --json, --jq, and --template"

var formattingHelpBody = heredoc.Doc(`
	OUTPUT FORMATTING

	Every data-returning command supports three machine-readable output modes
	via flags. Use them to integrate circleci output with scripts and other tools.

	--json
	  Output the full response as pretty-printed JSON.

	    $ circleci pipeline list --json
	    $ circleci context list --json

	--jq <expression>
	  Filter the JSON output with a jq expression. --jq implies --json.
	  Uses the same syntax as the standalone jq tool.

	  Get a single field:
	    $ circleci pipeline list --jq '.[0].id'

	  Filter by condition:
	    $ circleci pipeline list --jq '.[] | select(.state=="failed") | .id'

	  Count items:
	    $ circleci pipeline list --jq 'length'

	  Extract a nested field:
	    $ circleci pipeline list --jq '.[] | {id, branch: .vcsRevision}'

	--template <go-template>
	  Format output using a Go text/template string. --template implies --json.

	    $ circleci pipeline list --template '{{range .}}{{.id}}\t{{.state}}\n{{end}}'

	JSON FIELDS
	  Every JSON-capable command lists its available fields at the bottom of its
	  --help output under "JSON Fields:". Use these names in --jq and --template
	  expressions.

	    $ circleci pipeline list --help
	    ...
	    JSON Fields:
	      id, projectSlug, state, createdAt, branch, number, vcsRevision

	OUTPUT MODE PRECEDENCE
	  --json → emit full JSON array/object
	  --jq   → implies --json, then filter with jq expression
	  --template → implies --json, then render with Go template
	  --plain → no color, tab-separated columns
	  (default) → human-formatted tables with color

	When --json is active, all human-readable output is suppressed from stdout.
	Progress messages continue on stderr in non-CI mode.
`)

// ── circleci help api ─────────────────────────────────────────────────────────

const apiHelpTitle = "Raw API access with 'circleci api'"

var apiHelpBody = heredoc.Doc(`
	RAW API ACCESS

	The 'circleci api' command lets you call any CircleCI API v2 endpoint
	directly. Authentication (Circle-Token) and base URL are applied automatically.

	USAGE

	  circleci api <endpoint> [flags]

	  The endpoint must start with / and is relative to /api/v2.

	EXAMPLES

	  # Get the authenticated user:
	  $ circleci api /me

	  # List pipelines (all pages):
	  $ circleci api /project/gh/myorg/myrepo/pipeline --paginate

	  # Create a context via POST:
	  $ circleci api /context --method POST \
	      --field name=my-ctx \
	      --field owner.id=$ORG_ID \
	      --field owner.type=organization

	  # Filter output with jq:
	  $ circleci api /project/gh/myorg/myrepo/pipeline --jq '.items[].id'

	  # Add custom headers:
	  $ circleci api /me --header 'Accept: application/json'

	FLAGS

	  -X, --method      HTTP method (default: GET, or POST when --field is used)
	  -F, --field       Request body field: key=value (repeatable; supports nested keys: owner.id)
	  -H, --header      Additional HTTP header: 'Name: value' (repeatable)
	      --paginate    Follow next_page_token and collect all pages into one array
	      --jq          Filter response with a jq expression

	NOTES

	  Use 'circleci api --help' for the full flag reference.
	  All errors follow standard circleci exit codes (see: circleci help exit-codes).
`)
