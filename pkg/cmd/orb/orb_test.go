package orb_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdorb "github.com/CircleCI-Public/circleci-cli/pkg/cmd/orb"
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
	cmd := cmdorb.NewCmdOrb(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, golden, out.String())
}

func TestOrbHelp(t *testing.T)      { runHelp(t, []string{"--help"}, "orb-help.txt") }
func TestListHelp(t *testing.T)     { runHelp(t, []string{"list", "--help"}, "list-help.txt") }
func TestInfoHelp(t *testing.T)     { runHelp(t, []string{"info", "--help"}, "info-help.txt") }
func TestValidateHelp(t *testing.T) { runHelp(t, []string{"validate", "--help"}, "validate-help.txt") }
func TestPublishHelp(t *testing.T)  { runHelp(t, []string{"publish", "--help"}, "publish-help.txt") }
func TestPromoteHelp(t *testing.T)  { runHelp(t, []string{"promote", "--help"}, "promote-help.txt") }
func TestSearchHelp(t *testing.T)   { runHelp(t, []string{"search", "--help"}, "search-help.txt") }

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
