package runner_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdrunner "github.com/CircleCI-Public/circleci-cli/pkg/cmd/runner"
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
	cmd := cmdrunner.NewCmdRunner(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, golden, out.String())
}

func TestRunnerHelp(t *testing.T)              { runHelp(t, []string{"--help"}, "runner-help.txt") }
func TestResourceClassHelp(t *testing.T)       { runHelp(t, []string{"resource-class", "--help"}, "resource-class-help.txt") }
func TestResourceClassListHelp(t *testing.T)   { runHelp(t, []string{"resource-class", "list", "--help"}, "resource-class-list-help.txt") }
func TestResourceClassCreateHelp(t *testing.T) { runHelp(t, []string{"resource-class", "create", "--help"}, "resource-class-create-help.txt") }
func TestResourceClassDeleteHelp(t *testing.T) { runHelp(t, []string{"resource-class", "delete", "--help"}, "resource-class-delete-help.txt") }
func TestTokenHelp(t *testing.T)               { runHelp(t, []string{"token", "--help"}, "token-help.txt") }
func TestTokenListHelp(t *testing.T)           { runHelp(t, []string{"token", "list", "--help"}, "token-list-help.txt") }
func TestTokenCreateHelp(t *testing.T)         { runHelp(t, []string{"token", "create", "--help"}, "token-create-help.txt") }
func TestTokenDeleteHelp(t *testing.T)         { runHelp(t, []string{"token", "delete", "--help"}, "token-delete-help.txt") }
func TestInstanceHelp(t *testing.T)            { runHelp(t, []string{"instance", "--help"}, "instance-help.txt") }
func TestInstanceListHelp(t *testing.T)        { runHelp(t, []string{"instance", "list", "--help"}, "instance-list-help.txt") }

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
