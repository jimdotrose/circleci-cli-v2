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
