package stdlib

import (
	"strings"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// String constants matching Python's string module
const (
	asciiLowercase = "abcdefghijklmnopqrstuvwxyz"
	asciiUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	asciiLetters   = asciiLowercase + asciiUppercase
	digits         = "0123456789"
	hexdigits      = "0123456789abcdefABCDEF"
	octdigits      = "01234567"
	punctuation    = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	whitespace     = " \t\n\r\v\f"
	printable      = digits + asciiLetters + punctuation + whitespace
)

// InitStringModule registers the string module
func InitStringModule() {
	runtime.NewModuleBuilder("string").
		Doc("String constants and utilities.").
		// Constants
		Const("ascii_lowercase", runtime.NewString(asciiLowercase)).
		Const("ascii_uppercase", runtime.NewString(asciiUppercase)).
		Const("ascii_letters", runtime.NewString(asciiLetters)).
		Const("digits", runtime.NewString(digits)).
		Const("hexdigits", runtime.NewString(hexdigits)).
		Const("octdigits", runtime.NewString(octdigits)).
		Const("punctuation", runtime.NewString(punctuation)).
		Const("whitespace", runtime.NewString(whitespace)).
		Const("printable", runtime.NewString(printable)).
		// Utility functions
		Func("capwords", stringCapwords).
		Register()
}

// capwords(s, sep=None) -> str
// Split into words, capitalize each, and rejoin
func stringCapwords(vm *runtime.VM) int {
	s := vm.CheckString(1)

	// Get optional separator
	sep := " "
	if vm.GetTop() >= 2 {
		sepVal := vm.Get(2)
		if !runtime.IsNone(sepVal) {
			sep = vm.ToString(2)
		}
	}

	// Split, capitalize, and join
	words := strings.Split(s, sep)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	vm.Push(runtime.NewString(strings.Join(words, sep)))
	return 1
}
