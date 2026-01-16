package stdlib

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitBase64Module registers the base64 module
func InitBase64Module() {
	runtime.NewModuleBuilder("base64").
		Doc("Base16, Base32, Base64 data encodings.").
		Func("b64encode", b64Encode).
		Func("b64decode", b64Decode).
		Func("standard_b64encode", standardB64Encode).
		Func("standard_b64decode", standardB64Decode).
		Func("urlsafe_b64encode", urlsafeB64Encode).
		Func("urlsafe_b64decode", urlsafeB64Decode).
		Func("b32encode", b32Encode).
		Func("b32decode", b32Decode).
		Func("b32hexencode", b32HexEncode).
		Func("b32hexdecode", b32HexDecode).
		Func("b16encode", b16Encode).
		Func("b16decode", b16Decode).
		Func("encodebytes", encodeBytes).
		Func("decodebytes", decodeBytes).
		Register()
}

// base64TypeName returns the Python type name for a value
func base64TypeName(v runtime.Value) string {
	switch v.(type) {
	case *runtime.PyNone:
		return "NoneType"
	case *runtime.PyBool:
		return "bool"
	case *runtime.PyInt:
		return "int"
	case *runtime.PyFloat:
		return "float"
	case *runtime.PyString:
		return "str"
	case *runtime.PyBytes:
		return "bytes"
	case *runtime.PyList:
		return "list"
	case *runtime.PyTuple:
		return "tuple"
	case *runtime.PyDict:
		return "dict"
	default:
		return "object"
	}
}

// getBytes extracts bytes from a string or bytes object
func getBytes(vm *runtime.VM, idx int) ([]byte, bool) {
	val := vm.Get(idx)
	switch v := val.(type) {
	case *runtime.PyBytes:
		return v.Value, true
	case *runtime.PyString:
		return []byte(v.Value), true
	default:
		vm.RaiseError("TypeError: a bytes-like object is required, not '%s'", base64TypeName(val))
		return nil, false
	}
}

// b64Encode encodes bytes using base64.
// b64encode(s, altchars=None) -> bytes
func b64Encode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("b64encode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	// Get custom alphabet if provided
	encoding := base64.StdEncoding
	if vm.GetTop() >= 2 {
		altchars := vm.Get(2)
		if !runtime.IsNone(altchars) {
			altBytes, ok := getBytes(vm, 2)
			if !ok {
				return 0
			}
			if len(altBytes) != 2 {
				vm.RaiseError("TypeError: altchars must be a bytes-like object of length 2")
				return 0
			}
			// Create custom encoding with alternate chars for + and /
			customAlphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789" +
				string(altBytes[0]) + string(altBytes[1])
			encoding = base64.NewEncoding(customAlphabet)
		}
	}

	encoded := encoding.EncodeToString(data)
	vm.Push(runtime.NewBytes([]byte(encoded)))
	return 1
}

// b64Decode decodes base64 encoded bytes.
// b64decode(s, altchars=None, validate=False) -> bytes
func b64Decode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("b64decode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	// Get custom alphabet if provided
	encoding := base64.StdEncoding
	if vm.GetTop() >= 2 {
		altchars := vm.Get(2)
		if !runtime.IsNone(altchars) {
			altBytes, ok := getBytes(vm, 2)
			if !ok {
				return 0
			}
			if len(altBytes) != 2 {
				vm.RaiseError("TypeError: altchars must be a bytes-like object of length 2")
				return 0
			}
			customAlphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789" +
				string(altBytes[0]) + string(altBytes[1])
			encoding = base64.NewEncoding(customAlphabet)
		}
	}

	// Check validate flag (position 3)
	validate := false
	if vm.GetTop() >= 3 {
		validateVal := vm.Get(3)
		if !runtime.IsNone(validateVal) {
			validate = vm.ToBool(3)
		}
	}

	// If validate is true, check for non-base64 characters
	if validate {
		for _, b := range data {
			if !isBase64Char(b) && b != '=' {
				vm.RaiseError("binascii.Error: Invalid base64-encoded string: invalid character")
				return 0
			}
		}
	}

	decoded, err := encoding.DecodeString(string(data))
	if err != nil {
		vm.RaiseError("binascii.Error: Invalid base64-encoded string: %v", err)
		return 0
	}

	vm.Push(runtime.NewBytes(decoded))
	return 1
}

// isBase64Char checks if a byte is a valid base64 character
func isBase64Char(b byte) bool {
	return (b >= 'A' && b <= 'Z') ||
		(b >= 'a' && b <= 'z') ||
		(b >= '0' && b <= '9') ||
		b == '+' || b == '/'
}

// standardB64Encode encodes bytes using the standard base64 alphabet.
// standard_b64encode(s) -> bytes
func standardB64Encode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("standard_b64encode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	vm.Push(runtime.NewBytes([]byte(encoded)))
	return 1
}

// standardB64Decode decodes bytes using the standard base64 alphabet.
// standard_b64decode(s) -> bytes
func standardB64Decode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("standard_b64decode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		vm.RaiseError("binascii.Error: Invalid base64-encoded string: %v", err)
		return 0
	}

	vm.Push(runtime.NewBytes(decoded))
	return 1
}

// urlsafeB64Encode encodes bytes using the URL-safe base64 alphabet.
// urlsafe_b64encode(s) -> bytes
func urlsafeB64Encode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("urlsafe_b64encode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	encoded := base64.URLEncoding.EncodeToString(data)
	vm.Push(runtime.NewBytes([]byte(encoded)))
	return 1
}

