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
