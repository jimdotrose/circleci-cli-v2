package policy_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdpolicy "github.com/CircleCI-Public/circleci-cli/pkg/cmd/policy"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

func newFactory(ios *iostreams.IOStreams) *cmdutil.Factory {
	f := cmdutil.New()
	f.IOStreams = ios
	return f
}

func runHelp(t *testing.T, args []string, golden string) {
	t.Helper()
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdpolicy.NewCmdPolicy(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, golden, out.String())
}

func TestPolicyHelp(t *testing.T)    { runHelp(t, []string{"--help"}, "policy-help.txt") }
func TestPushHelp(t *testing.T)      { runHelp(t, []string{"push", "--help"}, "push-help.txt") }
func TestDiffHelp(t *testing.T)      { runHelp(t, []string{"diff", "--help"}, "diff-help.txt") }
func TestFetchHelp(t *testing.T)     { runHelp(t, []string{"fetch", "--help"}, "fetch-help.txt") }
func TestLogsHelp(t *testing.T)      { runHelp(t, []string{"logs", "--help"}, "logs-help.txt") }
func TestDecideHelp(t *testing.T)    { runHelp(t, []string{"decide", "--help"}, "decide-help.txt") }
func TestEvalHelp(t *testing.T)      { runHelp(t, []string{"eval", "--help"}, "eval-help.txt") }
func TestSettingsHelp(t *testing.T)  { runHelp(t, []string{"settings", "--help"}, "settings-help.txt") }
func TestTestHelp(t *testing.T)      { runHelp(t, []string{"test", "--help"}, "test-help.txt") }

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
