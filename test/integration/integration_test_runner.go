// Package main runs all Python test scripts and validates their output against expected data.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ATSOTECK/oink/pkg/rage"
)

// TestResult holds the result of running a single test script
type TestResult struct {
	Script     string           `json:"script"`
	Success    bool             `json:"success"`
	Duration   time.Duration    `json:"duration"`
	Error      string           `json:"error,omitempty"`
	Mismatches []MismatchDetail `json:"mismatches,omitempty"`
	Missing    []string         `json:"missing,omitempty"`
	Extra      []string         `json:"extra,omitempty"`
}

// MismatchDetail describes a single value mismatch
type MismatchDetail struct {
	Key      string `json:"key"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

// convertValue converts a rage.Value to a Go interface{} for JSON comparison
func convertValue(v rage.Value) interface{} {
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
		result := make([]interface{}, len(items))
		for i, item := range items {
			result[i] = convertValue(item)
		}
		return result
	case rage.IsTuple(v):
		items, _ := rage.AsTuple(v)
		result := make([]interface{}, len(items))
		for i, item := range items {
			result[i] = convertValue(item)
		}
		return result
	case rage.IsDict(v):
		items, _ := rage.AsDict(v)
		result := make(map[string]interface{})
		for k, val := range items {
			result[k] = convertValue(val)
		}
		return result
	default:
		return fmt.Sprintf("%v", v.GoValue())
	}
}

// normalizeForComparison normalizes values for comparison
func normalizeForComparison(v interface{}) interface{} {
	switch val := v.(type) {
	case float64:
		// JSON numbers are float64, but we might have int64 from runtime
		if val == float64(int64(val)) {
			return int64(val)
		}
		return val
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = normalizeForComparison(item)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = normalizeForComparison(v)
		}
		return result
	default:
		return v
	}
}

// valuesEqual compares two values for equality
func valuesEqual(expected, actual interface{}) bool {
	expected = normalizeForComparison(expected)
	actual = normalizeForComparison(actual)

	// Handle float comparison with tolerance
	if ef, ok := expected.(float64); ok {
		if af, ok := actual.(float64); ok {
			diff := ef - af
			if diff < 0 {
				diff = -diff
			}
			return diff < 0.0001
		}
		// Compare float64 expected with int64 actual
		if ai, ok := actual.(int64); ok {
			diff := ef - float64(ai)
			if diff < 0 {
				diff = -diff
			}
			return diff < 0.0001
		}
	}

	// Handle int comparison
	if ei, ok := expected.(int64); ok {
		if ai, ok := actual.(int64); ok {
			return ei == ai
		}
		if af, ok := actual.(float64); ok {
			return float64(ei) == af
		}
	}

	// Handle slice comparison
	if es, ok := expected.([]interface{}); ok {
		if as, ok := actual.([]interface{}); ok {
			if len(es) != len(as) {
				return false
			}
			for i := range es {
				if !valuesEqual(es[i], as[i]) {
					return false
				}
			}
			return true
		}
		return false
	}

	// Handle map comparison
	if em, ok := expected.(map[string]interface{}); ok {
		if am, ok := actual.(map[string]interface{}); ok {
			if len(em) != len(am) {
				return false
			}
			for k, ev := range em {
				av, ok := am[k]
				if !ok || !valuesEqual(ev, av) {
					return false
				}
			}
			return true
		}
		return false
	}

	return expected == actual
}

// runScript executes a Python script and returns the results dictionary
func runScript(scriptPath string, timeout time.Duration) (map[string]interface{}, error) {
	// Read script
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script: %w", err)
	}

	// Create a new state with all modules enabled
	state := rage.NewState()
	defer state.Close()

	// Execute with timeout
	_, err = state.RunWithTimeout(string(source), timeout)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	// Extract results dictionary
	resultsVal := state.GetGlobal("results")
	if resultsVal == nil || rage.IsNone(resultsVal) {
		return nil, fmt.Errorf("no 'results' dictionary found in script")
	}

	if !rage.IsDict(resultsVal) {
		return nil, fmt.Errorf("'results' is not a dictionary")
	}

	// Convert to Go map
	dictItems, _ := rage.AsDict(resultsVal)
	results := make(map[string]interface{})
	for k, v := range dictItems {
		results[k] = convertValue(v)
	}

	return results, nil
}

// loadExpectedData loads expected results from JSON file
func loadExpectedData(jsonPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var expected map[string]interface{}
	if err := json.Unmarshal(data, &expected); err != nil {
		return nil, err
	}

	return expected, nil
}

// compareResults compares actual results with expected data
func compareResults(expected, actual map[string]interface{}) ([]MismatchDetail, []string, []string) {
	var mismatches []MismatchDetail
	var missing []string
	var extra []string

	// Check for missing and mismatched keys
	for key, expectedVal := range expected {
		actualVal, exists := actual[key]
		if !exists {
			missing = append(missing, key)
		} else if !valuesEqual(expectedVal, actualVal) {
			mismatches = append(mismatches, MismatchDetail{
				Key:      key,
				Expected: fmt.Sprintf("%v", expectedVal),
				Actual:   fmt.Sprintf("%v", actualVal),
			})
		}
	}

	// Check for extra keys
	for key := range actual {
		if _, exists := expected[key]; !exists {
			extra = append(extra, key)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)
	sort.Slice(mismatches, func(i, j int) bool {
		return mismatches[i].Key < mismatches[j].Key
	})

	return mismatches, missing, extra
}

// generateExpectedData generates expected data JSON from script output
func generateExpectedData(scriptsDir, expectedDir string) error {
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".py") {
			continue
		}

		scriptPath := filepath.Join(scriptsDir, entry.Name())
		jsonName := strings.TrimSuffix(entry.Name(), ".py") + ".json"
		jsonPath := filepath.Join(expectedDir, jsonName)

		fmt.Printf("Generating expected data for %s...\n", entry.Name())

		results, err := runScript(scriptPath, 30*time.Second)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Printf("  ERROR marshaling JSON: %v\n", err)
			continue
		}

		if err := os.WriteFile(jsonPath, data, 0644); err != nil {
			fmt.Printf("  ERROR writing file: %v\n", err)
			continue
		}

		fmt.Printf("  Generated %s with %d keys\n", jsonName, len(results))
	}

	return nil
}

// runAllTests runs all test scripts and validates against expected data
func runAllTests(scriptsDir, expectedDir string) ([]TestResult, int, int) {
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		fmt.Printf("Error reading scripts directory: %v\n", err)
		return nil, 0, 0
	}

	var results []TestResult
	passed := 0
	failed := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".py") {
			continue
		}

		scriptPath := filepath.Join(scriptsDir, entry.Name())
		jsonName := strings.TrimSuffix(entry.Name(), ".py") + ".json"
		jsonPath := filepath.Join(expectedDir, jsonName)

		result := TestResult{
			Script: entry.Name(),
		}

		start := time.Now()

		// Run script
		actual, err := runScript(scriptPath, 30*time.Second)
		result.Duration = time.Since(start)

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			results = append(results, result)
			failed++
			continue
		}

		// Load expected data
		expected, err := loadExpectedData(jsonPath)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to load expected data: %v", err)
			results = append(results, result)
			failed++
			continue
		}

		// Compare results
		mismatches, missing, extra := compareResults(expected, actual)

		if len(mismatches) == 0 && len(missing) == 0 {
			result.Success = true
			passed++
		} else {
			result.Success = false
			result.Mismatches = mismatches
			result.Missing = missing
			result.Extra = extra
			failed++
		}

		results = append(results, result)
	}

	return results, passed, failed
}

func printResults(results []TestResult, passed, failed int) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	for _, r := range results {
		status := "PASS"
		if !r.Success {
			status = "FAIL"
		}

		fmt.Printf("\n[%s] %s (%.2fs)\n", status, r.Script, r.Duration.Seconds())

		if r.Error != "" {
			fmt.Printf("  Error: %s\n", r.Error)
		}

		if len(r.Mismatches) > 0 {
			fmt.Printf("  Mismatches (%d):\n", len(r.Mismatches))
			for _, m := range r.Mismatches {
				fmt.Printf("    - %s: expected=%s, actual=%s\n", m.Key, m.Expected, m.Actual)
			}
		}

		if len(r.Missing) > 0 {
			fmt.Printf("  Missing keys (%d): %s\n", len(r.Missing), strings.Join(r.Missing, ", "))
		}

		if len(r.Extra) > 0 && !r.Success {
			fmt.Printf("  Extra keys (%d): %s\n", len(r.Extra), strings.Join(r.Extra, ", "))
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("SUMMARY: %d passed, %d failed, %d total\n", passed, failed, passed+failed)
	fmt.Println(strings.Repeat("=", 60))
}

func main() {
	// Get directories relative to where we're running
	scriptsDir := "scripts"
	expectedDir := "expected_data"

	// Check if directories exist
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		// Try from project root
		scriptsDir = "test/integration/scripts"
		expectedDir = "test/integration/expected_data"
	}

	// Parse command line
	generate := false
	for _, arg := range os.Args[1:] {
		if arg == "--generate" || arg == "-g" {
			generate = true
		}
		if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: test-runner [options]")
			fmt.Println("\nOptions:")
			fmt.Println("  -g, --generate  Generate expected data files from script output")
			fmt.Println("  -h, --help      Show this help message")
			fmt.Println("\nRun from test/project directory or project root.")
			return
		}
	}

	if generate {
		fmt.Println("Generating expected data files...")
		if err := generateExpectedData(scriptsDir, expectedDir); err != nil {
			fmt.Printf("Error generating expected data: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\nExpected data generation complete!")
		return
	}

	// Run tests
	fmt.Println("Running Python language tests...")
	results, passed, failed := runAllTests(scriptsDir, expectedDir)
	printResults(results, passed, failed)

	if failed > 0 {
		os.Exit(1)
	}
}
