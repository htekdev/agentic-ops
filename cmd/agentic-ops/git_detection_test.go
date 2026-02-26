package main

import (
	"testing"
)

// TestDetectGitCommitCommand tests that git commit commands are detected
func TestDetectGitCommitCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		isCommit bool
	}{
		{"simple git commit", "git commit -m 'test'", true},
		{"git commit with message", `git commit -m "feat: add feature"`, true},
		{"git ci alias", "git ci -m 'test'", true},
		{"git commit amend", "git commit --amend", true},
		{"git commit all", "git commit -a -m 'test'", true},
		{"not a commit", "git status", false},
		{"not a commit - push", "git push origin main", false},
		{"not a commit - pull", "git pull", false},
		{"echo with git commit in string", `echo "git commit"`, false}, // This is tricky - word boundary should help
		{"git commit in middle", "cd repo && git commit -m 'test'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGitCommitCommand(tt.command)
			if got != tt.isCommit {
				t.Errorf("isGitCommitCommand(%q) = %v, want %v", tt.command, got, tt.isCommit)
			}
		})
	}
}

// TestDetectGitPushCommand tests that git push commands are detected
func TestDetectGitPushCommand(t *testing.T) {
	tests := []struct {
		name   string
		command string
		isPush bool
	}{
		{"simple git push", "git push", true},
		{"git push origin main", "git push origin main", true},
		{"git push with tags", "git push --tags", true},
		{"git push origin tag", "git push origin v1.0.0", true},
		{"git push force", "git push --force", true},
		{"git push upstream", "git push -u origin feature", true},
		{"not a push", "git status", false},
		{"not a push - commit", "git commit -m 'test'", false},
		{"not a push - pull", "git pull", false},
		{"git push in middle", "cd repo && git push origin main", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGitPushCommand(tt.command)
			if got != tt.isPush {
				t.Errorf("isGitPushCommand(%q) = %v, want %v", tt.command, got, tt.isPush)
			}
		})
	}
}

// TestExtractCommitMessage tests extracting commit messages from commands
func TestExtractCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		command string
		message string
	}{
		{"double quotes", `git commit -m "feat: add feature"`, "feat: add feature"},
		{"single quotes", `git commit -m 'fix: bug fix'`, "fix: bug fix"},
		{"no message flag", "git commit --amend", ""},
		{"message with spaces", `git commit -m "this is a long message"`, "this is a long message"},
		{"message at end", `git add . && git commit -m "done"`, "done"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCommitMessage(tt.command)
			if got != tt.message {
				t.Errorf("extractCommitMessage(%q) = %q, want %q", tt.command, got, tt.message)
			}
		})
	}
}

// TestExtractPushRef tests extracting push refs from commands
func TestExtractPushRef(t *testing.T) {
	tests := []struct {
		name    string
		command string
		branch  string // current branch
		wantRef string
	}{
		{"simple push", "git push", "main", "refs/heads/main"},
		{"push to origin", "git push origin main", "main", "refs/heads/main"},
		{"push feature branch", "git push origin feature/test", "feature/test", "refs/heads/feature/test"},
		{"push tag", "git push origin v1.0.0", "main", "refs/tags/v1.0.0"},
		{"push tags flag", "git push --tags", "main", "refs/heads/main"}, // Still branch, tags are separate
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPushRef(tt.command, tt.branch)
			if got != tt.wantRef {
				t.Errorf("extractPushRef(%q, %q) = %q, want %q", tt.command, tt.branch, got, tt.wantRef)
			}
		})
	}
}

// TestParseEventWithGitCommit tests that git commit events are properly parsed
func TestParseEventWithGitCommit(t *testing.T) {
	data := map[string]interface{}{
		"hook": map[string]interface{}{
			"type": "preToolUse",
			"tool": map[string]interface{}{
				"name": "powershell",
				"args": map[string]interface{}{
					"command": `git commit -m "feat: new feature"`,
				},
			},
		},
		"tool": map[string]interface{}{
			"name": "powershell",
			"args": map[string]interface{}{
				"command": `git commit -m "feat: new feature"`,
			},
		},
		"commit": map[string]interface{}{
			"sha":     "pending",
			"message": "feat: new feature",
			"author":  "test@example.com",
			"branch":  "main",
			"files": []interface{}{
				map[string]interface{}{
					"path":   "src/feature.go",
					"status": "added",
				},
			},
		},
	}

	event := parseEventData(data)

	if event.Commit == nil {
		t.Fatal("Expected Commit to be set")
	}
	if event.Commit.Message != "feat: new feature" {
		t.Errorf("Expected Commit.Message = 'feat: new feature', got '%s'", event.Commit.Message)
	}
	if len(event.Commit.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(event.Commit.Files))
	}
	if event.Commit.Files[0].Path != "src/feature.go" {
		t.Errorf("Expected file path = 'src/feature.go', got '%s'", event.Commit.Files[0].Path)
	}
}

// TestParseEventWithGitPush tests that git push events are properly parsed
func TestParseEventWithGitPush(t *testing.T) {
	data := map[string]interface{}{
		"hook": map[string]interface{}{
			"type": "preToolUse",
			"tool": map[string]interface{}{
				"name": "bash",
				"args": map[string]interface{}{
					"command": "git push origin main",
				},
			},
		},
		"push": map[string]interface{}{
			"ref":    "refs/heads/main",
			"before": "0000000000000000000000000000000000000000",
			"after":  "abc123def456",
		},
	}

	event := parseEventData(data)

	if event.Push == nil {
		t.Fatal("Expected Push to be set")
	}
	if event.Push.Ref != "refs/heads/main" {
		t.Errorf("Expected Push.Ref = 'refs/heads/main', got '%s'", event.Push.Ref)
	}
	if event.Push.After != "abc123def456" {
		t.Errorf("Expected Push.After = 'abc123def456', got '%s'", event.Push.After)
	}
}
