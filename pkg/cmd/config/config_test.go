package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/CircleCI-Public/circleci-cli/internal/testutil"
	cmdconfig "github.com/CircleCI-Public/circleci-cli/pkg/cmd/config"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	"github.com/CircleCI-Public/circleci-cli/pkg/iostreams"
)

func newFactory(ios *iostreams.IOStreams) *cmdutil.Factory {
	f := cmdutil.New()
	f.IOStreams = ios
	return f
}

func TestConfigHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "config-help.txt", out.String())
}

func TestValidateHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"validate", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "validate-help.txt", out.String())
}

func TestProcessHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"process", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "process-help.txt", out.String())
}

func TestPackHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"pack", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "pack-help.txt", out.String())
}

func TestGenerateHelp(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"generate", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testutil.AssertGolden(t, "generate-help.txt", out.String())
}

func TestGenerate_createsFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "config.yml")

	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"generate", "--out", outPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("generated file not found: %v", err)
	}
	if len(data) == 0 {
		t.Error("generated config is empty")
	}
}

func TestGenerate_refusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "config.yml")
	os.WriteFile(outPath, []byte("existing"), 0644)

	ios, _, _, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"generate", "--out", outPath})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when file exists without --force")
	}
}

func TestGenerate_forceOverwrite(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "config.yml")
	os.WriteFile(outPath, []byte("existing"), 0644)

	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"generate", "--force", "--out", outPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPack_mergesYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.yml"), []byte("jobs:\n  build:\n    docker:\n      - image: cimg/base:stable\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.yml"), []byte("version: 2.1\n"), 0644)

	ios, _, out, _ := iostreams.Test()
	f := newFactory(ios)
	cmd := cmdconfig.NewCmdConfig(f)
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"pack", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if got == "" {
		t.Error("pack produced no output")
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll(filepath.Join("testdata", "golden"), 0755)
	os.Exit(m.Run())
}
