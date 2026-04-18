package iostreams_test

import (
	"os"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

func TestTest_defaults(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	if ios.IsInteractive {
		t.Error("Test() IOStreams should not be interactive")
	}
	if ios.ColorEnabled {
		t.Error("Test() IOStreams should not have color enabled")
	}
	if ios.SpinnerEnabled {
		t.Error("Test() IOStreams should not have spinner enabled")
	}
}

func TestSystem_CIMode(t *testing.T) {
	// Ensure CI=true disables interactivity and color (when not forced).
	t.Setenv("CI", "true")
	t.Setenv("CLICOLOR_FORCE", "")
	t.Setenv("NO_COLOR", "")
	t.Setenv("CIRCLECI_NO_COLOR", "")

	// System() reads os.Stdout which is not a TTY in tests — CI mode is already
	// active via non-TTY stdout. Setting CI=true makes it explicit.
	ios := iostreams.System()
	if ios.IsInteractive {
		t.Error("CI=true should disable IsInteractive")
	}
	if ios.UpdatesEnabled {
		t.Error("CI=true should disable UpdatesEnabled")
	}
}

func TestSystem_NoColor(t *testing.T) {
	for _, tc := range []struct {
		env   string
		value string
	}{
		{"NO_COLOR", "1"},
		{"CIRCLECI_NO_COLOR", "1"},
		{"CLICOLOR", "0"},
		{"TERM", "dumb"},
	} {
		t.Run(tc.env+"="+tc.value, func(t *testing.T) {
			os.Setenv(tc.env, tc.value)
			t.Cleanup(func() { os.Unsetenv(tc.env) })
			ios := iostreams.System()
			if ios.ColorEnabled {
				t.Errorf("%s=%s should disable color", tc.env, tc.value)
			}
		})
	}
}

func TestSystem_CLICOLORForce(t *testing.T) {
	t.Setenv("CLICOLOR_FORCE", "1")
	t.Setenv("NO_COLOR", "") // force takes precedence over no-color
	ios := iostreams.System()
	if !ios.ColorEnabled {
		t.Error("CLICOLOR_FORCE=1 should enable color even without a TTY")
	}
}

func TestSetColorEnabled(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.SpinnerEnabled = true // manually set to verify it gets cleared
	ios.SetColorEnabled(false)
	if ios.ColorEnabled {
		t.Error("SetColorEnabled(false) should disable color")
	}
	if ios.SpinnerEnabled {
		t.Error("SetColorEnabled(false) should disable spinner")
	}
}
