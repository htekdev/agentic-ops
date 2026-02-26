package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/htekdev/agentic-ops/internal/runner"
	"github.com/htekdev/agentic-ops/internal/schema"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "agentic-ops",
	Short: "Local workflow engine for agentic DevOps",
	Long: `agentic-ops is a CLI tool that executes local workflows triggered by
Copilot agent hooks, file changes, commits, and pushes.

Workflows are defined in .github/agent-workflows/*.yml using a GitHub Actions-like syntax.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("agentic-ops version %s\n", version)
	},
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover workflow files in the current directory",
	Long:  `Searches for .github/agent-workflows/*.yml files and lists them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}
		fmt.Printf("Discovering workflows in: %s\n", dir)
		// TODO: Implement discovery
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate workflow files",
	Long:  `Validates workflow YAML files against the schema.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		file, _ := cmd.Flags().GetString("file")

		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		// Validate specific file or directory
		var result *schema.ValidationResult
		if file != "" {
			fmt.Printf("Validating file: %s\n", file)
			result = schema.ValidateWorkflow(file)
		} else {
			fmt.Printf("Validating workflows in: %s\n", dir)
			result = schema.ValidateWorkflowsInDir(dir)
		}

		// Print results
		if result.Valid {
			if file != "" {
				fmt.Printf("✓ File is valid\n")
			} else {
				fmt.Printf("✓ All workflows are valid\n")
			}
			return nil
		}

		// Print errors
		for _, err := range result.Errors {
			fmt.Printf("✗ %s\n", err.File)
			fmt.Printf("  Error: %s\n", err.Message)
			for _, detail := range err.Details {
				fmt.Printf("    - %s\n", detail)
			}
		}

		// Exit with error code
		os.Exit(1)
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run workflows for an event",
	Long:  `Executes matching workflows based on the provided event payload.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		event, _ := cmd.Flags().GetString("event")
		workflow, _ := cmd.Flags().GetString("workflow")
		dir, _ := cmd.Flags().GetString("dir")

		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		// If workflow is specified, load and run it
		if workflow != "" {
			return runWorkflow(dir, workflow)
		}

		// If no workflow specified, discover and run matching workflows
		return runMatchingWorkflows(dir, event)
	},
}

var triggersCmd = &cobra.Command{
	Use:   "triggers",
	Short: "List available trigger types",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available trigger types:")
		fmt.Println("  hooks    - Agent hook events (preToolUse, postToolUse)")
		fmt.Println("  tool     - Tool-specific triggers with argument filtering")
		fmt.Println("  file     - File create/edit events")
		fmt.Println("  commit   - Git commit events")
		fmt.Println("  push     - Git push events")
	},
}

func init() {
	// Add commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(triggersCmd)

	// discover flags
	discoverCmd.Flags().StringP("dir", "d", "", "Directory to search (default: current directory)")

	// validate flags
	validateCmd.Flags().StringP("dir", "d", "", "Directory to search (default: current directory)")
	validateCmd.Flags().StringP("file", "f", "", "Specific file to validate")

	// run flags
	runCmd.Flags().StringP("event", "e", "", "Event JSON (use '-' for stdin)")
	runCmd.Flags().StringP("workflow", "w", "", "Specific workflow to run")
	runCmd.Flags().StringP("dir", "d", "", "Directory to search (default: current directory)")
}

// runWorkflow loads and executes a specific workflow
func runWorkflow(dir, workflowName string) error {
	// Try to find the workflow file
	path, found := findWorkflowFile(dir, workflowName)
	if !found {
		return fmt.Errorf("workflow '%s' not found", workflowName)
	}

	// Load the workflow
	wf, err := schema.LoadWorkflow(path)
	if err != nil {
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	// Execute the workflow
	ctx := context.Background()
	r := runner.NewRunner(wf, nil, dir)
	result := r.RunWithBlocking(ctx)

	// Output the result as JSON
	return outputWorkflowResult(result)
}

// runMatchingWorkflows discovers and runs all matching workflows
func runMatchingWorkflows(dir, event string) error {
	// For now, just return allow
	// TODO: Implement workflow discovery and matching
	result := schema.NewAllowResult()
	return outputWorkflowResult(result)
}

// findWorkflowFile finds a workflow file by name
func findWorkflowFile(dir, workflowName string) (string, bool) {
	for _, ext := range []string{".yml", ".yaml"} {
		path := fmt.Sprintf("%s/.github/agent-workflows/%s%s", dir, workflowName, ext)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

// outputWorkflowResult outputs the workflow result as JSON
func outputWorkflowResult(result *schema.WorkflowResult) error {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}
