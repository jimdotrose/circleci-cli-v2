package project_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdproject "github.com/CircleCI-Public/circleci-cli/pkg/cmd/project"
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
	cmd := cmdproject.NewCmdProject(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, golden, out.String())
}

func TestProjectHelp(t *testing.T)       { runHelp(t, []string{"--help"}, "project-help.txt") }
func TestListHelp(t *testing.T)          { runHelp(t, []string{"list", "--help"}, "list-help.txt") }
func TestFollowHelp(t *testing.T)        { runHelp(t, []string{"follow", "--help"}, "follow-help.txt") }
func TestEnvHelp(t *testing.T)           { runHelp(t, []string{"env", "--help"}, "env-help.txt") }
func TestEnvListHelp(t *testing.T)       { runHelp(t, []string{"env", "list", "--help"}, "env-list-help.txt") }
func TestEnvGetHelp(t *testing.T)        { runHelp(t, []string{"env", "get", "--help"}, "env-get-help.txt") }
func TestEnvSetHelp(t *testing.T)        { runHelp(t, []string{"env", "set", "--help"}, "env-set-help.txt") }
func TestEnvDeleteHelp(t *testing.T)     { runHelp(t, []string{"env", "delete", "--help"}, "env-delete-help.txt") }

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
