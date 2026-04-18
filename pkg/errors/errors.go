package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
)

// CLIError is a structured error with a machine-readable code, human title,
// exit code, optional suggestions, and optional documentation URL.
type CLIError struct {
	Code        string
	Title       string
	Message     string
	Suggestions []string
	Ref         string
	ExitCode    int
}

func (e *CLIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error [%s]: %s\n%s", e.Code, e.Title, e.Message)
	if len(e.Suggestions) > 0 {
		b.WriteString("\n\nSuggestions:")
		for _, s := range e.Suggestions {
			fmt.Fprintf(&b, "\n  → %s", s)
		}
	}
	if e.Ref != "" {
		fmt.Fprintf(&b, "\n\nDocumentation: %s", e.Ref)
	}
	return b.String()
}

// New creates a CLIError with the given code, title, message, and exit code.
func New(code, title, message string, exitCode int) *CLIError {
	return &CLIError{
		Code:     code,
		Title:    title,
		Message:  message,
		ExitCode: exitCode,
	}
}

// WithSuggestions attaches suggestions to the error and returns it for chaining.
func (e *CLIError) WithSuggestions(suggestions ...string) *CLIError {
	e.Suggestions = suggestions
	return e
}

// WithRef attaches a documentation URL and returns it for chaining.
func (e *CLIError) WithRef(ref string) *CLIError {
	e.Ref = ref
	return e
}

// GetExitCode returns the exit code from a CLIError, or ExitGeneralError for
// any other non-nil error.
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	var cliErr *CLIError
	if stderrors.As(err, &cliErr) {
		return cliErr.ExitCode
	}
	return ExitGeneralError
}