// urlsafeB64Decode decodes bytes using the URL-safe base64 alphabet.
// urlsafe_b64decode(s) -> bytes
func urlsafeB64Decode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("urlsafe_b64decode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	decoded, err := base64.URLEncoding.DecodeString(string(data))
	if err != nil {
		vm.RaiseError("binascii.Error: Invalid base64-encoded string: %v", err)
		return 0
	}

	vm.Push(runtime.NewBytes(decoded))
	return 1
}

// b32Encode encodes bytes using base32.
// b32encode(s) -> bytes
func b32Encode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("b32encode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	encoded := base32.StdEncoding.EncodeToString(data)
	vm.Push(runtime.NewBytes([]byte(encoded)))
	return 1
}

// b32Decode decodes base32 encoded bytes.
// b32decode(s, casefold=False, map01=None) -> bytes
func b32Decode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("b32decode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	dataStr := string(data)

	// Check casefold flag (position 2)
	casefold := false
	if vm.GetTop() >= 2 {
		casefoldVal := vm.Get(2)
		if !runtime.IsNone(casefoldVal) {
			casefold = vm.ToBool(2)
		}
	}

	// Check map01 argument (position 3) - maps 0 and 1 to O and I/L
	if vm.GetTop() >= 3 {
		map01Val := vm.Get(3)
		if !runtime.IsNone(map01Val) {
			map01Bytes, ok := getBytes(vm, 3)
			if !ok {
				return 0
			}
			if len(map01Bytes) != 1 {
				vm.RaiseError("TypeError: map01 must be a single character")
				return 0
			}
			// Replace 0 with O and 1 with the specified character
			dataStr = strings.ReplaceAll(dataStr, "0", "O")
			dataStr = strings.ReplaceAll(dataStr, "1", string(map01Bytes[0]))
		}
	}

	if casefold {
		dataStr = strings.ToUpper(dataStr)
	}

	decoded, err := base32.StdEncoding.DecodeString(dataStr)
	if err != nil {
		vm.RaiseError("binascii.Error: Invalid base32-encoded string: %v", err)
		return 0
	}

	vm.Push(runtime.NewBytes(decoded))
	return 1
}

// b32HexEncode encodes bytes using the Extended Hex base32 alphabet.
// b32hexencode(s) -> bytes
func b32HexEncode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("b32hexencode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	encoded := base32.HexEncoding.EncodeToString(data)
	vm.Push(runtime.NewBytes([]byte(encoded)))
	return 1
}

// b32HexDecode decodes Extended Hex base32 encoded bytes.
// b32hexdecode(s, casefold=False) -> bytes
func b32HexDecode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("b32hexdecode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	dataStr := string(data)

	// Check casefold flag (position 2)
	casefold := false
	if vm.GetTop() >= 2 {
		casefoldVal := vm.Get(2)
		if !runtime.IsNone(casefoldVal) {
			casefold = vm.ToBool(2)
		}
	}

	if casefold {
		dataStr = strings.ToUpper(dataStr)
	}

	decoded, err := base32.HexEncoding.DecodeString(dataStr)
	if err != nil {
		vm.RaiseError("binascii.Error: Invalid base32-encoded string: %v", err)
		return 0
	}

	vm.Push(runtime.NewBytes(decoded))
	return 1
}

// b16Encode encodes bytes using base16 (hexadecimal).
// b16encode(s) -> bytes
func b16Encode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("b16encode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	encoded := strings.ToUpper(hex.EncodeToString(data))
	vm.Push(runtime.NewBytes([]byte(encoded)))
	return 1
}

// b16Decode decodes base16 (hexadecimal) encoded bytes.
// b16decode(s, casefold=False) -> bytes
func b16Decode(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("b16decode() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	dataStr := string(data)

	// Check casefold flag (position 2)
	casefold := false
	if vm.GetTop() >= 2 {
		casefoldVal := vm.Get(2)
		if !runtime.IsNone(casefoldVal) {
			casefold = vm.ToBool(2)
		}
	}

	if casefold {
		dataStr = strings.ToUpper(dataStr)
	} else {
		// Without casefold, verify all letters are uppercase
		for _, b := range dataStr {
			if b >= 'a' && b <= 'f' {
				vm.RaiseError("binascii.Error: Non-hexadecimal digit found")
				return 0
			}
		}
	}

	decoded, err := hex.DecodeString(dataStr)
	if err != nil {
		vm.RaiseError("binascii.Error: Non-hexadecimal digit found")
		return 0
	}

	vm.Push(runtime.NewBytes(decoded))
	return 1
}

// encodeBytes encodes bytes using base64 with newlines every 76 characters.
// encodebytes(s) -> bytes
func encodeBytes(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("encodebytes() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	// Add newlines every 76 characters (MIME-style)
	var result strings.Builder
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		result.WriteString(encoded[i:end])
		result.WriteByte('\n')
	}

	vm.Push(runtime.NewBytes([]byte(result.String())))
	return 1
}

// decodeBytes decodes base64 encoded bytes (ignoring newlines).
// decodebytes(s) -> bytes
func decodeBytes(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("decodebytes() missing required argument: 's'")
		return 0
	}

	data, ok := getBytes(vm, 1)
	if !ok {
		return 0
	}

	// Remove all whitespace
	dataStr := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, string(data))

	decoded, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		vm.RaiseError("binascii.Error: Invalid base64-encoded string: %v", err)
		return 0
	}

	vm.Push(runtime.NewBytes(decoded))
	return 1
}
