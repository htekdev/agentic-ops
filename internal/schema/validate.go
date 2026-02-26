package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// ValidationError represents a validation error
type ValidationError struct {
	File    string
	Message string
	Details []string
}

// ValidationResult contains the results of validating workflows
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidateWorkflow validates a single workflow file against the schema
func ValidateWorkflow(filePath string) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			File:    filePath,
			Message: fmt.Sprintf("File not found: %v", err),
		})
		return result
	}

	// Read the workflow file
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			File:    filePath,
			Message: fmt.Sprintf("Failed to read file: %v", err),
		})
		return result
	}

	// Parse YAML to JSON
	var data interface{}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			File:    filePath,
			Message: fmt.Sprintf("Invalid YAML syntax: %v", err),
		})
		return result
	}

	// Convert to JSON for schema validation
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			File:    filePath,
			Message: fmt.Sprintf("Failed to convert to JSON: %v", err),
		})
		return result
	}

	// Load the schema
	schemaLoader, err := loadSchemaLoader()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			File:    filePath,
			Message: fmt.Sprintf("Failed to load schema: %v", err),
		})
		return result
	}

	// Create document loader from JSON bytes
	documentLoader := gojsonschema.NewBytesLoader(jsonBytes)

	// Validate
	validationResult, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			File:    filePath,
			Message: fmt.Sprintf("Validation error: %v", err),
		})
		return result
	}

	if !validationResult.Valid() {
		result.Valid = false
		details := []string{}
		for _, err := range validationResult.Errors() {
			details = append(details, err.String())
		}
		result.Errors = append(result.Errors, ValidationError{
			File:    filePath,
			Message: "Workflow validation failed",
			Details: details,
		})
	}

	return result
}

// ValidateWorkflowsInDir validates all workflow files in a directory
func ValidateWorkflowsInDir(dir string) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Find all YAML files in .github/agent-workflows
	workflowDir := filepath.Join(dir, ".github", "agent-workflows")

	// Check if directory exists
	if _, err := os.Stat(workflowDir); err != nil {
		// No workflows directory is not an error - just return valid
		return result
	}

	// Walk the directory
	err := filepath.Walk(workflowDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a YAML file
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".yml") &&
			!strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") {
			return nil
		}

		// Validate this file
		fileResult := ValidateWorkflow(path)
		if !fileResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fileResult.Errors...)
		}

		return nil
	})

	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			File:    dir,
			Message: fmt.Sprintf("Failed to scan directory: %v", err),
		})
	}

	return result
}

// loadSchemaLoader loads the workflow schema
func loadSchemaLoader() (gojsonschema.JSONLoader, error) {
	// Try to find the schema file relative to the executable or working directory
	possiblePaths := []string{
		"schema/workflow.schema.json",
		"./schema/workflow.schema.json",
		"../schema/workflow.schema.json",
		"../../schema/workflow.schema.json",
	}

	var schemaContent []byte
	var lastErr error

	for _, path := range possiblePaths {
		content, err := os.ReadFile(path)
		if err == nil {
			schemaContent = content
			break
		}
		lastErr = err
	}

	if schemaContent == nil {
		return nil, fmt.Errorf("failed to load schema file from any of the standard locations: %w", lastErr)
	}

	return gojsonschema.NewBytesLoader(schemaContent), nil
}
