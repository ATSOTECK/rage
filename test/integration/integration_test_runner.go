// Package main runs all Python test scripts and validates their output.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ATSOTECK/rage/pkg/rage"
)

// Test framework injected into each Python script
const testFramework = `
__test_passed__ = 0
__test_failed__ = 0
__test_failures__ = ""

def expect(expected, actual):
    if expected != actual:
        raise Exception("Expected " + str(expected) + " but got " + str(actual))

def test(name, fn):
    global __test_passed__, __test_failed__, __test_failures__
    try:
        fn()
        __test_passed__ = __test_passed__ + 1
        print("[PASS] " + name)
    except Exception as e:
        __test_failed__ = __test_failed__ + 1
        __test_failures__ = __test_failures__ + name + ": " + str(e) + "\n"
        print("[FAIL] " + name + ": " + str(e))
`

// ScriptResult holds the results of running a single test script
type ScriptResult struct {
	Script   string        `json:"script"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Failures string        `json:"failures,omitempty"`
}

// convertValue converts a rage.Value to a Go any
func convertValue(v rage.Value) any {
	if v == nil || rage.IsNone(v) {
		return nil
	}

	switch {
	case rage.IsBool(v):
		b, _ := rage.AsBool(v)
		return b
	case rage.IsInt(v):
		i, _ := rage.AsInt(v)
		return i
	case rage.IsFloat(v):
		f, _ := rage.AsFloat(v)
		return f
	case rage.IsString(v):
		s, _ := rage.AsString(v)
		return s
	case rage.IsList(v):
		items, _ := rage.AsList(v)
		result := make([]any, len(items))
		for i, item := range items {
			result[i] = convertValue(item)
		}
		return result
	case rage.IsTuple(v):
		items, _ := rage.AsTuple(v)
		result := make([]any, len(items))
		for i, item := range items {
			result[i] = convertValue(item)
		}
		return result
	case rage.IsDict(v):
		items, _ := rage.AsDict(v)
		result := make(map[string]any)
		for k, val := range items {
			result[k] = convertValue(val)
		}
		return result
	default:
		return fmt.Sprintf("%v", v.GoValue())
	}
}

// extractTestCounts extracts pass/fail counts from Python globals
func extractTestCounts(state *rage.State) (passed, failed int, failures string) {
	if v := state.GetGlobal("__test_passed__"); v != nil {
		if i, ok := rage.AsInt(v); ok {
			passed = int(i)
		}
	}
	if v := state.GetGlobal("__test_failed__"); v != nil {
		if i, ok := rage.AsInt(v); ok {
			failed = int(i)
		}
	}
	if v := state.GetGlobal("__test_failures__"); v != nil {
		if s, ok := rage.AsString(v); ok {
			failures = s
		}
	}
	return
}

// runScript executes a Python test script and returns the results
func runScript(scriptPath string, timeout time.Duration) (passed, failed int, failures string, err error) {
	// Read script
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to read script: %w", err)
	}

	// Prepend test framework
	fullSource := testFramework + "\n" + string(source)

	// Create a new state with all modules enabled
	state := rage.NewState()
	defer state.Close()

	// Set up temp directory for file I/O tests
	if strings.Contains(filepath.Base(scriptPath), "file_io") ||
		strings.Contains(filepath.Base(scriptPath), "io_") ||
		strings.Contains(filepath.Base(scriptPath), "_os") {
		tmpDir, err := os.MkdirTemp("", "rage_io_test_")
		if err != nil {
			return 0, 0, "", fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)
		state.SetGlobal("__test_tmp_dir__", rage.String(tmpDir))
	}

	// Execute with timeout
	_, err = state.RunWithTimeout(fullSource, timeout)
	if err != nil {
		return 0, 0, "", fmt.Errorf("execution error: %w", err)
	}

	// Extract test counts
	passed, failed, failures = extractTestCounts(state)
	return passed, failed, failures, nil
}

// runAllTests runs all test scripts
func runAllTests(scriptsDir string) ([]ScriptResult, int, int) {
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		fmt.Printf("Error reading scripts directory: %v\n", err)
		return nil, 0, 0
	}

	var results []ScriptResult
	totalPassed := 0
	totalFailed := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".py") {
			continue
		}

		scriptPath := filepath.Join(scriptsDir, entry.Name())

		result := ScriptResult{
			Script: entry.Name(),
		}

		start := time.Now()

		// Run script
		passed, failed, failures, err := runScript(scriptPath, 30*time.Second)
		result.Duration = time.Since(start)

		if err != nil {
			result.Error = err.Error()
			totalFailed++
			results = append(results, result)
			continue
		}

		result.Passed = passed
		result.Failed = failed
		result.Failures = failures
		totalPassed += passed
		totalFailed += failed

		results = append(results, result)
	}

	return results, totalPassed, totalFailed
}

func printResults(results []ScriptResult, totalPassed, totalFailed int) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("TEST RESULTS")
	fmt.Println(strings.Repeat("=", 70))

	for _, r := range results {
		// Script header
		status := "PASS"
		if r.Error != "" || r.Failed > 0 {
			status = "FAIL"
		}

		fmt.Printf("\n[%s] %s (%d passed, %d failed) %.2fs\n", status, r.Script, r.Passed, r.Failed, r.Duration.Seconds())

		if r.Error != "" {
			fmt.Printf("  Script Error: %s\n", r.Error)
		}

		if r.Failures != "" {
			fmt.Printf("  Failures:\n")
			for _, line := range strings.Split(strings.TrimSpace(r.Failures), "\n") {
				if line != "" {
					fmt.Printf("    %s\n", line)
				}
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("TOTAL: %d passed, %d failed\n", totalPassed, totalFailed)
	fmt.Println(strings.Repeat("=", 70))
}

func main() {
	// Get scripts directory
	scriptsDir := "scripts"
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		scriptsDir = "test/integration/scripts"
	}

	// Parse command line
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: integration_test_runner [options]")
			fmt.Println("\nOptions:")
			fmt.Println("  -h, --help      Show this help message")
			fmt.Println("\nRun from test/integration directory or project root.")
			return
		}
	}

	// Run tests
	fmt.Println("Running Python integration tests...")
	results, totalPassed, totalFailed := runAllTests(scriptsDir)
	printResults(results, totalPassed, totalFailed)

	if totalFailed > 0 {
		os.Exit(1)
	}
}
