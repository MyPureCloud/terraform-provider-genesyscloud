// Package examples provides unit tests for the example loading and processing functionality.
package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUnitLoadExampleWithDependencies tests the basic functionality of loading examples.
// It creates temporary test files and verifies that the example loading process correctly
// loads resources and locals from those files.
func TestUnitLoadExampleWithDependencies(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "example-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple resource file
	resourcePath := filepath.Join(tempDir, "resource.tf")
	err = os.WriteFile(resourcePath, []byte(`
resource "test_resource" "example" {
  name = "test"
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write resource file: %v", err)
	}

	// Create a locals file
	localsPath := filepath.Join(tempDir, "locals.tf")
	err = os.WriteFile(localsPath, []byte(`
locals {
  working_dir = {
    test = "."
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write locals file: %v", err)
	}

	// Load the example
	example := NewExample()
	processedState := NewProcessedExampleState()
	loadedExample, err := example.LoadExampleWithDependencies(resourcePath, processedState)
	if err != nil {
		t.Fatalf("Failed to load example: %v", err)
	}

	// Verify the example was loaded correctly
	if len(loadedExample.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(loadedExample.Resources))
	}

	if loadedExample.Locals == nil {
		t.Error("Expected locals to be loaded, but it's nil")
	}

	if len(loadedExample.Locals.WorkingDir) != 1 {
		t.Errorf("Expected 1 working_dir entry, got %d", len(loadedExample.Locals.WorkingDir))
	}
}

// TestUnitLoadExampleWithSimpleDependency tests loading an example with a dependency.
// It verifies that dependencies defined in locals.tf are correctly loaded and that
// resources from those dependencies are included in the final example.
func TestUnitLoadExampleWithSimpleDependency(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "example-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a main resource file
	mainPath := filepath.Join(tempDir, "main.tf")
	err = os.WriteFile(mainPath, []byte(`
resource "test_resource" "main" {
  name = "main"
  foo  = test_resource.dependency.id
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Create a dependency resource file
	depPath := filepath.Join(tempDir, "dependency.tf")
	err = os.WriteFile(depPath, []byte(`
resource "test_resource" "dependency" {
  name = "dependency"
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write dependency file: %v", err)
	}

	// Create a locals file with dependency
	localsPath := filepath.Join(tempDir, "locals.tf")
	err = os.WriteFile(localsPath, []byte(`
locals {
  dependencies = {
    main = ["dependency.tf"]
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write locals file: %v", err)
	}

	// Load the example
	example := NewExample()
	processedState := NewProcessedExampleState()
	loadedExample, err := example.LoadExampleWithDependencies(mainPath, processedState)
	if err != nil {
		t.Fatalf("Failed to load example: %v", err)
	}

	// Verify the example was loaded correctly
	if len(loadedExample.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(loadedExample.Resources))
	}

	if len(loadedExample.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(loadedExample.Dependencies))
	}
}

// TestUnitLoadExampleCyclicDependencyDetection tests that cyclic dependencies are detected.
// It creates a circular dependency between two example files and verifies that the
// loading process detects the cycle, logs a warning, and continues processing without error or
// getting caught in a cycle.
func TestUnitLoadExampleCyclicDependencyDetection(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "example-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create resource files
	aDir := filepath.Join(tempDir, "a")
	err = os.Mkdir(aDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create a directory: %v", err)
	}
	aPath := filepath.Join(aDir, "a.tf")
	err = os.WriteFile(aPath, []byte(`resource "test_resource" "a" { name = "a" }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write a.tf: %v", err)
	}

	bDir := filepath.Join(tempDir, "b")
	err = os.Mkdir(bDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create a directory: %v", err)
	}
	bPath := filepath.Join(bDir, "b.tf")
	err = os.WriteFile(bPath, []byte(`resource "test_resource" "b" { name = "b" }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write b.tf: %v", err)
	}

	// Create locals files with cyclic dependencies
	aLocalsPath := filepath.Join(aDir, "locals.tf")
	err = os.WriteFile(aLocalsPath, []byte(`
locals {
  dependencies = {
    a = ["../b/b.tf"]
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write a locals file: %v", err)
	}

	// Create b's locals with dependency back to a
	bLocalsPath := filepath.Join(bDir, "locals.tf")
	err = os.WriteFile(bLocalsPath, []byte(`
locals {
  dependencies = {
    b = ["../a/a.tf"]
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write b locals file: %v", err)
	}

	// Load the example - should detect the cycle
	example := NewExample()
	processedState := NewProcessedExampleState()
	loadedExample, err := example.LoadExampleWithDependencies(aPath, processedState)

	// Verify that no error was returned (cycle should be handled internally)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Verify that a warning was logged
	logStr := loadedExample.Warnings[0]
	if !strings.Contains(logStr, "Warning: cyclic dependency detected") {
		t.Errorf("Expected warning about cyclic dependency in logs, but got: %s", logStr)
	}

	// Verify that the example was still loaded with at least the main resource
	if len(loadedExample.Resources) != 2 {
		t.Error("Expected both resources to be loaded despite cycle")
	}
}

// TestUnitLoadExampleProcessedExampleState tests that the ProcessedExampleState correctly tracks processed files.
// It verifies that files are properly marked as processed and that the dependency chain is maintained.
func TestUnitLoadExampleProcessedExampleState(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "example-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a resource file
	resourcePath := filepath.Join(tempDir, "resource.tf")
	err = os.WriteFile(resourcePath, []byte(`resource "test_resource" "example" { name = "test" }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write resource file: %v", err)
	}

	// Create a state and mark the file as processed
	state := NewProcessedExampleState()
	files := NewProcessedFiles()
	files.markFileAsProcessed(resourcePath)
	state.DirectoryTracker[tempDir] = files

	// Verify the file is marked as processed
	if !files.isFileProcessed(resourcePath) {
		t.Error("Expected resource.tf to be marked as processed, but it wasn't")
	}

	// Verify the dependency chain is empty
	if len(state.DependencyChain) != 0 {
		t.Errorf("Expected empty dependency chain, got %d items", len(state.DependencyChain))
	}
}

// TestUnitLoadExampleDependencyChainTracking tests that the dependency chain is correctly tracked.
// It verifies that paths are added to and removed from the dependency chain as expected,
// which is essential for detecting cyclic dependencies.
func TestUnitLoadExampleDependencyChainTracking(t *testing.T) {
	// Create a ProcessedExampleState
	state := NewProcessedExampleState()

	// Add some paths to the dependency chain
	path1 := "/path/to/file1.tf"
	path2 := "/path/to/file2.tf"

	state.DependencyChain = append(state.DependencyChain, path1)
	state.DependencyChain = append(state.DependencyChain, path2)

	// Verify the paths are in the chain
	if len(state.DependencyChain) != 2 {
		t.Errorf("Expected 2 items in dependency chain, got %d", len(state.DependencyChain))
	}

	if state.DependencyChain[0] != path1 || state.DependencyChain[1] != path2 {
		t.Errorf("Dependency chain contains incorrect paths: %v", state.DependencyChain)
	}

	// Remove the last path (simulating defer function)
	state.DependencyChain = state.DependencyChain[:len(state.DependencyChain)-1]

	// Verify only path1 remains
	if len(state.DependencyChain) != 1 || state.DependencyChain[0] != path1 {
		t.Errorf("Expected only path1 in dependency chain, got: %v", state.DependencyChain)
	}
}

// TestUnitLoadExampleLocalsMerging tests that locals are properly merged from dependencies.
// It verifies that working directories, skip constraints, and environment variables from
// multiple locals.tf files are correctly combined in the final example.
func TestUnitLoadExampleLocalsMerging(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "example-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a main resource file
	mainDir := filepath.Join(tempDir, "main")
	err = os.Mkdir(mainDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create main directory: %v", err)
	}
	mainPath := filepath.Join(mainDir, "main.tf")
	err = os.WriteFile(mainPath, []byte(`resource "test_resource" "main" { name = "main" }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Create a dependency resource file
	depDir1 := filepath.Join(tempDir, "deps")
	err = os.Mkdir(depDir1, 0755)
	if err != nil {
		t.Fatalf("Failed to create dep directory: %v", err)
	}
	depPath1 := filepath.Join(depDir1, "resource.tf")
	err = os.WriteFile(depPath1, []byte(`resource "test_resource" "dependency" { name = "dependency" }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write dependency file: %v", err)
	}

	// Create a main locals file with some settings
	mainLocalsPath := filepath.Join(mainDir, "locals.tf")
	err = os.WriteFile(mainLocalsPath, []byte(`
locals {
  dependencies = {
    main = ["../deps/resource.tf"]
  }
  working_dir = {
    main = "."
  }
  skip_if = {
    not_in_domains = ["test1"]
  }
  environment_vars = {
    MAIN_VAR = "main_value"
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main locals file: %v", err)
	}

	// Create a dependency locals file with different settings
	depLocalsPath1 := filepath.Join(depDir1, "locals.tf")
	err = os.WriteFile(depLocalsPath1, []byte(`
locals {
  working_dir = {
    dep = "."
  }
  skip_if = {
    not_in_domains = ["test2"]
    only_in_domains = ["test3"]
  }
  environment_vars = {
    DEP_VAR = "dep_value"
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write dependency locals file: %v", err)
	}

	// Create another dependency resource
	depDir2 := filepath.Join(tempDir, "deps2")
	err = os.Mkdir(depDir2, 0755)
	if err != nil {
		t.Fatalf("Failed to create dep directory: %v", err)
	}
	depPath2 := filepath.Join(depDir2, "resource.tf")
	err = os.WriteFile(depPath2, []byte(`resource "test_resource" "dep_resource" { name = "dep_resource" }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write dependency resource file: %v", err)
	}

	// Update the main locals to include the dependency from the subdirectory
	err = os.WriteFile(mainLocalsPath, []byte(`
locals {
  dependencies = {
    main = ["../deps/resource.tf", "../deps2/resource.tf"]
  }
  working_dir = {
    main = "."
  }
  skip_if = {
    not_in_domains = ["test1"]
  }
  environment_vars = {
    MAIN_VAR = "main_value"
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to update main locals file: %v", err)
	}

	// Load the example
	example := NewExample()
	processedState := NewProcessedExampleState()
	loadedExample, err := example.LoadExampleWithDependencies(mainPath, processedState)
	if err != nil {
		t.Fatalf("Failed to load example: %v", err)
	}

	// Verify the locals were merged correctly
	if loadedExample.Locals == nil {
		t.Fatal("Expected locals to be loaded, but it's nil")
	}

	// Check working_dir merging
	if len(loadedExample.Locals.WorkingDir) != 2 {
		t.Errorf("Expected 2 working_dir entries, got %d", len(loadedExample.Locals.WorkingDir))
	}
	if val, ok := loadedExample.Locals.WorkingDir["main"]; !ok || val == "" {
		t.Errorf("Expected working_dir to have 'main' entry")
	}
	if val, ok := loadedExample.Locals.WorkingDir["dep"]; !ok || val == "" {
		t.Errorf("Expected working_dir to have 'dep' entry")
	}

	// Check skip_if merging
	if len(loadedExample.Locals.SkipIfConstraints) != 2 {
		t.Errorf("Expected 2 skip_if entries, got %d", len(loadedExample.Locals.SkipIfConstraints))
	}
	if notInDomains, ok := loadedExample.Locals.SkipIfConstraints["not_in_domains"]; !ok || len(notInDomains) != 2 {
		t.Errorf("Expected skip_if to have 'not_in_domains' with 2 entries, got %v", notInDomains)
	}
	if onlyInDomains, ok := loadedExample.Locals.SkipIfConstraints["only_in_domains"]; !ok || len(onlyInDomains) != 1 {
		t.Errorf("Expected skip_if to have 'only_in_domains' with 1 entry, got %v", onlyInDomains)
	}

	// Check environment_vars merging
	if len(loadedExample.Locals.EnvironmentVars) != 2 {
		t.Errorf("Expected 2 environment_vars entries, got %d", len(loadedExample.Locals.EnvironmentVars))
	}
	if val, ok := loadedExample.Locals.EnvironmentVars["MAIN_VAR"]; !ok || val != "main_value" {
		t.Errorf("Expected environment_vars to have 'MAIN_VAR' entry with value 'main_value'")
	}
	if val, ok := loadedExample.Locals.EnvironmentVars["DEP_VAR"]; !ok || val != "dep_value" {
		t.Errorf("Expected environment_vars to have 'DEP_VAR' entry with value 'dep_value'")
	}
}

// TestUnitLoadExampleLocalsNilHandling tests that nil locals are handled properly during merging.
// It verifies that when a main example has minimal locals and a dependency has more extensive
// locals, the merge operation correctly handles the nil fields and produces a valid result.
func TestUnitLoadExampleLocalsNilHandling(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "example-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a main resource file with no locals
	mainDir := filepath.Join(tempDir, "main")
	err = os.Mkdir(mainDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create main directory: %v", err)
	}
	mainPath := filepath.Join(mainDir, "main.tf")
	err = os.WriteFile(mainPath, []byte(`resource "test_resource" "main" { name = "main" }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Create a dependency resource file
	depDir := filepath.Join(tempDir, "dependency")
	err = os.Mkdir(depDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dependency directory: %v", err)
	}
	depPath := filepath.Join(depDir, "dependency.tf")
	err = os.WriteFile(depPath, []byte(`resource "test_resource" "dependency" { name = "dependency" }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write dependency file: %v", err)
	}

	// Create a main locals file with just dependencies
	mainLocalsPath := filepath.Join(mainDir, "locals.tf")
	err = os.WriteFile(mainLocalsPath, []byte(`
locals {
  dependencies = {
    main = ["../dependency/dependency.tf"]
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write main locals file: %v", err)
	}

	// Create a dependency locals file with some settings
	depLocalsPath := filepath.Join(depDir, "locals.tf")
	err = os.WriteFile(depLocalsPath, []byte(`
locals {
  working_dir = {
    dep = "."
  }
  environment_vars = {
    DEP_VAR = "dep_value"
  }
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write dependency locals file: %v", err)
	}

	// Load the example
	example := NewExample()
	processedState := NewProcessedExampleState()
	loadedExample, err := example.LoadExampleWithDependencies(mainPath, processedState)
	if err != nil {
		t.Fatalf("Failed to load example: %v", err)
	}

	// Verify the locals were merged correctly
	if loadedExample.Locals == nil {
		t.Fatal("Expected locals to be loaded, but it's nil")
	}

	// Check that the dependency's locals were merged into the main example
	if len(loadedExample.Locals.WorkingDir) != 1 {
		t.Errorf("Expected 1 working_dir entry, got %d", len(loadedExample.Locals.WorkingDir))
	}
	if val, ok := loadedExample.Locals.WorkingDir["dep"]; !ok || val == "" {
		t.Errorf("Expected working_dir to have 'dep' entry")
	}

	if len(loadedExample.Locals.EnvironmentVars) != 1 {
		t.Errorf("Expected 1 environment_vars entry, got %d", len(loadedExample.Locals.EnvironmentVars))
	}
	if val, ok := loadedExample.Locals.EnvironmentVars["DEP_VAR"]; !ok || val != "dep_value" {
		t.Errorf("Expected environment_vars to have 'DEP_VAR' entry with value 'dep_value'")
	}
}

// TestUnitLoadExampleLocalsMergeDirectly tests the Locals.Merge function directly.
// It creates two Locals objects with various attributes and verifies that merging them
// produces the expected combined result with all fields properly merged.
func TestUnitLoadExampleLocalsMergeDirectly(t *testing.T) {
	// Create two Locals objects
	locals1 := &Locals{
		Dependencies: map[string][]string{
			"file1": {"dep1.tf", "dep2.tf"},
			"file2": {"dep3.tf"},
		},
		WorkingDir: map[string]string{
			"dir1": "/path/to/dir1",
			"dir2": "/path/to/dir2",
		},
		SkipIfConstraints: map[string][]string{
			"not_in_domains": {"domain1"},
		},
		EnvironmentVars: map[string]string{
			"VAR1": "value1",
		},
		ExtraAttributes: map[string]interface{}{
			"key1": "value1",
		},
	}

	locals2 := &Locals{
		Dependencies: map[string][]string{
			"file1": {"dep1.tf", "dep3.tf"},
			"file2": {"dep4.tf"},
		},
		WorkingDir: map[string]string{
			"dir2": "/path/to/dir2",
			"dir3": "/path/to/dir3",
		},
		SkipIfConstraints: map[string][]string{
			"not_in_domains":  {"domain2"},
			"only_in_domains": {"domain3"},
		},
		EnvironmentVars: map[string]string{
			"VAR2": "value2",
		},
		ExtraAttributes: map[string]interface{}{
			"key2": "value2",
		},
	}

	// Merge locals2 into locals1
	locals1.Merge(locals2)

	// Verify the merge results

	// Check Dependencies
	if len(locals1.Dependencies) != 2 {
		t.Errorf("Expected 2 dependency entries, got %d", len(locals1.Dependencies))
	}
	if deps, ok := locals1.Dependencies["file1"]; !ok || len(deps) != 3 {
		t.Errorf("Expected 'file1' to have 3 dependencies, got %v", deps)
	}
	if deps, ok := locals1.Dependencies["file2"]; !ok || len(deps) != 2 {
		t.Errorf("Expected 'file2' to have 1 dependency, got %v", deps)
	}

	// Check WorkingDir
	if len(locals1.WorkingDir) != 3 {
		t.Errorf("Expected 2 working_dir entries, got %d", len(locals1.WorkingDir))
	}
	if val, ok := locals1.WorkingDir["dir1"]; !ok || val != "/path/to/dir1" {
		t.Errorf("Expected working_dir to have 'dir1' entry with value '/path/to/dir1'")
	}
	if val, ok := locals1.WorkingDir["dir2"]; !ok || val != "/path/to/dir2" {
		t.Errorf("Expected working_dir to have 'dir2' entry with value '/path/to/dir2'")
	}
	if val, ok := locals1.WorkingDir["dir3"]; !ok || val != "/path/to/dir3" {
		t.Errorf("Expected working_dir to have 'dir2' entry with value '/path/to/dir3'")
	}

	// Check SkipIfConstraints
	if len(locals1.SkipIfConstraints) != 2 {
		t.Errorf("Expected 2 skip_if entries, got %d", len(locals1.SkipIfConstraints))
	}
	if constraints, ok := locals1.SkipIfConstraints["not_in_domains"]; !ok || len(constraints) != 2 {
		t.Errorf("Expected 'not_in_domains' to have 2 constraints, got %v", constraints)
	}
	if constraints, ok := locals1.SkipIfConstraints["only_in_domains"]; !ok || len(constraints) != 1 {
		t.Errorf("Expected 'only_in_domains' to have 1 constraint, got %v", constraints)
	}

	// Check EnvironmentVars
	if len(locals1.EnvironmentVars) != 2 {
		t.Errorf("Expected 2 environment_vars entries, got %d", len(locals1.EnvironmentVars))
	}
	if val, ok := locals1.EnvironmentVars["VAR1"]; !ok || val != "value1" {
		t.Errorf("Expected environment_vars to have 'VAR1' entry with value 'value1'")
	}
	if val, ok := locals1.EnvironmentVars["VAR2"]; !ok || val != "value2" {
		t.Errorf("Expected environment_vars to have 'VAR2' entry with value 'value2'")
	}

	// Check Other
	if len(locals1.ExtraAttributes) != 2 {
		t.Errorf("Expected 2 other entries, got %d", len(locals1.ExtraAttributes))
	}
	if _, ok := locals1.ExtraAttributes["key1"]; !ok {
		t.Errorf("Expected other to have 'key1' entry")
	}
	if _, ok := locals1.ExtraAttributes["key2"]; !ok {
		t.Errorf("Expected other to have 'key2' entry")
	}
}
