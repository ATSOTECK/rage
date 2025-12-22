package main

import (
	"fmt"
	"os"

	"github.com/ATSOTECK/RAGE/internal/compiler"
	"github.com/ATSOTECK/RAGE/internal/runtime"
	"github.com/ATSOTECK/RAGE/internal/stdlib"
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
	_, err = vm.Execute(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}
