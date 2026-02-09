package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/ATSOTECK/rage/internal/stdlib"
)

func main() {
	// Initialize standard library modules
	stdlib.InitAllModules()

	if len(os.Args) < 2 {
		fmt.Println("Usage: rage <script.py>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Read the script file
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Compile the source
	code, errs := compiler.CompileSource(string(source), filename)
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "Compilation errors:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, " ", e)
		}
		os.Exit(1)
	}

	// Execute
	vm := runtime.NewVM()

	// Set up filesystem imports so scripts can import local .py files
	absFilename, _ := filepath.Abs(filename)
	vm.SearchPaths = []string{filepath.Dir(absFilename)}
	vm.FileImporter = func(path string) (*runtime.CodeObject, error) {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		code, errs := compiler.CompileSource(string(src), path)
		if len(errs) > 0 {
			return nil, errs[0]
		}
		return code, nil
	}

	_, err = vm.Execute(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}
