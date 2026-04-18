package pipeline_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdpipeline "github.com/CircleCI-Public/circleci-cli/pkg/cmd/pipeline"
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
	cmd := cmdpipeline.NewCmdPipeline(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, golden, out.String())
}

func TestPipelineHelp(t *testing.T) { runHelp(t, []string{"--help"}, "pipeline-help.txt") }
func TestListHelp(t *testing.T)     { runHelp(t, []string{"list", "--help"}, "list-help.txt") }
func TestGetHelp(t *testing.T)      { runHelp(t, []string{"get", "--help"}, "get-help.txt") }
func TestTriggerHelp(t *testing.T)  { runHelp(t, []string{"trigger", "--help"}, "trigger-help.txt") }

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
