package apiclient

import "time"

// ── Config ────────────────────────────────────────────────────────────────────

// CompileResponse is the API response from POST /compile-config-with-defaults.
type CompileResponse struct {
	Valid        bool   `json:"valid"`
	SourceYaml   string `json:"source-yaml"`
	OutputYaml   string `json:"output-yaml"`
	Errors       []struct {
		Message string `json:"message"`
		Rule    string `json:"rule,omitempty"`
	} `json:"errors"`
}

// ── Context ───────────────────────────────────────────────────────────────────

// Context represents a CircleCI context.
type Context struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ContextVariable represents an environment variable within a context.
type ContextVariable struct {
	Variable  string    `json:"variable"`
	ContextID string    `json:"context_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ── Pipeline ──────────────────────────────────────────────────────────────────

// Pipeline represents a CircleCI pipeline.
type Pipeline struct {
	ID          string    `json:"id"`
	ProjectSlug string    `json:"project_slug"`
	Number      int       `json:"number"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Trigger     struct {
		Type       string    `json:"type"`
		ReceivedAt time.Time `json:"received_at"`
		Actor      struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		} `json:"actor"`
	} `json:"trigger"`
	VCS *struct {
		Branch   string `json:"branch"`
		Tag      string `json:"tag"`
		Revision string `json:"revision"`
	} `json:"vcs"`
}

// TriggerPipelineResponse is returned by POST /project/:slug/pipeline.
type TriggerPipelineResponse struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	Number     int    `json:"number"`
	CreatedAt  time.Time `json:"created_at"`
}

// ── Workflow ──────────────────────────────────────────────────────────────────

// Workflow represents a CircleCI workflow.
type Workflow struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	PipelineID string    `json:"pipeline_id"`
	ProjectSlug string   `json:"project_slug"`
	Status     string    `json:"status"`
	StartedBy  string    `json:"started_by"`
	PipelineNumber int   `json:"pipeline_number"`
	CreatedAt  time.Time `json:"created_at"`
	StoppedAt  *time.Time `json:"stopped_at"`
}

// ── Job ───────────────────────────────────────────────────────────────────────

// Job represents a CircleCI job within a workflow.
type Job struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	ProjectSlug   string     `json:"project_slug"`
	Status        string     `json:"status"`
	Type          string     `json:"type"`
	StartedAt     *time.Time `json:"started_at"`
	StoppedAt     *time.Time `json:"stopped_at"`
	JobNumber     *int       `json:"job_number"`
	ApprovalID    string     `json:"approval_request_id,omitempty"`
	ApprovalName  string     `json:"approval_name,omitempty"`
	Dependencies  []string   `json:"dependencies"`
}

// Artifact represents a job artifact.
type Artifact struct {
	Path       string `json:"path"`
	NodeIndex  int    `json:"node_index"`
	URL        string `json:"url"`
}

// ── Project ───────────────────────────────────────────────────────────────────

// Project represents a CircleCI project.
type Project struct {
	Slug             string    `json:"slug"`
	Name             string    `json:"name"`
	OrganizationName string    `json:"organization_name"`
	VCSInfo          ProjectVCS `json:"vcs_info"`
}

// ProjectVCS holds VCS metadata for a project.
type ProjectVCS struct {
	VCSUrl        string `json:"vcs_url"`
	Provider      string `json:"provider"`
	DefaultBranch string `json:"default_branch"`
}

// EnvVar represents a project-level environment variable (value redacted by API).
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"` // always "xxxx" — API never returns real values
}

// ── Runner ────────────────────────────────────────────────────────────────────

// RunnerResourceClass represents a self-hosted runner resource class.
type RunnerResourceClass struct {
	ResourceClass string `json:"resource_class"`
	Description   string `json:"description"`
}

// RunnerToken represents an authentication token for a runner resource class.
type RunnerToken struct {
	ID            string    `json:"id"`
	ResourceClass string    `json:"resource_class"`
	Nickname      string    `json:"nickname"`
	CreatedAt     time.Time `json:"created_at"`
	Token         string    `json:"token,omitempty"` // only present at creation time
}

// RunnerInstance represents a registered self-hosted runner agent.
type RunnerInstance struct {
	ResourceClass  string    `json:"resource_class"`
	Hostname       string    `json:"hostname"`
	Name           string    `json:"name"`
	FirstConnected time.Time `json:"first_connected"`
	LastConnected  time.Time `json:"last_connected"`
	LastUsed       time.Time `json:"last_used"`
	Version        string    `json:"version"`
	IP             string    `json:"ip"`
}

// ── Policy ────────────────────────────────────────────────────────────────────

// PolicyDecision is the result of evaluating a policy bundle against config.
type PolicyDecision struct {
	Status        string              `json:"status"`
	EnabledRules  []string            `json:"enabled_rules"`
	HardFailures  []PolicyViolation   `json:"hard_failures"`
	SoftFailures  []PolicyViolation   `json:"soft_failures"`
	Reason        string              `json:"reason,omitempty"`
}

// PolicyViolation describes a single policy rule failure.
type PolicyViolation struct {
	Rule   string `json:"rule"`
	Reason string `json:"reason"`
}

// PolicyLog represents a historical policy evaluation event.
type PolicyLog struct {
	ID          string          `json:"id"`
	CreatedAt   time.Time       `json:"created_at"`
	Decision    PolicyDecision  `json:"decision"`
	Metadata    interface{}     `json:"metadata"`
}

// PolicySettings holds the policy evaluation settings for an organization.
type PolicySettings struct {
	Enabled bool `json:"enabled"`
}

// ── Trigger ───────────────────────────────────────────────────────────────────

// ScheduledTrigger represents a scheduled pipeline trigger.
type ScheduledTrigger struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ProjectSlug string      `json:"project_slug"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Timetable   interface{} `json:"timetable"`
	Actor       struct {
		ID    string `json:"id"`
		Login string `json:"login"`
	} `json:"actor"`
}

// ── Namespace ─────────────────────────────────────────────────────────────────

// Namespace represents a CircleCI orb namespace.
type Namespace struct {
	Name string `json:"name"`
}
