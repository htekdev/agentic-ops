package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "agent-workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test workflow files
	files := []string{"lint.yml", "security.yaml", "test.yml"}
	for _, f := range files {
		path := filepath.Join(workflowDir, f)
		if err := os.WriteFile(path, []byte("name: test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a non-workflow file
	if err := os.WriteFile(filepath.Join(workflowDir, "readme.md"), []byte("# readme"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test discovery
	workflows, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(workflows) != 3 {
		t.Errorf("Discover() found %d workflows, want 3", len(workflows))
	}

	// Verify names
	names := make(map[string]bool)
	for _, w := range workflows {
		names[w.Name] = true
	}

	for _, expected := range []string{"lint", "security", "test"} {
		if !names[expected] {
			t.Errorf("Discover() missing workflow %q", expected)
		}
	}
}

func TestDiscoverEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	workflows, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(workflows) != 0 {
		t.Errorf("Discover() found %d workflows, want 0", len(workflows))
	}
}

func TestDiscoverNoWorkflowDir(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create .github but not agent-workflows
	if err := os.MkdirAll(filepath.Join(tmpDir, ".github"), 0755); err != nil {
		t.Fatal(err)
	}

	workflows, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(workflows) != 0 {
		t.Errorf("Discover() found %d workflows, want 0", len(workflows))
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	workflowDir := filepath.Join(tmpDir, ".github", "agent-workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a workflow file
	if err := os.WriteFile(filepath.Join(workflowDir, "lint.yml"), []byte("name: lint"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		workflow   string
		wantExists bool
	}{
		{"exists yml", "lint", true},
		{"not exists", "security", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, exists := Exists(tmpDir, tt.workflow)
			if exists != tt.wantExists {
				t.Errorf("Exists(%q) = %v, want %v", tt.workflow, exists, tt.wantExists)
			}
			if exists && path == "" {
				t.Error("Exists() returned empty path for existing workflow")
			}
		})
	}
}

func TestWorkflowFile(t *testing.T) {
	wf := WorkflowFile{
		Path:    "/repo/.github/agent-workflows/lint.yml",
		Name:    "lint",
		RelPath: ".github/agent-workflows/lint.yml",
	}

	if wf.Name != "lint" {
		t.Errorf("WorkflowFile.Name = %q, want %q", wf.Name, "lint")
	}

	if wf.RelPath != ".github/agent-workflows/lint.yml" {
		t.Errorf("WorkflowFile.RelPath = %q, want %q", wf.RelPath, ".github/agent-workflows/lint.yml")
	}
}
