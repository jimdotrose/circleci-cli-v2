package version_test

import (
	"strings"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/version"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

func TestVersionCmd(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := cmdutil.New()
	f.IOStreams = ios

	cmd := version.NewCmdVersion(f, "1.2.3")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "1.2.3") {
		t.Errorf("version output %q missing version string", got)
	}
	if !strings.Contains(got, "circleci") {
		t.Errorf("version output %q missing binary name", got)
	}
}
