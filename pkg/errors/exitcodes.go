package errors

// Exit code constants for circleci CLI commands.
// Documented at: circleci help exit-codes
const (
	ExitSuccess        = 0 // Command succeeded
	ExitGeneralError   = 1 // General / unclassified error
	ExitBadArguments   = 2 // Invalid arguments or flags (misuse)
	ExitAuthError      = 3 // Missing or invalid API token
	ExitAPIError       = 4 // CircleCI API returned an error (4xx/5xx)
	ExitNotFound       = 5 // Requested resource does not exist
	ExitCancelled      = 6 // Operation cancelled by user (Ctrl+C)
	ExitValidationFail = 7 // Config or policy validation failed
	ExitTimeout        = 8 // Operation timed out
)
