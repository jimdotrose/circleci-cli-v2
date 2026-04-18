package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config provides read/write access to the CLI configuration.
type Config interface {
	// Token returns the effective API token (env var > file > "").
	Token() string
	// Host returns the effective CircleCI host URL (env var > file > default).
	Host() string
	// Get returns the string representation of a setting and whether the key
	// is known. Env-var overrides are reflected in the returned value.
	Get(key string) (string, bool)
	// Set updates a setting in memory. Call Save to persist.
	Set(key string, value string) error
	// Keys returns the ordered list of known setting keys.
	Keys() []string
	// Save writes the configuration to disk.
	Save() error
	// Path returns the absolute path to the config file.
	Path() string
}

// DefaultPath returns ~/.circleci/cli.yml.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".circleci/cli.yml"
	}
	return filepath.Join(home, ".circleci", "cli.yml")
}

// knownKeys is the ordered set of recognised setting names.
var knownKeys = []string{"host", "token", "update_check", "telemetry"}

// Load reads the config file at path, applying built-in defaults for any
// missing fields. Returns a valid Config even if the file does not yet exist.
func Load(path string) (Config, error) {
	c := &fileConfig{path: path}
	c.setDefaults()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &c.d); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	// Re-apply defaults for fields that yaml left as zero values.
	if c.d.Host == "" {
		c.d.Host = "https://circleci.com"
	}
	if c.d.UpdateCheck == nil {
		t := true
		c.d.UpdateCheck = &t
	}
	if c.d.Telemetry == nil {
		t := true
		c.d.Telemetry = &t
	}
	return c, nil
}

// fileConfig is the YAML-backed implementation of Config.
type fileConfig struct {
	path string
	d    fileData
}

// fileData mirrors the YAML structure of ~/.circleci/cli.yml.
type fileData struct {
	Host        string `yaml:"host,omitempty"`
	Token       string `yaml:"token,omitempty"`
	UpdateCheck *bool  `yaml:"update_check,omitempty"`
	Telemetry   *bool  `yaml:"telemetry,omitempty"`
}

func (c *fileConfig) setDefaults() {
	t := true
	c.d = fileData{
		Host:        "https://circleci.com",
		UpdateCheck: &t,
		Telemetry:   &t,
	}
}

func (c *fileConfig) Token() string {
	if t := os.Getenv("CIRCLECI_TOKEN"); t != "" {
		return t
	}
	if t := os.Getenv("CIRCLECI_CLI_TOKEN"); t != "" {
		return t
	}
	return c.d.Token
}

func (c *fileConfig) Host() string {
	if h := os.Getenv("CIRCLECI_HOST"); h != "" {
		return h
	}
	if c.d.Host != "" {
		return c.d.Host
	}
	return "https://circleci.com"
}

func (c *fileConfig) Get(key string) (string, bool) {
	switch key {
	case "host":
		return c.Host(), true
	case "token":
		return c.Token(), true
	case "update_check":
		if c.d.UpdateCheck == nil {
			return "true", true
		}
		return strconv.FormatBool(*c.d.UpdateCheck), true
	case "telemetry":
		if c.d.Telemetry == nil {
			return "true", true
		}
		return strconv.FormatBool(*c.d.Telemetry), true
	}
	return "", false
}

func (c *fileConfig) Set(key string, value string) error {
	switch key {
	case "host":
		c.d.Host = value
	case "token":
		c.d.Token = value
	case "update_check":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("update_check requires true or false, got %q", value)
		}
		c.d.UpdateCheck = &b
	case "telemetry":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("telemetry requires true or false, got %q", value)
		}
		c.d.Telemetry = &b
	default:
		return fmt.Errorf("unknown setting %q; known settings: host, token, update_check, telemetry", key)
	}
	return nil
}

func (c *fileConfig) Keys() []string { return knownKeys }

func (c *fileConfig) Path() string { return c.path }

func (c *fileConfig) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := yaml.Marshal(&c.d)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(c.path, data, 0600); err != nil {
		return fmt.Errorf("writing config %s: %w", c.path, err)
	}
	return nil
}
