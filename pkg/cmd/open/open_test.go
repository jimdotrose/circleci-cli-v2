package open_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdopen "github.com/CircleCI-Public/circleci-cli/pkg/cmd/open"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

func newFactory(ios *iostreams.IOStreams) *cmdutil.Factory {
	f := cmdutil.New()
	f.IOStreams = ios
	return f
}

func noop(url string) error { return nil }

func runCmd(t *testing.T, args []string) (string, error) {
	t.Helper()
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdopen.NewCmdOpen(f, noop)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestOpenHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdopen.NewCmdOpen(f, noop)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "open-help.txt", out.String())
}

func TestOpenExplicitProject(t *testing.T) {
	output, err := runCmd(t, []string{"--project", "github/myorg/myrepo", "--no-browser"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://app.circleci.com/pipelines/github/myorg/myrepo\n"
	if output != want {
		t.Errorf("got %q, want %q", output, want)
	}
}

func TestOpenMissingProject(t *testing.T) {
	// Run outside of any git repo so inference fails.
	dir := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(orig) //nolint:errcheck

	_, err := runCmd(t, []string{"--no-browser"})
	if err == nil {
		t.Fatal("expected error for missing project slug, got nil")
	}
	if !strings.Contains(err.Error(), "project slug") {
		t.Errorf("error message should mention project slug, got: %v", err)
	}
}

func TestParseRemoteURL(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"https://github.com/myorg/myrepo.git", "github/myorg/myrepo"},
		{"https://github.com/myorg/myrepo", "github/myorg/myrepo"},
		{"git@github.com:myorg/myrepo.git", "github/myorg/myrepo"},
		{"git@github.com:myorg/myrepo", "github/myorg/myrepo"},
		{"https://bitbucket.org/myorg/myrepo.git", "bitbucket/myorg/myrepo"},
		{"git@bitbucket.org:myorg/myrepo.git", "bitbucket/myorg/myrepo"},
	}
	for _, tc := range cases {
		got, err := cmdopen.ParseRemoteURL(tc.raw)
		if err != nil {
			t.Errorf("ParseRemoteURL(%q) error: %v", tc.raw, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseRemoteURL(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
