package auth_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/auth"
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

// ── auth login ────────────────────────────────────────────────────────────────

func TestLogin_withToken(t *testing.T) {
	ios, stdin, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := newFactory(ios, cfg)

	stdin.WriteString("tok-test-123\n")

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"login", "--with-token"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.TokenVal != "tok-test-123" {
		t.Errorf("token = %q; want tok-test-123", cfg.TokenVal)
	}
	if !cfg.SaveCalled {
		t.Error("config was not saved")
	}
}

func TestLogin_nonInteractive_envVar(t *testing.T) {
	t.Setenv("CIRCLECI_TOKEN", "env-tok-456")

	ios, _, out, _ := iostreams.Test()
	// IsInteractive is already false in test mode.
	cfg := config.NewMockConfig()
	f := newFactory(ios, cfg)

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	// Test IOStreams already has IsInteractive=false; no --no-prompt needed.
	cmd.SetArgs([]string{"login"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.TokenVal != "env-tok-456" {
		t.Errorf("token = %q; want env-tok-456", cfg.TokenVal)
	}
}

func TestLogin_nonInteractive_noToken(t *testing.T) {
	os.Unsetenv("CIRCLECI_TOKEN")
	os.Unsetenv("CIRCLECI_CLI_TOKEN")

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := newFactory(ios, cfg)

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"login"}) // non-interactive (test IOStreams has IsInteractive=false)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no token in non-interactive mode")
	}
	if !strings.Contains(err.Error(), "AUTH_REQUIRED") {
		t.Errorf("error = %q; want AUTH_REQUIRED", err)
	}
}

func TestLogin_help(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios, config.NewMockConfig())
	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"login", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "login-help.txt", out.String())
}

// ── auth logout ───────────────────────────────────────────────────────────────

func TestLogout_notAuthenticated(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig() // no token
	f := newFactory(ios, cfg)

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"logout"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Not currently") {
		t.Errorf("output = %q; want 'Not currently'", out.String())
	}
}

func TestLogout_withYes(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TokenVal = "existing-token"
	f := newFactory(ios, cfg)

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"logout", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TokenVal != "" {
		t.Errorf("token not cleared after logout: %q", cfg.TokenVal)
	}
}

// ── auth status ───────────────────────────────────────────────────────────────

func TestStatus_authenticated(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TokenVal = "tok-abcdefgh"
	f := newFactory(ios, cfg)

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Authenticated") {
		t.Errorf("output = %q; want 'Authenticated'", out.String())
	}
	// Token should be masked (only last 4 chars).
	if strings.Contains(out.String(), "tok-abcdefgh") {
		t.Error("full token exposed in status output")
	}
	if !strings.Contains(out.String(), "fgh") {
		t.Error("last 4 chars of token missing in status output")
	}
}

func TestStatus_notAuthenticated(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig() // no token
	f := newFactory(ios, cfg)

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when not authenticated")
	}
}

// ── auth token ────────────────────────────────────────────────────────────────

func TestToken_configured(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TokenVal = "tok-xyz"
	f := newFactory(ios, cfg)

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"token"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if got != "tok-xyz" {
		t.Errorf("token output = %q; want tok-xyz (no trailing newline)", got)
	}
}

func TestToken_notConfigured(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := newFactory(ios, cfg)

	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"token"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when token not configured")
	}
}

// ── golden help tests ─────────────────────────────────────────────────────────

func TestAuthHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios, config.NewMockConfig())
	cmd := auth.NewCmdAuth(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "auth-help.txt", out.String())
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
