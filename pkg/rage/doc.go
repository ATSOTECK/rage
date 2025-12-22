/*
Package rage provides a public API for embedding the RAGE Python runtime in Go applications.

# Quick Start

The simplest way to run Python code:

	result, err := rage.Run(`print("Hello, World!")`)
	if err != nil {
	    log.Fatal(err)
	}

To evaluate an expression and get the result:

	result, err := rage.Eval(`1 + 2 * 3`)
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Println(result) // 7

# Using State for More Control

For more complex scenarios, create a State:

	state := rage.NewState()
	defer state.Close()

	// Set variables accessible from Python
	state.SetGlobal("name", rage.String("World"))
	state.SetGlobal("count", rage.Int(42))

	// Run Python code
	_, err := state.Run(`
	    greeting = "Hello, " + name + "!"
	    result = count * 2
	`)
	if err != nil {
	    log.Fatal(err)
	}

	// Get variables set by Python
	greeting := state.GetGlobal("greeting")
	fmt.Println(greeting) // Hello, World!

# Controlling Stdlib Modules

By default, NewState() enables all stdlib modules. For more control over which
modules are available, use NewStateWithModules or NewBareState:

	// Create state with only specific modules
	state := rage.NewStateWithModules(
	    rage.WithModule(rage.ModuleMath),
	    rage.WithModule(rage.ModuleString),
	)
	defer state.Close()

	// Or enable multiple modules at once
	state := rage.NewStateWithModules(
	    rage.WithModules(rage.ModuleMath, rage.ModuleString, rage.ModuleTime),
	)

	// Create a bare state with no modules, then enable them later
	state := rage.NewBareState()
	defer state.Close()
	state.EnableModule(rage.ModuleMath)
	state.EnableModules(rage.ModuleString, rage.ModuleTime)

	// Enable all modules on an existing state
	state.EnableAllModules()

Available modules:

	rage.ModuleMath        // math module (sin, cos, sqrt, etc.)
	rage.ModuleRandom      // random module (random, randint, choice, etc.)
	rage.ModuleString      // string module (ascii_letters, digits, etc.)
	rage.ModuleSys         // sys module (version, platform, etc.)
	rage.ModuleTime        // time module (time, sleep, etc.)
	rage.ModuleRe          // re module (match, search, findall, etc.)
	rage.ModuleCollections // collections module (Counter, defaultdict, etc.)

# Registering Go Functions

You can make Go functions callable from Python:

	state := rage.NewState()
	defer state.Close()

	// Register a function
	state.Register("greet", func(s *rage.State, args ...rage.Value) rage.Value {
	    name, _ := rage.AsString(args[0])
	    return rage.String("Hello, " + name + "!")
	})

	// Call it from Python
	result, _ := state.Run(`message = greet("World")`)
	fmt.Println(state.GetGlobal("message")) // Hello, World!

# Working with Values

The rage.Value interface wraps Python values. Use constructors and type assertions:

	// Create values
	intVal := rage.Int(42)
	strVal := rage.String("hello")
	listVal := rage.List(rage.Int(1), rage.Int(2), rage.Int(3))
	dictVal := rage.Dict("name", rage.String("Alice"), "age", rage.Int(30))

	// Convert from Go values
	val := rage.FromGo(map[string]interface{}{"key": "value"})

	// Type checking
	if rage.IsInt(val) {
	    n, _ := rage.AsInt(val)
	    fmt.Println(n)
	}

	// Get underlying Go value
	goVal := val.GoValue()

# Timeouts and Cancellation

Execute code with timeouts to prevent infinite loops:

	result, err := rage.RunWithTimeout(`
	    while True:
	        pass  # infinite loop
	`, 5*time.Second)
	// Returns error after 5 seconds

Or use context for cancellation:

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state := rage.NewState()
	result, err := state.RunWithContext(ctx, `some_long_running_code()`)

# Compilation and Execution

For repeated execution, compile once and run multiple times:

	state := rage.NewState()
	defer state.Close()

	code, err := state.Compile(`result = x * 2`, "multiply.py")
	if err != nil {
	    log.Fatal(err)
	}

	// Execute multiple times with different inputs
	for i := 0; i < 10; i++ {
	    state.SetGlobal("x", rage.Int(int64(i)))
	    state.Execute(code)
	    result := state.GetGlobal("result")
	    fmt.Println(result)
	}

# Error Handling

Compilation errors are returned as *CompileErrors:

	_, err := rage.Run(`invalid python syntax here`)
	if compErr, ok := err.(*rage.CompileErrors); ok {
	    for _, e := range compErr.Errors {
	        fmt.Println(e)
	    }
	}

Runtime errors are returned as standard errors.

# Thread Safety

Each State is NOT safe for concurrent use. Create separate States for concurrent
execution, or use appropriate synchronization.
*/
package rage
