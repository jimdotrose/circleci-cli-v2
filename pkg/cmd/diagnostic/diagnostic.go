package diagnostic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdDiagnostic returns the `circleci diagnostic` command.
func NewCmdDiagnostic(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "diagnostic",
		Short: "Check configuration and API connectivity",
		Long: heredoc.Doc(`
			Verify the CircleCI CLI configuration and connectivity.

			Checks performed:
			  1. Configuration file — exists and is readable
			  2. Authentication    — an API token is configured
			  3. API connectivity  — the configured host responds and accepts the token

			Use this command to diagnose authentication or connectivity issues
			before filing a bug report or contacting support.
		`),
		Example: heredoc.Doc(`
			# Run a full diagnostic:
			$ circleci diagnostic

			# Check connectivity to a CircleCI Server instance:
			$ circleci diagnostic --host https://circleci.mycompany.com

			# Use in CI to fail fast on misconfiguration:
			$ circleci diagnostic || exit 1
		`),
		Annotations: map[string]string{"group": "developer"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams
			out := ios.Out

			cfg, err := f.Config()
			if err != nil {
				fmt.Fprintf(out, "✗ Configuration file: could not load (%v)\n", err)
				return cierrors.New("CONFIG_ERROR", "Configuration error",
					err.Error(), cierrors.ExitGeneralError)
			}

			// ── Check 1: config file ──────────────────────────────────────────
			fmt.Fprintf(out, "✓ Configuration file: %s\n", cfg.Path())
			fmt.Fprintf(out, "  Host: %s\n", cfg.Host())

			// ── Check 2: token ────────────────────────────────────────────────
			token := cfg.Token()
			if token == "" {
				fmt.Fprintln(out, "✗ Authentication: no token configured")
				fmt.Fprintln(out, "  Run: circleci auth login")
				return cierrors.ErrAuthRequired
			}
			fmt.Fprintln(out, "✓ Authentication: token configured")

			// ── Check 3: API connectivity ─────────────────────────────────────
			stop := ios.StartSpinner("Checking API connectivity...")
			login, apiErr := callMe(cfg.Host(), token)
			stop()

			if apiErr != nil {
				fmt.Fprintf(out, "✗ API connectivity: %v\n", apiErr)
				return apiErr
			}

			fmt.Fprintf(out, "✓ API connectivity: connected as %s\n", login)
			fmt.Fprintln(out, "\nAll checks passed.")
			return nil
		},
	}
}

// callMe performs GET {host}/api/v2/me and returns the login name.
func callMe(host, token string) (string, error) {
	url := strings.TrimRight(host, "/") + "/api/v2/me"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", cierrors.New("REQUEST_ERROR", "Could not build request",
			err.Error(), cierrors.ExitAPIError)
	}
	req.Header.Set("Circle-Token", token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", cierrors.New("CONNECT_ERROR", "Could not connect to API",
			fmt.Sprintf("connecting to %s: %v", host, err), cierrors.ExitAPIError)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", cierrors.New(
			"INVALID_TOKEN",
			"Invalid or expired token",
			fmt.Sprintf("The token was rejected by %s.", host),
			cierrors.ExitAuthError,
		).WithSuggestions("Run: circleci auth login")
	}
	if resp.StatusCode != http.StatusOK {
		return "", cierrors.New("API_ERROR", fmt.Sprintf("API error %d", resp.StatusCode),
			fmt.Sprintf("unexpected response %d from %s", resp.StatusCode, host), cierrors.ExitAPIError)
	}

	var body struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", cierrors.New("DECODE_ERROR", "Could not decode API response",
			err.Error(), cierrors.ExitAPIError)
	}
	return body.Login, nil
}
