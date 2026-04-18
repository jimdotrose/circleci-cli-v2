package settings_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmd/settings"
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

// ── settings list ─────────────────────────────────────────────────────────────

func TestList_output(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TokenVal = "tok-secret"
	f := newFactory(ios, cfg)

	cmd := settings.NewCmdSettings(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	// Token value should be masked.
	if strings.Contains(got, "tok-secret") {
		t.Error("token value must not appear in settings list output")
	}
	if !strings.Contains(got, "[set]") {
		t.Error("settings list should show [set] for configured token")
	}
	if !strings.Contains(got, "https://circleci.com") {
		t.Errorf("settings list missing host: %s", got)
	}
}

// ── settings get ──────────────────────────────────────────────────────────────

func TestGet_host(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := newFactory(ios, cfg)

	cmd := settings.NewCmdSettings(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"get", "host"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "https://circleci.com") {
		t.Errorf("get host = %q; want default", out.String())
	}
}

func TestGet_unknownKey(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := newFactory(ios, cfg)

	cmd := settings.NewCmdSettings(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"get", "nonexistent"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for unknown key")
	}
}

// ── settings set ──────────────────────────────────────────────────────────────

func TestSet_host(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := newFactory(ios, cfg)

	cmd := settings.NewCmdSettings(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"set", "host", "https://example.com"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HostVal != "https://example.com" {
		t.Errorf("host = %q; want https://example.com", cfg.HostVal)
	}
	if !cfg.SaveCalled {
		t.Error("config was not saved")
	}
}

func TestSet_invalidBool(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := newFactory(ios, cfg)

	cmd := settings.NewCmdSettings(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"set", "telemetry", "maybe"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for invalid bool value")
	}
}

// ── golden help tests ─────────────────────────────────────────────────────────

func TestSettingsHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios, config.NewMockConfig())
	cmd := settings.NewCmdSettings(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "settings-help.txt", out.String())
}

func TestSetHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios, config.NewMockConfig())
	cmd := settings.NewCmdSettings(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"set", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "set-help.txt", out.String())
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
