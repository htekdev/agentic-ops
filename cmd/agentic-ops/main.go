package main

import (
	"fmt"
	"os"

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
		if file != "" {
			fmt.Printf("Validating file: %s\n", file)
		} else {
			if dir == "" {
				var err error
				dir, err = os.Getwd()
				if err != nil {
					return err
				}
			}
			fmt.Printf("Validating workflows in: %s\n", dir)
		}
		// TODO: Implement validation
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

		if workflow != "" {
			fmt.Printf("Running workflow '%s' with event\n", workflow)
		} else {
			fmt.Printf("Running matching workflows for event\n")
		}

		if event == "-" {
			fmt.Println("Reading event from stdin...")
			// TODO: Read from stdin
		} else if event != "" {
			fmt.Printf("Event: %s\n", event)
		}

		// TODO: Implement workflow execution
		// Default: allow
		fmt.Println(`{"permissionDecision":"allow"}`)
		return nil
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
