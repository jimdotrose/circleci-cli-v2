package config

import (
	"os"

	keyring "github.com/zalando/go-keyring"
)

const keychainService = "circleci-cli"
const keychainUser = "token"

// keychainConfig wraps fileConfig and stores the API token in the OS keychain
// instead of in ~/.circleci/cli.yml. All other settings (host, telemetry, etc.)
// continue to live in the file.
//
// Token priority:  CIRCLECI_TOKEN env > CIRCLECI_CLI_TOKEN env > keychain > file (migration)
//
// If the OS keychain is unavailable (e.g. headless Linux without a secret
// service daemon) Set("token") transparently falls back to the file, so the
// CLI degrades gracefully in CI environments.
type keychainConfig struct {
	file           *fileConfig
	tokenInKeychain bool // true when the last Set("token") succeeded via keychain
}

// LoadWithKeychain loads the config file at path and wraps it in a
// keychainConfig. On load it probes the keychain so TokenBackend() reports
// correctly before any Set call.
func LoadWithKeychain(path string) (Config, error) {
	fc := &fileConfig{path: path}
	fc.setDefaults()

	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	if fc2, ok := cfg.(*fileConfig); ok {
		kc := &keychainConfig{file: fc2}
		// Probe: if a token is already in the keychain we're in keychain mode.
		if t, kerr := keyring.Get(keychainService, keychainUser); kerr == nil && t != "" {
			kc.tokenInKeychain = true
		}
		return kc, nil
	}
	return cfg, nil
}

// TokenBackend reports where the effective token is stored.
// Returns "env", "keychain", "file", or "none".
func TokenBackend(cfg Config) string {
	if os.Getenv("CIRCLECI_TOKEN") != "" || os.Getenv("CIRCLECI_CLI_TOKEN") != "" {
		return "env"
	}
	if kc, ok := cfg.(*keychainConfig); ok && kc.tokenInKeychain {
		return "keychain"
	}
	if cfg.Token() != "" {
		return "file"
	}
	return "none"
}

// Token returns the effective API token respecting the full priority chain.
func (c *keychainConfig) Token() string {
	if t := os.Getenv("CIRCLECI_TOKEN"); t != "" {
		return t
	}
	if t := os.Getenv("CIRCLECI_CLI_TOKEN"); t != "" {
		return t
	}
	// Keychain takes priority over file.
	if t, err := keyring.Get(keychainService, keychainUser); err == nil && t != "" {
		return t
	}
	// File fallback — supports migration from v1 without forcing re-login.
	return c.file.d.Token
}

func (c *keychainConfig) Host() string    { return c.file.Host() }
func (c *keychainConfig) Keys() []string  { return c.file.Keys() }
func (c *keychainConfig) Path() string    { return c.file.Path() }

func (c *keychainConfig) Get(key string) (string, bool) {
	if key == "token" {
		t := c.Token()
		return t, true
	}
	return c.file.Get(key)
}

func (c *keychainConfig) Set(key, value string) error {
	if key != "token" {
		return c.file.Set(key, value)
	}

	if value == "" {
		// Logout: remove from both keychain and file.
		_ = keyring.Delete(keychainService, keychainUser)
		c.file.d.Token = ""
		c.tokenInKeychain = false
		return nil
	}

	// Try keychain first.
	if err := keyring.Set(keychainService, keychainUser, value); err == nil {
		// Clear any plaintext token from the file to avoid leaving credentials there.
		c.file.d.Token = ""
		c.tokenInKeychain = true
		return nil
	}

	// Keychain unavailable — fall back to file.
	c.file.d.Token = value
	c.tokenInKeychain = false
	return nil
}

// Save persists non-token settings to the config file. The token is never
// written to the file when it is stored in the keychain.
func (c *keychainConfig) Save() error {
	if c.tokenInKeychain {
		// Temporarily blank the in-memory token so it isn't serialised to YAML.
		saved := c.file.d.Token
		c.file.d.Token = ""
		err := c.file.Save()
		c.file.d.Token = saved
		return err
	}
	return c.file.Save()
}
