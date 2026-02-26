package schema

// Workflow represents a complete agent workflow definition
type Workflow struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Blocking    *bool             `yaml:"blocking,omitempty" json:"blocking,omitempty"` // Default: true
	Concurrency *ConcurrencyConfig `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	On          OnConfig          `yaml:"on" json:"on"`
	Env         map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Steps       []Step            `yaml:"steps" json:"steps"`
}

// IsBlocking returns whether the workflow should block on failure (default: true)
func (w *Workflow) IsBlocking() bool {
	if w.Blocking == nil {
		return true
	}
	return *w.Blocking
}

// ConcurrencyConfig controls parallel execution
type ConcurrencyConfig struct {
	Group       string `yaml:"group" json:"group"`
	MaxParallel int    `yaml:"max-parallel,omitempty" json:"max-parallel,omitempty"` // Default: 1
}

// OnConfig defines all trigger types
type OnConfig struct {
	Hooks  *HooksTrigger   `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Tool   *ToolTrigger    `yaml:"tool,omitempty" json:"tool,omitempty"`
	Tools  []ToolTrigger   `yaml:"tools,omitempty" json:"tools,omitempty"`
	File   *FileTrigger    `yaml:"file,omitempty" json:"file,omitempty"`
	Commit *CommitTrigger  `yaml:"commit,omitempty" json:"commit,omitempty"`
	Push   *PushTrigger    `yaml:"push,omitempty" json:"push,omitempty"`
}

// HooksTrigger matches agent hook events
type HooksTrigger struct {
	Types []string `yaml:"types,omitempty" json:"types,omitempty"` // preToolUse, postToolUse
	Tools []string `yaml:"tools,omitempty" json:"tools,omitempty"` // Filter by tool name
}

// ToolTrigger matches specific tools with argument filtering
type ToolTrigger struct {
	Name string            `yaml:"name" json:"name"`
	Args map[string]string `yaml:"args,omitempty" json:"args,omitempty"` // Glob patterns on arg values
	If   string            `yaml:"if,omitempty" json:"if,omitempty"`     // Expression condition
}

// FileTrigger matches file create/edit events
type FileTrigger struct {
	Types       []string `yaml:"types,omitempty" json:"types,omitempty"`             // create, edit
	Paths       []string `yaml:"paths,omitempty" json:"paths,omitempty"`             // Include patterns
	PathsIgnore []string `yaml:"paths-ignore,omitempty" json:"paths-ignore,omitempty"` // Exclude patterns
}

// CommitTrigger matches git commit events
type CommitTrigger struct {
	Paths         []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	PathsIgnore   []string `yaml:"paths-ignore,omitempty" json:"paths-ignore,omitempty"`
	Branches      []string `yaml:"branches,omitempty" json:"branches,omitempty"`
	BranchesIgnore []string `yaml:"branches-ignore,omitempty" json:"branches-ignore,omitempty"`
}

// PushTrigger matches git push events
type PushTrigger struct {
	Paths         []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	PathsIgnore   []string `yaml:"paths-ignore,omitempty" json:"paths-ignore,omitempty"`
	Branches      []string `yaml:"branches,omitempty" json:"branches,omitempty"`
	BranchesIgnore []string `yaml:"branches-ignore,omitempty" json:"branches-ignore,omitempty"`
	Tags          []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	TagsIgnore    []string `yaml:"tags-ignore,omitempty" json:"tags-ignore,omitempty"`
}

// Step represents a single step in a workflow
type Step struct {
	Name            string            `yaml:"name,omitempty" json:"name,omitempty"`
	If              string            `yaml:"if,omitempty" json:"if,omitempty"`
	Run             string            `yaml:"run,omitempty" json:"run,omitempty"`
	Shell           string            `yaml:"shell,omitempty" json:"shell,omitempty"` // pwsh, bash, sh, cmd
	Uses            string            `yaml:"uses,omitempty" json:"uses,omitempty"`   // Reusable action
	With            map[string]string `yaml:"with,omitempty" json:"with,omitempty"`   // Action inputs
	Env             map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	WorkingDirectory string           `yaml:"working-directory,omitempty" json:"working-directory,omitempty"`
	Timeout         int               `yaml:"timeout,omitempty" json:"timeout,omitempty"` // Seconds
	ContinueOnError bool              `yaml:"continue-on-error,omitempty" json:"continue-on-error,omitempty"`
}

// Event represents the runtime event context passed to workflows
type Event struct {
	Hook      *HookEvent   `json:"hook,omitempty"`
	Tool      *ToolEvent   `json:"tool,omitempty"`
	File      *FileEvent   `json:"file,omitempty"`
	Commit    *CommitEvent `json:"commit,omitempty"`
	Push      *PushEvent   `json:"push,omitempty"`
	Cwd       string       `json:"cwd"`
	Timestamp string       `json:"timestamp"`
}

// HookEvent contains hook-specific event data
type HookEvent struct {
	Type string     `json:"type"` // preToolUse, postToolUse
	Tool *ToolEvent `json:"tool"`
	Cwd  string     `json:"cwd"`
}

// ToolEvent contains tool invocation data
type ToolEvent struct {
	Name     string                 `json:"name"`
	Args     map[string]interface{} `json:"args"`
	HookType string                 `json:"hook_type,omitempty"`
}

// FileEvent contains file change data
type FileEvent struct {
	Path    string `json:"path"`
	Action  string `json:"action"` // create, edit
	Content string `json:"content,omitempty"`
}

// CommitEvent contains git commit data
type CommitEvent struct {
	SHA     string       `json:"sha"`
	Message string       `json:"message"`
	Author  string       `json:"author"`
	Files   []FileStatus `json:"files"`
}

// PushEvent contains git push data
type PushEvent struct {
	Ref     string        `json:"ref"`
	Before  string        `json:"before"`
	After   string        `json:"after"`
	Commits []CommitEvent `json:"commits"`
}

// FileStatus represents a file's status in a commit
type FileStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"` // added, modified, deleted
}

// WorkflowResult represents the outcome of running a workflow
type WorkflowResult struct {
	PermissionDecision       string `json:"permissionDecision"` // allow, deny
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
}

// NewAllowResult creates an allow result
func NewAllowResult() *WorkflowResult {
	return &WorkflowResult{PermissionDecision: "allow"}
}

// NewDenyResult creates a deny result with a reason
func NewDenyResult(reason string) *WorkflowResult {
	return &WorkflowResult{
		PermissionDecision:       "deny",
		PermissionDecisionReason: reason,
	}
}
