package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/htekdev/agentic-ops/internal/expression"
	"github.com/htekdev/agentic-ops/internal/schema"
)

// Runner executes workflow steps
type Runner struct {
	workflow   *schema.Workflow
	event      *schema.Event
	exprCtx    *expression.Context
	workingDir string
	env        map[string]string
}

// StepResult contains the result of running a step
type StepResult struct {
	Name     string
	Success  bool
	Output   string
	Error    error
	Duration time.Duration
}

// NewRunner creates a new step runner
func NewRunner(workflow *schema.Workflow, event *schema.Event, workingDir string) *Runner {
	exprCtx := expression.NewContext()

	// Populate event context
	if event != nil {
		exprCtx.Event["cwd"] = event.Cwd
		exprCtx.Event["timestamp"] = event.Timestamp

		if event.Hook != nil {
			exprCtx.Event["hook"] = map[string]interface{}{
				"type": event.Hook.Type,
				"cwd":  event.Hook.Cwd,
			}
			if event.Hook.Tool != nil {
				exprCtx.Event["hook"].(map[string]interface{})["tool"] = map[string]interface{}{
					"name": event.Hook.Tool.Name,
					"args": event.Hook.Tool.Args,
				}
			}
		}

		if event.Tool != nil {
			exprCtx.Event["tool"] = map[string]interface{}{
				"name":      event.Tool.Name,
				"args":      event.Tool.Args,
				"hook_type": event.Tool.HookType,
			}
		}

		if event.File != nil {
			exprCtx.Event["file"] = map[string]interface{}{
				"path":    event.File.Path,
				"action":  event.File.Action,
				"content": event.File.Content,
			}
		}

		if event.Commit != nil {
			files := make([]map[string]string, len(event.Commit.Files))
			for i, f := range event.Commit.Files {
				files[i] = map[string]string{"path": f.Path, "status": f.Status}
			}
			exprCtx.Event["commit"] = map[string]interface{}{
				"sha":     event.Commit.SHA,
				"message": event.Commit.Message,
				"author":  event.Commit.Author,
				"files":   files,
			}
		}

		if event.Push != nil {
			exprCtx.Event["push"] = map[string]interface{}{
				"ref":    event.Push.Ref,
				"before": event.Push.Before,
				"after":  event.Push.After,
			}
		}
	}

	// Merge workflow env with event env
	env := make(map[string]string)
	for k, v := range workflow.Env {
		env[k] = v
	}
	exprCtx.Env = env

	return &Runner{
		workflow:   workflow,
		event:      event,
		exprCtx:    exprCtx,
		workingDir: workingDir,
		env:        env,
	}
}

// Run executes all steps in the workflow
func (r *Runner) Run(ctx context.Context) ([]StepResult, error) {
	var results []StepResult
	var prevStepFailed bool

	for i, step := range r.workflow.Steps {
		stepName := step.Name
		if stepName == "" {
			stepName = fmt.Sprintf("Step %d", i+1)
		}

		// Update step context for expressions
		r.exprCtx.Steps[stepName] = expression.StepContext{
			Outputs: make(map[string]string),
			Outcome: "pending",
		}

		// Check if condition
		if step.If != "" {
			// Evaluate if condition
			shouldRun, err := r.exprCtx.EvaluateBool(step.If)
			if err != nil {
				results = append(results, StepResult{
					Name:    stepName,
					Success: false,
					Error:   fmt.Errorf("failed to evaluate if condition: %w", err),
				})
				if !step.ContinueOnError {
					prevStepFailed = true
				}
				continue
			}
			if !shouldRun {
				results = append(results, StepResult{
					Name:    stepName,
					Success: true,
					Output:  "Skipped (condition not met)",
				})
				continue
			}
		}

		// If previous step failed and this doesn't have always(), skip
		if prevStepFailed && !strings.Contains(step.If, "always()") {
			results = append(results, StepResult{
				Name:    stepName,
				Success: false,
				Output:  "Skipped (previous step failed)",
			})
			continue
		}

		// Execute the step
		result := r.runStep(ctx, step, stepName)
		results = append(results, result)

		// Update step context
		outcome := "success"
		if !result.Success {
			outcome = "failure"
			if !step.ContinueOnError {
				prevStepFailed = true
			}
		}
		r.exprCtx.Steps[stepName] = expression.StepContext{
			Outputs: make(map[string]string),
			Outcome: outcome,
		}
	}

	return results, nil
}

// runStep executes a single step
func (r *Runner) runStep(ctx context.Context, step schema.Step, name string) StepResult {
	start := time.Now()

	// Handle timeout
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(step.Timeout)*time.Second)
		defer cancel()
	}

	// Check for uses: action
	if step.Uses != "" {
		return r.runAction(ctx, step, name, start)
	}

	// Execute run: command
	if step.Run != "" {
		return r.runCommand(ctx, step, name, start)
	}

	return StepResult{
		Name:     name,
		Success:  false,
		Error:    fmt.Errorf("step has neither 'run' nor 'uses'"),
		Duration: time.Since(start),
	}
}

// runCommand executes a shell command
func (r *Runner) runCommand(ctx context.Context, step schema.Step, name string, start time.Time) StepResult {
	// Evaluate expressions in command
	command, err := r.exprCtx.EvaluateString(step.Run)
	if err != nil {
		return StepResult{
			Name:     name,
			Success:  false,
			Error:    fmt.Errorf("failed to evaluate command: %w", err),
			Duration: time.Since(start),
		}
	}

	// Determine shell
	shell := step.Shell
	if shell == "" {
		shell = defaultShell()
	}

	// Build command
	var cmd *exec.Cmd
	switch shell {
	case "pwsh", "powershell":
		cmd = exec.CommandContext(ctx, "pwsh", "-NoProfile", "-NonInteractive", "-Command", command)
	case "bash":
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	case "sh":
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	case "cmd":
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	default:
		cmd = exec.CommandContext(ctx, shell, "-c", command)
	}

	// Set working directory
	workDir := r.workingDir
	if step.WorkingDirectory != "" {
		wd, err := r.exprCtx.EvaluateString(step.WorkingDirectory)
		if err == nil {
			workDir = wd
		}
	}
	cmd.Dir = workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range r.env {
		val, _ := r.exprCtx.EvaluateString(v)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, val))
	}
	for k, v := range step.Env {
		val, _ := r.exprCtx.EvaluateString(v)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, val))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err = cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return StepResult{
				Name:     name,
				Success:  false,
				Output:   output,
				Error:    fmt.Errorf("step timed out after %d seconds", step.Timeout),
				Duration: time.Since(start),
			}
		}
		return StepResult{
			Name:     name,
			Success:  false,
			Output:   output,
			Error:    err,
			Duration: time.Since(start),
		}
	}

	return StepResult{
		Name:     name,
		Success:  true,
		Output:   output,
		Duration: time.Since(start),
	}
}

// runAction executes a reusable action
func (r *Runner) runAction(ctx context.Context, step schema.Step, name string, start time.Time) StepResult {
	// TODO: Implement action execution
	// For now, return a placeholder
	return StepResult{
		Name:     name,
		Success:  false,
		Error:    fmt.Errorf("uses: actions not yet implemented: %s", step.Uses),
		Duration: time.Since(start),
	}
}

// defaultShell returns the default shell for the current OS
func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "pwsh"
	}
	return "bash"
}
