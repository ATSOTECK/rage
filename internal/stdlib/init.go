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
	InitOSModule()
	InitDatetimeModule()
	InitTypingModule()
	InitCSVModule()
	InitItertoolsModule()
	InitFunctoolsModule()
	InitBase64Module()
	InitAbcModule()
	InitDataclassesModule()
}
