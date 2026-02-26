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
