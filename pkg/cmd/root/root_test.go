package root_test

import (
	"bytes"
	"os"
	"path/filepath" // used in TestMain for testdata dir creation
	"strings"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/root"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

func newTestFactory(ios *iostreams.IOStreams) *cmdutil.Factory {
	f := cmdutil.New()
	f.IOStreams = ios
	return f
}

func TestRootHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newTestFactory(ios)
	cmd := root.NewCmdRoot(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	testutil.AssertGolden(t, "help.txt", got)
}

func TestRootHelp_NoANSI(t *testing.T) {
	// NO_COLOR must produce output with no ANSI escape sequences.
	t.Setenv("NO_COLOR", "1")
	ios := iostreams.System()
	if ios.ColorEnabled {
		t.Fatal("NO_COLOR=1 should disable color")
	}

	f := newTestFactory(ios)
	cmd := root.NewCmdRoot(f)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out.String(), "\x1b[") {
		t.Error("output contains ANSI escape codes when NO_COLOR=1")
	}
}

func TestRootHelp_CIMode(t *testing.T) {
	// CI=true must disable interactivity; --help should still work.
	t.Setenv("CI", "true")
	ios := iostreams.System()
	if ios.IsInteractive {
		t.Error("CI=true should set IsInteractive=false")
	}
	if ios.UpdatesEnabled {
		t.Error("CI=true should set UpdatesEnabled=false")
	}

	f := newTestFactory(ios)
	cmd := root.NewCmdRoot(f)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "circleci") {
		t.Error("expected circleci in help output")
	}
}

func TestHelpTopic_ExitCodes(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newTestFactory(ios)
	cmd := root.NewCmdRoot(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"help", "exit-codes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"0", "3", "7", "circleci auth login", "circleci config validate"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("exit-codes help missing %q", want)
		}
	}
}

func TestHelpTopic_Environment(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newTestFactory(ios)
	cmd := root.NewCmdRoot(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"help", "environment"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"CIRCLECI_TOKEN", "NO_COLOR", "CI", "CIRCLECI_DEBUG"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("environment help missing %q", want)
		}
	}
}

func TestNoColorFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.ColorEnabled = true  // start enabled
	ios.SpinnerEnabled = true
	f := newTestFactory(ios)
	cmd := root.NewCmdRoot(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--no-color", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ios.ColorEnabled {
		t.Error("--no-color flag should disable color on IOStreams")
	}
	if ios.SpinnerEnabled {
		t.Error("--no-color flag should disable spinner on IOStreams")
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
