package context_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdcontext "github.com/CircleCI-Public/circleci-cli/pkg/cmd/context"
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
	cmd := cmdcontext.NewCmdContext(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, golden, out.String())
}

func TestContextHelp(t *testing.T)       { runHelp(t, []string{"--help"}, "context-help.txt") }
func TestListHelp(t *testing.T)          { runHelp(t, []string{"list", "--help"}, "list-help.txt") }
func TestCreateHelp(t *testing.T)        { runHelp(t, []string{"create", "--help"}, "create-help.txt") }
func TestShowHelp(t *testing.T)          { runHelp(t, []string{"show", "--help"}, "show-help.txt") }
func TestDeleteHelp(t *testing.T)        { runHelp(t, []string{"delete", "--help"}, "delete-help.txt") }
func TestSecretHelp(t *testing.T)        { runHelp(t, []string{"secret", "--help"}, "secret-help.txt") }
func TestSecretSetHelp(t *testing.T)     { runHelp(t, []string{"secret", "set", "--help"}, "secret-set-help.txt") }
func TestSecretRemoveHelp(t *testing.T)  { runHelp(t, []string{"secret", "remove", "--help"}, "secret-remove-help.txt") }

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
