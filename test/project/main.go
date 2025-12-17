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

	"github.com/ATSOTECK/oink/internal/compiler"
	"github.com/ATSOTECK/oink/internal/runtime"
	"github.com/ATSOTECK/oink/internal/stdlib"
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

// convertValue converts a Python runtime Value to a Go interface{} for JSON comparison
func convertValue(v runtime.Value) interface{} {
	switch val := v.(type) {
	case *runtime.PyNone:
		return nil
	case *runtime.PyBool:
		return val.Value
	case *runtime.PyInt:
		return val.Value
	case *runtime.PyFloat:
		return val.Value
	case *runtime.PyString:
		return val.Value
	case *runtime.PyList:
		result := make([]interface{}, len(val.Items))
		for i, item := range val.Items {
			result[i] = convertValue(item)
		}
		return result
	case *runtime.PyTuple:
		result := make([]interface{}, len(val.Items))
		for i, item := range val.Items {
			result[i] = convertValue(item)
		}
		return result
	case *runtime.PyDict:
		result := make(map[string]interface{})
		for k, v := range val.Items {
			keyStr := fmt.Sprintf("%v", convertValue(k))
			result[keyStr] = convertValue(v)
		}
		return result
	case *runtime.PySet:
		// Convert set to sorted list for comparison
		result := make([]interface{}, 0, len(val.Items))
		for k := range val.Items {
			result = append(result, convertValue(k))
		}
		return result
	default:
		return fmt.Sprintf("%v", v)
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

	// Reset and initialize modules
	runtime.ResetModules()
	stdlib.InitAllModules()

	// Compile
	code, errs := compiler.CompileSource(string(source), scriptPath)
	if len(errs) > 0 {
		var errMsgs []string
		for _, e := range errs {
			errMsgs = append(errMsgs, e.Error())
		}
		return nil, fmt.Errorf("compilation errors: %s", strings.Join(errMsgs, "; "))
	}

	// Execute with timeout
	vm := runtime.NewVM()
	vm.SetCheckInterval(100)

	_, err = vm.ExecuteWithTimeout(timeout, code)
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	// Extract results dictionary
	resultsVal := vm.GetGlobal("results")
	if resultsVal == nil || runtime.IsNone(resultsVal) {
		return nil, fmt.Errorf("no 'results' dictionary found in script")
	}

	resultsDict, ok := resultsVal.(*runtime.PyDict)
	if !ok {
		return nil, fmt.Errorf("'results' is not a dictionary")
	}

	// Convert to Go map
	results := make(map[string]interface{})
	for k, v := range resultsDict.Items {
		keyStr := ""
		if ks, ok := k.(*runtime.PyString); ok {
			keyStr = ks.Value
		} else {
			keyStr = fmt.Sprintf("%v", convertValue(k))
		}
		results[keyStr] = convertValue(v)
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
	scriptsDir := "../scripts"
	expectedDir := "expected_data"

	// Check if directories exist
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		// Try from project root
		scriptsDir = "test/scripts"
		expectedDir = "test/project/expected_data"
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
