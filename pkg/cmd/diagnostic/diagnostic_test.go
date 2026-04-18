package diagnostic_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/diagnostic"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/config"
	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

func newFactory(ios *iostreams.IOStreams, cfg config.Config) *cmdutil.Factory {
	f := cmdutil.New()
	f.IOStreams = ios
	f.Config = func() (config.Config, error) { return cfg, nil }
	return f
}

func fakeAPIServer(login string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == http.StatusOK {
			json.NewEncoder(w).Encode(map[string]string{"login": login})
		}
	}))
}

func TestDiagnostic_allPass(t *testing.T) {
	srv := fakeAPIServer("alice", http.StatusOK)
	defer srv.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TokenVal = "valid-token"
	cfg.HostVal = srv.URL
	f := newFactory(ios, cfg)

	cmd := diagnostic.NewCmdDiagnostic(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	for _, want := range []string{"✓", "alice", "All checks passed"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
}

func TestDiagnostic_noToken(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig() // no token
	f := newFactory(ios, cfg)

	cmd := diagnostic.NewCmdDiagnostic(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when token not configured")
	}
	if !strings.Contains(out.String(), "no token") {
		t.Errorf("output = %q; want 'no token' message", out.String())
	}
}

func TestDiagnostic_invalidToken(t *testing.T) {
	srv := fakeAPIServer("", http.StatusUnauthorized)
	defer srv.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TokenVal = "bad-token"
	cfg.HostVal = srv.URL
	f := newFactory(ios, cfg)

	cmd := diagnostic.NewCmdDiagnostic(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestDiagnostic_help(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios, config.NewMockConfig())
	cmd := diagnostic.NewCmdDiagnostic(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "diagnostic-help.txt", out.String())
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
