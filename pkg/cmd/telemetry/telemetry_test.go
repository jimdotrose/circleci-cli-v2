package telemetry_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdtelemetry "github.com/CircleCI-Public/circleci-cli/pkg/cmd/telemetry"
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

func runHelp(t *testing.T, args []string, golden string) {
	t.Helper()
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios, config.NewMockConfig())
	cmd := cmdtelemetry.NewCmdTelemetry(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, golden, out.String())
}

func TestTelemetryHelp(t *testing.T) { runHelp(t, []string{"--help"}, "telemetry-help.txt") }
func TestStatusHelp(t *testing.T)    { runHelp(t, []string{"status", "--help"}, "status-help.txt") }
func TestEnableHelp(t *testing.T)    { runHelp(t, []string{"enable", "--help"}, "enable-help.txt") }
func TestDisableHelp(t *testing.T)   { runHelp(t, []string{"disable", "--help"}, "disable-help.txt") }

func TestStatus_enabled(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TelemetryVal = "true"
	f := newFactory(ios, cfg)

	cmd := cmdtelemetry.NewCmdTelemetry(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "enabled") {
		t.Errorf("output = %q; want 'enabled'", out.String())
	}
}

func TestStatus_disabled(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TelemetryVal = "false"
	f := newFactory(ios, cfg)

	cmd := cmdtelemetry.NewCmdTelemetry(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "disabled") {
		t.Errorf("output = %q; want 'disabled'", out.String())
	}
}

func TestEnable_setsTelemetryTrue(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TelemetryVal = "false"
	f := newFactory(ios, cfg)

	cmd := cmdtelemetry.NewCmdTelemetry(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"enable"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelemetryVal != "true" {
		t.Errorf("telemetry = %q; want true", cfg.TelemetryVal)
	}
}

func TestDisable_setsTelemetryFalse(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TelemetryVal = "true"
	f := newFactory(ios, cfg)

	cmd := cmdtelemetry.NewCmdTelemetry(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"disable"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelemetryVal != "false" {
		t.Errorf("telemetry = %q; want false", cfg.TelemetryVal)
	}
}

func TestStatus_envVarDisables(t *testing.T) {
	t.Setenv("CIRCLECI_NO_TELEMETRY", "1")
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.TelemetryVal = "true"
	f := newFactory(ios, cfg)

	cmd := cmdtelemetry.NewCmdTelemetry(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "disabled") {
		t.Errorf("CIRCLECI_NO_TELEMETRY should disable: output = %q", out.String())
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
