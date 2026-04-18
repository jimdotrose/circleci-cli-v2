package errors

// ErrAuthRequired is returned by commands that need an API token when none
// is configured. It carries exit code 3 so callers can distinguish auth
// failures from general errors in scripts.
var ErrAuthRequired = New(
	"AUTH_REQUIRED",
	"Authentication required",
	"Your API token is missing or invalid.",
	ExitAuthError,
).WithSuggestions(
	"Run: circleci auth login",
	"Or set the CIRCLECI_TOKEN environment variable",
).WithRef("https://circleci.com/docs/local-cli/")
