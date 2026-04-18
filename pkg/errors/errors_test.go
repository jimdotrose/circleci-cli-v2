package errors_test

import (
	"fmt"
	"strings"
	"testing"

	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

func TestCLIError_Error(t *testing.T) {
	err := cierrors.New("AUTH_REQUIRED", "Authentication required", "Your API token is missing.", cierrors.ExitAuthError).
		WithSuggestions("Run: circleci auth login", "Or set CIRCLECI_TOKEN environment variable").
		WithRef("https://circleci.com/docs/local-cli/")

	got := err.Error()
	for _, want := range []string{
		"Error [AUTH_REQUIRED]:",
		"Authentication required",
		"Your API token is missing.",
		"→ Run: circleci auth login",
		"Documentation: https://circleci.com/docs/local-cli/",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Error() missing %q in:\n%s", want, got)
		}
	}
}

func TestGetExitCode(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{nil, cierrors.ExitSuccess},
		{fmt.Errorf("plain error"), cierrors.ExitGeneralError},
		{cierrors.New("X", "title", "msg", cierrors.ExitAuthError), cierrors.ExitAuthError},
		{cierrors.New("X", "title", "msg", cierrors.ExitNotFound), cierrors.ExitNotFound},
		{cierrors.New("X", "title", "msg", cierrors.ExitValidationFail), cierrors.ExitValidationFail},
	}
	for _, c := range cases {
		got := cierrors.GetExitCode(c.err)
		if got != c.want {
			t.Errorf("GetExitCode(%v) = %d; want %d", c.err, got, c.want)
		}
	}
}
