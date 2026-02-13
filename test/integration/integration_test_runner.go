// Package main runs all Python test scripts and validates their output.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ATSOTECK/rage/pkg/rage"
	"golang.org/x/term"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// colorEnabled tracks whether to use colored output
var colorEnabled bool

func init() {
	// Enable colors if stdout is a terminal
	colorEnabled = term.IsTerminal(int(os.Stdout.Fd()))
}

// color wraps text in ANSI color codes if colors are enabled
func color(c, text string) string {
	if !colorEnabled {
		return text
	}
	return c + text + colorReset
}

// ScriptResult holds the results of running a single test script
type ScriptResult struct {
	Script   string        `json:"script"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Failures string        `json:"failures,omitempty"`
}

// extractTestCounts extracts pass/fail counts from the test_framework module
func extractTestCounts(state *rage.State) (passed, failed int, failures string) {
	if v := state.GetModuleAttr("test_framework", "__test_passed__"); v != nil {
		if i, ok := rage.AsInt(v); ok {
			passed = int(i)
		}
	}
	if v := state.GetModuleAttr("test_framework", "__test_failed__"); v != nil {
		if i, ok := rage.AsInt(v); ok {
			failed = int(i)
		}
	}
	if v := state.GetModuleAttr("test_framework", "__test_failures__"); v != nil {
		if s, ok := rage.AsString(v); ok {
			failures = s
		}
	}
	return
}

// runScript executes a Python test script and returns the results
func runScript(scriptPath string, scriptsDir string, timeout time.Duration) (passed, failed int, failures string, err error) {
	// Read test framework module
	frameworkPath := filepath.Join(scriptsDir, "test_framework.py")
	frameworkSource, err := os.ReadFile(frameworkPath)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to read test framework: %w", err)
	}

	// Read script
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to read script: %w", err)
	}

	// Create a new state with all modules enabled
	// Enable reflection builtins for the reflection test
	var state *rage.State
	if strings.Contains(filepath.Base(scriptPath), "reflection") {
		state = rage.NewStateWithModules(rage.WithAllModules(), rage.WithAllBuiltins())
	} else {
		state = rage.NewState()
	}
	defer state.Close()

	// Inject color setting into the test framework source
	colorValue := "False"
	if colorEnabled {
		colorValue = "True"
	}
	modifiedFramework := strings.Replace(string(frameworkSource), "__test_use_color__ = False", "__test_use_color__ = "+colorValue, 1)

	// Register the test framework as an importable module
	if err := state.RegisterPythonModule("test_framework", modifiedFramework); err != nil {
		return 0, 0, "", fmt.Errorf("failed to register test framework module: %w", err)
	}

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
	_, err = state.RunWithTimeout(string(source), timeout)
	if err != nil {
		return 0, 0, "", fmt.Errorf("execution error: %w", err)
	}

	// Extract test counts from the test_framework module
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
		// Skip non-Python files, directories, and the test framework module itself
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".py") || entry.Name() == "test_framework.py" {
			continue
		}

		scriptPath := filepath.Join(scriptsDir, entry.Name())

		result := ScriptResult{
			Script: entry.Name(),
		}

		start := time.Now()

		// Run script
		passed, failed, failures, err := runScript(scriptPath, scriptsDir, 30*time.Second)
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
	fmt.Println("\n" + color(colorDim, strings.Repeat("═", 70)))
	fmt.Println(color(colorBold, "TEST RESULTS"))
	fmt.Println(color(colorDim, strings.Repeat("═", 70)))

	// Separate passing and failing tests for better organization
	var failingTests []ScriptResult
	var passingTests []ScriptResult

	for _, r := range results {
		if r.Error != "" || r.Failed > 0 {
			failingTests = append(failingTests, r)
		} else {
			passingTests = append(passingTests, r)
		}
	}

	// Print passing tests first (collapsed)
	if len(passingTests) > 0 {
		fmt.Println()
		for _, r := range passingTests {
			status := color(colorGreen+colorBold, "✓ PASS")
			duration := color(colorDim, fmt.Sprintf("%.2fs", r.Duration.Seconds()))
			fmt.Printf("%s %s %s (%d tests)\n", status, r.Script, duration, r.Passed)
		}
	}

	// Print failing tests with details
	if len(failingTests) > 0 {
		fmt.Println()
		fmt.Println(color(colorRed+colorBold, "─── FAILURES ───"))

		for _, r := range failingTests {
			status := color(colorRed+colorBold, "✗ FAIL")
			duration := color(colorDim, fmt.Sprintf("%.2fs", r.Duration.Seconds()))

			// Show different info depending on whether script crashed or tests failed
			if r.Error != "" {
				// Script crashed - don't show test counts
				fmt.Printf("\n%s %s %s %s\n", status, r.Script, duration, color(colorRed, "(script error)"))
				fmt.Println()
				fmt.Println(color(colorRed, "  Error: ") + formatErrorMessage(r.Error))
			} else {
				// Tests ran but some failed
				fmt.Printf("\n%s %s %s (%d passed, %d failed)\n", status, r.Script, duration, r.Passed, r.Failed)
			}

			if r.Failures != "" {
				fmt.Println()
				fmt.Println(color(colorYellow, "  Failed assertions:"))
				for _, line := range strings.Split(strings.TrimSpace(r.Failures), "\n") {
					if line != "" {
						fmt.Printf("    %s %s\n", color(colorRed, "•"), line)
					}
				}
			}
		}
	}

	// Summary
	fmt.Println("\n" + color(colorDim, strings.Repeat("═", 70)))

	var summaryColor string
	var summaryIcon string
	if totalFailed > 0 {
		summaryColor = colorRed + colorBold
		summaryIcon = "✗"
	} else {
		summaryColor = colorGreen + colorBold
		summaryIcon = "✓"
	}

	passed := color(colorGreen, fmt.Sprintf("%d passed", totalPassed))
	failed := color(colorRed, fmt.Sprintf("%d failed", totalFailed))
	scripts := fmt.Sprintf("%d scripts", len(results))

	fmt.Printf("%s %s  %s, %s  %s\n",
		color(summaryColor, summaryIcon),
		color(summaryColor, "TOTAL:"),
		passed,
		failed,
		color(colorDim, "("+scripts+")"))
	fmt.Println(color(colorDim, strings.Repeat("═", 70)))
}

// formatErrorMessage improves error message readability
func formatErrorMessage(err string) string {
	// Extract the core error from wrapped errors
	if strings.Contains(err, "execution error:") {
		parts := strings.SplitN(err, "execution error:", 2)
		if len(parts) == 2 {
			coreError := strings.TrimSpace(parts[1])
			return formatPythonError(coreError)
		}
	}
	return err
}

// formatPythonError formats Python-style errors for better readability
func formatPythonError(err string) string {
	// Handle common error patterns
	if strings.Contains(err, "line ") {
		// Try to extract line number and highlight it
		return err
	}

	// Handle name errors
	if strings.Contains(err, "NameError") {
		return color(colorRed, "NameError: ") + strings.TrimPrefix(err, "NameError: ")
	}

	// Handle type errors
	if strings.Contains(err, "TypeError") {
		return color(colorRed, "TypeError: ") + strings.TrimPrefix(err, "TypeError: ")
	}

	// Handle syntax errors
	if strings.Contains(err, "SyntaxError") {
		return color(colorRed, "SyntaxError: ") + strings.TrimPrefix(err, "SyntaxError: ")
	}

	return err
}

func main() {
	// Get scripts directory
	scriptsDir := "scripts"
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		scriptsDir = "test/integration/scripts"
	}

	// Parse command line
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--help", "-h":
			fmt.Println("Usage: integration_test_runner [options]")
			fmt.Println("\nOptions:")
			fmt.Println("  -h, --help      Show this help message")
			fmt.Println("  --no-color      Disable colored output")
			fmt.Println("\nRun from test/integration directory or project root.")
			return
		case "--no-color":
			colorEnabled = false
		}
	}

	// Run tests
	fmt.Printf("%s Running Python integration tests...\n", color(colorBlue+colorBold, "▶"))
	results, totalPassed, totalFailed := runAllTests(scriptsDir)
	printResults(results, totalPassed, totalFailed)

	if totalFailed > 0 {
		os.Exit(1)
	}
}
