package config

import (
	"fmt"
	"strconv"
)

// MockConfig is an in-memory Config implementation for use in tests.
// It is exported so test packages in other modules can use it.
type MockConfig struct {
	HostVal        string
	TokenVal       string
	UpdateCheckVal string
	TelemetryVal   string
	SaveCalled     bool
	SaveErr        error
}

// NewMockConfig returns a MockConfig with sensible defaults.
func NewMockConfig() *MockConfig {
	return &MockConfig{
		HostVal:        "https://circleci.com",
		TokenVal:       "",
		UpdateCheckVal: "true",
		TelemetryVal:   "true",
	}
}

func (m *MockConfig) Token() string { return m.TokenVal }
func (m *MockConfig) Host() string  { return m.HostVal }

func (m *MockConfig) Get(key string) (string, bool) {
	switch key {
	case "host":
		return m.HostVal, true
	case "token":
		return m.TokenVal, true
	case "update_check":
		return m.UpdateCheckVal, true
	case "telemetry":
		return m.TelemetryVal, true
	}
	return "", false
}

func (m *MockConfig) Set(key string, value string) error {
	switch key {
	case "host":
		m.HostVal = value
	case "token":
		m.TokenVal = value
	case "update_check":
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("update_check requires true or false, got %q", value)
		}
		m.UpdateCheckVal = value
	case "telemetry":
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("telemetry requires true or false, got %q", value)
		}
		m.TelemetryVal = value
	default:
		return fmt.Errorf("unknown setting %q", key)
	}
	return nil
}

func (m *MockConfig) Keys() []string {
	return []string{"host", "token", "update_check", "telemetry"}
}

func (m *MockConfig) Save() error {
	m.SaveCalled = true
	return m.SaveErr
}

func (m *MockConfig) Path() string { return "/mock/.circleci/cli.yml" }
