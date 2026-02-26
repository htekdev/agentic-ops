package trigger

import (
	"testing"

	"github.com/htekdev/agentic-ops/internal/schema"
)

func TestMatchToolTrigger(t *testing.T) {
	tests := []struct {
		name    string
		trigger *schema.ToolTrigger
		event   *schema.ToolEvent
		want    bool
	}{
		{
			name: "exact tool match",
			trigger: &schema.ToolTrigger{
				Name: "edit",
			},
			event: &schema.ToolEvent{
				Name: "edit",
				Args: map[string]interface{}{},
			},
			want: true,
		},
		{
			name: "tool name mismatch",
			trigger: &schema.ToolTrigger{
				Name: "edit",
			},
			event: &schema.ToolEvent{
				Name: "create",
				Args: map[string]interface{}{},
			},
			want: false,
		},
		{
			name: "args glob match",
			trigger: &schema.ToolTrigger{
				Name: "edit",
				Args: map[string]string{
					"path": "**/*.js",
				},
			},
			event: &schema.ToolEvent{
				Name: "edit",
				Args: map[string]interface{}{
					"path": "src/utils/helper.js",
				},
			},
			want: true,
		},
		{
			name: "args glob no match",
			trigger: &schema.ToolTrigger{
				Name: "edit",
				Args: map[string]string{
					"path": "**/*.ts",
				},
			},
			event: &schema.ToolEvent{
				Name: "edit",
				Args: map[string]interface{}{
					"path": "src/utils/helper.js",
				},
			},
			want: false,
		},
		{
			name: "missing arg",
			trigger: &schema.ToolTrigger{
				Name: "edit",
				Args: map[string]string{
					"path": "**/*.js",
				},
			},
			event: &schema.ToolEvent{
				Name: "edit",
				Args: map[string]interface{}{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &schema.Workflow{
				On: schema.OnConfig{
					Tool: tt.trigger,
				},
			}
			matcher := NewMatcher(workflow)
			event := &schema.Event{
				Tool: tt.event,
			}
			if got := matcher.Match(event); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchHooksTrigger(t *testing.T) {
	tests := []struct {
		name    string
		trigger *schema.HooksTrigger
		event   *schema.HookEvent
		want    bool
	}{
		{
			name: "match hook type",
			trigger: &schema.HooksTrigger{
				Types: []string{"preToolUse"},
			},
			event: &schema.HookEvent{
				Type: "preToolUse",
			},
			want: true,
		},
		{
			name: "no match hook type",
			trigger: &schema.HooksTrigger{
				Types: []string{"postToolUse"},
			},
			event: &schema.HookEvent{
				Type: "preToolUse",
			},
			want: false,
		},
		{
			name: "match with tool filter",
			trigger: &schema.HooksTrigger{
				Types: []string{"preToolUse"},
				Tools: []string{"edit", "create"},
			},
			event: &schema.HookEvent{
				Type: "preToolUse",
				Tool: &schema.ToolEvent{Name: "edit"},
			},
			want: true,
		},
		{
			name: "no match tool filter",
			trigger: &schema.HooksTrigger{
				Types: []string{"preToolUse"},
				Tools: []string{"edit", "create"},
			},
			event: &schema.HookEvent{
				Type: "preToolUse",
				Tool: &schema.ToolEvent{Name: "powershell"},
			},
			want: false,
		},
		{
			name: "empty types matches all",
			trigger: &schema.HooksTrigger{
				Tools: []string{"edit"},
			},
			event: &schema.HookEvent{
				Type: "preToolUse",
				Tool: &schema.ToolEvent{Name: "edit"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &schema.Workflow{
				On: schema.OnConfig{
					Hooks: tt.trigger,
				},
			}
			matcher := NewMatcher(workflow)
			event := &schema.Event{
				Hook: tt.event,
			}
			if got := matcher.Match(event); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchFileTrigger(t *testing.T) {
	tests := []struct {
		name    string
		trigger *schema.FileTrigger
		event   *schema.FileEvent
		want    bool
	}{
		{
			name: "match file type",
			trigger: &schema.FileTrigger{
				Types: []string{"edit"},
			},
			event: &schema.FileEvent{
				Path:   "src/main.go",
				Action: "edit",
			},
			want: true,
		},
		{
			name: "match path pattern",
			trigger: &schema.FileTrigger{
				Paths: []string{"**/*.go"},
			},
			event: &schema.FileEvent{
				Path:   "src/main.go",
				Action: "edit",
			},
			want: true,
		},
		{
			name: "path ignore",
			trigger: &schema.FileTrigger{
				Paths:       []string{"**/*.go"},
				PathsIgnore: []string{"**/test_*.go"},
			},
			event: &schema.FileEvent{
				Path:   "src/test_main.go",
				Action: "edit",
			},
			want: false,
		},
		{
			name: "no path match",
			trigger: &schema.FileTrigger{
				Paths: []string{"**/*.ts"},
			},
			event: &schema.FileEvent{
				Path:   "src/main.go",
				Action: "edit",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &schema.Workflow{
				On: schema.OnConfig{
					File: tt.trigger,
				},
			}
			matcher := NewMatcher(workflow)
			event := &schema.Event{
				File: tt.event,
			}
			if got := matcher.Match(event); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchPushTrigger(t *testing.T) {
	tests := []struct {
		name    string
		trigger *schema.PushTrigger
		event   *schema.PushEvent
		want    bool
	}{
		{
			name: "match branch",
			trigger: &schema.PushTrigger{
				Branches: []string{"main"},
			},
			event: &schema.PushEvent{
				Ref: "refs/heads/main",
			},
			want: true,
		},
		{
			name: "match branch pattern",
			trigger: &schema.PushTrigger{
				Branches: []string{"feature/**"},
			},
			event: &schema.PushEvent{
				Ref: "refs/heads/feature/new-thing",
			},
			want: true,
		},
		{
			name: "branch ignore",
			trigger: &schema.PushTrigger{
				BranchesIgnore: []string{"main"},
			},
			event: &schema.PushEvent{
				Ref: "refs/heads/main",
			},
			want: false,
		},
		{
			name: "match tag",
			trigger: &schema.PushTrigger{
				Tags: []string{"v*"},
			},
			event: &schema.PushEvent{
				Ref: "refs/tags/v1.0.0",
			},
			want: true,
		},
		{
			name: "tag ignore",
			trigger: &schema.PushTrigger{
				TagsIgnore: []string{"v*-beta"},
			},
			event: &schema.PushEvent{
				Ref: "refs/tags/v1.0.0-beta",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &schema.Workflow{
				On: schema.OnConfig{
					Push: tt.trigger,
				},
			}
			matcher := NewMatcher(workflow)
			event := &schema.Event{
				Push: tt.event,
			}
			if got := matcher.Match(event); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"*.js", "test.js", true},
		{"*.js", "test.ts", false},
		{"**/*.js", "src/test.js", true},
		{"**/*.js", "deep/nested/test.js", true},
		{"src/**/*.go", "src/pkg/main.go", true},
		{"src/**/*.go", "other/main.go", false},
		{"src/**/test_*.go", "src/pkg/test_main.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.path, func(t *testing.T) {
			if got := matchGlob(tt.pattern, tt.path); got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestExtractBranch(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"refs/heads/main", "main"},
		{"refs/heads/feature/test", "feature/test"},
		{"refs/tags/v1.0.0", ""},
		{"main", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			if got := extractBranch(tt.ref); got != tt.want {
				t.Errorf("extractBranch(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}

func TestExtractTag(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"refs/tags/v1.0.0", "v1.0.0"},
		{"refs/tags/release-1", "release-1"},
		{"refs/heads/main", ""},
		{"v1.0.0", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			if got := extractTag(tt.ref); got != tt.want {
				t.Errorf("extractTag(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}

func TestMatchCommitTrigger(t *testing.T) {
	tests := []struct {
		name    string
		trigger *schema.CommitTrigger
		event   *schema.CommitEvent
		want    bool
	}{
		{
			name: "match path pattern",
			trigger: &schema.CommitTrigger{
				Paths: []string{"src/**/*.go"},
			},
			event: &schema.CommitEvent{
				SHA:     "abc123",
				Message: "feat: add feature",
				Files: []schema.FileStatus{
					{Path: "src/main.go", Status: "modified"},
				},
			},
			want: true,
		},
		{
			name: "no match path pattern",
			trigger: &schema.CommitTrigger{
				Paths: []string{"src/**/*.ts"},
			},
			event: &schema.CommitEvent{
				SHA:     "abc123",
				Message: "feat: add feature",
				Files: []schema.FileStatus{
					{Path: "src/main.go", Status: "modified"},
				},
			},
			want: false,
		},
		{
			name: "path ignore",
			trigger: &schema.CommitTrigger{
				PathsIgnore: []string{"**/*_test.go"},
			},
			event: &schema.CommitEvent{
				SHA:     "abc123",
				Message: "test: add tests",
				Files: []schema.FileStatus{
					{Path: "src/main_test.go", Status: "added"},
				},
			},
			want: false,
		},
		{
			name: "multiple files - one matches",
			trigger: &schema.CommitTrigger{
				Paths: []string{"**/*.go"},
			},
			event: &schema.CommitEvent{
				SHA:     "abc123",
				Message: "refactor",
				Files: []schema.FileStatus{
					{Path: "src/main.go", Status: "modified"},
					{Path: "README.md", Status: "modified"},
				},
			},
			want: true,
		},
		{
			name: "empty trigger matches all",
			trigger: &schema.CommitTrigger{},
			event: &schema.CommitEvent{
				SHA:     "abc123",
				Message: "any commit",
				Files: []schema.FileStatus{
					{Path: "anything.txt", Status: "added"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &schema.Workflow{
				On: schema.OnConfig{
					Commit: tt.trigger,
				},
			}
			matcher := NewMatcher(workflow)
			event := &schema.Event{
				Commit: tt.event,
			}
			if got := matcher.Match(event); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchToolsArray(t *testing.T) {
	tests := []struct {
		name     string
		triggers []schema.ToolTrigger
		event    *schema.ToolEvent
		want     bool
	}{
		{
			name: "match first tool",
			triggers: []schema.ToolTrigger{
				{Name: "edit"},
				{Name: "create"},
			},
			event: &schema.ToolEvent{
				Name: "edit",
				Args: map[string]interface{}{},
			},
			want: true,
		},
		{
			name: "match second tool",
			triggers: []schema.ToolTrigger{
				{Name: "edit"},
				{Name: "create"},
			},
			event: &schema.ToolEvent{
				Name: "create",
				Args: map[string]interface{}{},
			},
			want: true,
		},
		{
			name: "no match any tool",
			triggers: []schema.ToolTrigger{
				{Name: "edit"},
				{Name: "create"},
			},
			event: &schema.ToolEvent{
				Name: "powershell",
				Args: map[string]interface{}{},
			},
			want: false,
		},
		{
			name: "match with args pattern",
			triggers: []schema.ToolTrigger{
				{Name: "edit", Args: map[string]string{"path": "src/**"}},
				{Name: "create", Args: map[string]string{"path": "tests/**"}},
			},
			event: &schema.ToolEvent{
				Name: "create",
				Args: map[string]interface{}{"path": "tests/new_test.go"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &schema.Workflow{
				On: schema.OnConfig{
					Tools: tt.triggers,
				},
			}
			matcher := NewMatcher(workflow)
			event := &schema.Event{
				Tool: tt.event,
			}
			if got := matcher.Match(event); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCommitTriggerFromShellCommand tests that commit triggers fire when a shell command
// contains git commit and the event has a commit field populated
func TestCommitTriggerFromShellCommand(t *testing.T) {
	// This simulates what happens when PowerShell detects "git commit" in a command
	// and populates both the tool event and commit event
	workflow := &schema.Workflow{
		On: schema.OnConfig{
			Commit: &schema.CommitTrigger{
				Paths: []string{"src/**/*.go"},
			},
		},
	}
	matcher := NewMatcher(workflow)

	// Event with both tool (powershell with git commit command) and commit data
	event := &schema.Event{
		Tool: &schema.ToolEvent{
			Name: "powershell",
			Args: map[string]interface{}{
				"command": `git commit -m "feat: add feature"`,
			},
		},
		Commit: &schema.CommitEvent{
			SHA:     "pending",
			Message: "feat: add feature",
			Files: []schema.FileStatus{
				{Path: "src/main.go", Status: "modified"},
			},
		},
	}

	if !matcher.Match(event) {
		t.Error("Expected commit trigger to match when git commit command detected with matching files")
	}

	// Event with commit but non-matching files
	eventNoMatch := &schema.Event{
		Commit: &schema.CommitEvent{
			SHA:     "pending",
			Message: "docs: update readme",
			Files: []schema.FileStatus{
				{Path: "README.md", Status: "modified"},
			},
		},
	}

	if matcher.Match(eventNoMatch) {
		t.Error("Expected commit trigger to NOT match when files don't match path pattern")
	}
}

// TestPushTriggerFromShellCommand tests that push triggers fire when a shell command
// contains git push and the event has a push field populated
func TestPushTriggerFromShellCommand(t *testing.T) {
	workflow := &schema.Workflow{
		On: schema.OnConfig{
			Push: &schema.PushTrigger{
				Branches: []string{"main", "release/**"},
			},
		},
	}
	matcher := NewMatcher(workflow)

	// Event with push to main branch
	event := &schema.Event{
		Tool: &schema.ToolEvent{
			Name: "bash",
			Args: map[string]interface{}{
				"command": "git push origin main",
			},
		},
		Push: &schema.PushEvent{
			Ref:    "refs/heads/main",
			Before: "0000000000000000000000000000000000000000",
			After:  "abc123",
		},
	}

	if !matcher.Match(event) {
		t.Error("Expected push trigger to match for main branch")
	}

	// Event with push to feature branch (not in trigger)
	eventNoMatch := &schema.Event{
		Push: &schema.PushEvent{
			Ref:    "refs/heads/feature/test",
			Before: "0000000000000000000000000000000000000000",
			After:  "abc123",
		},
	}

	if matcher.Match(eventNoMatch) {
		t.Error("Expected push trigger to NOT match for feature branch")
	}
}

// TestCombinedToolAndCommitEvent tests that both tool and commit triggers can be checked
func TestCombinedToolAndCommitEvent(t *testing.T) {
	// Workflow that triggers on tool:powershell
	workflow := &schema.Workflow{
		On: schema.OnConfig{
			Tool: &schema.ToolTrigger{
				Name: "powershell",
			},
		},
	}
	matcher := NewMatcher(workflow)

	// Event has both tool and commit - tool trigger should match
	event := &schema.Event{
		Tool: &schema.ToolEvent{
			Name: "powershell",
			Args: map[string]interface{}{
				"command": `git commit -m "test"`,
			},
		},
		Commit: &schema.CommitEvent{
			Message: "test",
			Files:   []schema.FileStatus{{Path: "test.txt", Status: "added"}},
		},
	}

	if !matcher.Match(event) {
		t.Error("Expected tool trigger to match even when commit data is also present")
	}
}
