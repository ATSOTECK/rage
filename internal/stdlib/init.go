// Package stdlib provides standard library modules for the RAGE Python runtime.
package stdlib

// InitAllModules initializes all standard library modules.
// Call this before running any Python code that might import standard modules.
func InitAllModules() {
	InitMathModule()
	InitRandomModule()
	InitStringModule()
	InitSysModule()
	InitTimeModule()
	InitReModule()
	InitCollectionsModule()
	InitAsyncioModule()
	InitJSONModule()
	// Add more module initializations here as they are implemented
	// InitOS()
	// etc.
}
