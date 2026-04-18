package open

import (
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/browser"
	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdOpen returns the `circleci open` command.
// opener is the function used to launch the browser; pass nil to use the default.
func NewCmdOpen(f *cmdutil.Factory, opener func(string) error) *cobra.Command {
	var project string

	if opener == nil {
		opener = browser.Open
	}

	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open the CircleCI dashboard for a project in the browser",
		Long: heredoc.Doc(`
			Open the CircleCI pipelines dashboard for a project in the default
			web browser.

			If --project is omitted, the project slug is inferred from the
			'origin' git remote of the current directory. Supported remotes:
			  github.com    → github/<org>/<repo>
			  bitbucket.org → bitbucket/<org>/<repo>

			The URL opened is:
			  https://app.circleci.com/pipelines/<project-slug>

			The URL is always printed to stdout so you can copy it or use it
			in scripts even when running non-interactively.
		`),
		Example: heredoc.Doc(`
			# Open dashboard, inferring project from git remote:
			$ circleci open

			# Open dashboard for an explicit project:
			$ circleci open --project github/myorg/myrepo

			# Print the URL without opening a browser:
			$ circleci open --project github/myorg/myrepo --no-browser

			# Open a Bitbucket-hosted project:
			$ circleci open --project bitbucket/myorg/myrepo
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := project
			if slug == "" {
				var err error
				slug, err = inferProjectSlug()
				if err != nil {
					return cierrors.New(
						"MISSING_ARG",
						"Could not determine project slug",
						err.Error(),
						cierrors.ExitBadArguments,
					).WithSuggestions(
						"Pass --project explicitly: --project github/myorg/myrepo",
					)
				}
			}

			dashURL := "https://app.circleci.com/pipelines/" + slug

			noBrowser, _ := cmd.Flags().GetBool("no-browser")
			fmt.Fprintln(f.IOStreams.Out, dashURL)

			if noBrowser {
				return nil
			}

			if err := opener(dashURL); err != nil {
				return cierrors.New(
					"BROWSER_ERROR",
					"Could not open browser",
					err.Error(),
					cierrors.ExitGeneralError,
				).WithSuggestions("Copy the URL above and paste it into your browser.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project slug (e.g. github/myorg/myrepo); inferred from git remote if omitted")
	cmd.Flags().Bool("no-browser", false, "Print the URL but do not open a browser")
	return cmd
}

// sshRemoteRe matches git SSH remote URLs: git@github.com:org/repo.git
var sshRemoteRe = regexp.MustCompile(`^git@([^:]+):([^/]+)/(.+?)(?:\.git)?$`)

// inferProjectSlug derives a CircleCI project slug from the origin git remote.
func inferProjectSlug() (string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("no git remote 'origin' found in current directory")
	}
	rawURL := strings.TrimSpace(string(out))
	return ParseRemoteURL(rawURL)
}

// ParseRemoteURL converts a git remote URL into a CircleCI project slug.
// Exported for use in tests.
func ParseRemoteURL(rawURL string) (string, error) {
	// SSH form: git@github.com:org/repo.git
	if m := sshRemoteRe.FindStringSubmatch(rawURL); m != nil {
		host, org, repo := m[1], m[2], m[3]
		vcs, err := vcsType(host)
		if err != nil {
			return "", err
		}
		return vcs + "/" + org + "/" + repo, nil
	}

	// HTTPS form: https://github.com/org/repo.git
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("unrecognised git remote URL: %q", rawURL)
	}
	vcs, err := vcsType(u.Host)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("could not parse org/repo from remote URL: %q", rawURL)
	}
	org := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	return vcs + "/" + org + "/" + repo, nil
}

// vcsType maps a git host to the CircleCI VCS type string.
func vcsType(host string) (string, error) {
	switch {
	case strings.Contains(host, "github.com"):
		return "github", nil
	case strings.Contains(host, "bitbucket.org"):
		return "bitbucket", nil
	default:
		return "", fmt.Errorf("unsupported VCS host %q — use --project to specify the slug directly", host)
	}
}
