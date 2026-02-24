package runtime

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
)

// getAttrInt handles attribute access on *PyInt values.
func (vm *VM) getAttrInt(v *PyInt, name string) (Value, error) {
	switch name {
	// Properties
	case "real":
		return v, nil
	case "imag":
		return MakeInt(0), nil
	case "numerator":
		return v, nil
	case "denominator":
		return MakeInt(1), nil

	// Methods
	case "bit_length":
		return &PyBuiltinFunc{Name: "int.bit_length", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if v.BigValue != nil {
				return MakeInt(int64(v.BigValue.BitLen())), nil
			}
			val := v.Value
			if val < 0 {
				val = -val
			}
			if val == 0 {
				return MakeInt(0), nil
			}
			return MakeInt(int64(bits.Len64(uint64(val)))), nil
		}}, nil

	case "bit_count":
		return &PyBuiltinFunc{Name: "int.bit_count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if v.BigValue != nil {
				// Count bits in absolute value
				abs := new(big.Int).Abs(v.BigValue)
				count := 0
				for _, word := range abs.Bits() {
					count += bits.OnesCount(uint(word))
				}
				return MakeInt(int64(count)), nil
			}
			val := v.Value
			if val < 0 {
				val = -val
			}
			return MakeInt(int64(bits.OnesCount64(uint64(val)))), nil
		}}, nil

	case "conjugate":
		return &PyBuiltinFunc{Name: "int.conjugate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return v, nil
		}}, nil

	case "as_integer_ratio":
		return &PyBuiltinFunc{Name: "int.as_integer_ratio", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyTuple{Items: []Value{v, MakeInt(1)}}, nil
		}}, nil

	case "to_bytes":
		return &PyBuiltinFunc{Name: "int.to_bytes", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// to_bytes(length, byteorder, *, signed=False)
			length := 0
			byteorder := ""
			signed := false

			if len(args) >= 1 {
				length = int(vm.toInt(args[0]))
			} else if kv, ok := kwargs["length"]; ok {
				length = int(vm.toInt(kv))
			} else {
				return nil, fmt.Errorf("TypeError: to_bytes() missing required argument: 'length'")
			}

			if len(args) >= 2 {
				if s, ok := args[1].(*PyString); ok {
					byteorder = s.Value
				} else {
					return nil, fmt.Errorf("TypeError: to_bytes() argument 'byteorder' must be str")
				}
			} else if kv, ok := kwargs["byteorder"]; ok {
				if s, ok := kv.(*PyString); ok {
					byteorder = s.Value
				} else {
					return nil, fmt.Errorf("TypeError: to_bytes() argument 'byteorder' must be str")
				}
			} else {
				return nil, fmt.Errorf("TypeError: to_bytes() missing required argument: 'byteorder'")
			}

			if byteorder != "big" && byteorder != "little" {
				return nil, fmt.Errorf("ValueError: byteorder must be either 'little' or 'big'")
			}

			if kv, ok := kwargs["signed"]; ok {
				signed = vm.Truthy(kv)
			}

			val := v.Value
			if v.BigValue != nil {
				if v.BigValue.IsInt64() {
					val = v.BigValue.Int64()
				} else {
					return nil, fmt.Errorf("OverflowError: int too big to convert")
				}
			}

			if val < 0 && !signed {
				return nil, fmt.Errorf("OverflowError: can't convert negative int to unsigned")
			}

			var uval uint64
			if val < 0 {
				// Two's complement for negative values
				uval = uint64(val)
			} else {
				uval = uint64(val)
			}

			result := make([]byte, length)

			if signed && val < 0 {
				// Fill with 0xff for negative numbers (sign extension)
				for i := range result {
					result[i] = 0xff
				}
			}

			// Write bytes in big-endian order first
			for i := length - 1; i >= 0; i-- {
				result[i] = byte(uval & 0xff)
				uval >>= 8
			}

			// Check for overflow: if uval still has bits, the number doesn't fit
			if val >= 0 && uval != 0 {
				return nil, fmt.Errorf("OverflowError: int too big to convert")
			}
			if signed && val < 0 {
				// For negative signed, check that sign bit is set in the MSB
				if length > 0 && result[0]&0x80 == 0 {
					return nil, fmt.Errorf("OverflowError: int too big to convert")
				}
			} else if !signed && val >= 0 {
				// Already checked uval == 0 above
			} else if signed && val >= 0 {
				// Check sign bit isn't set (would look negative)
				if length > 0 && result[0]&0x80 != 0 {
					return nil, fmt.Errorf("OverflowError: int too big to convert")
				}
			}

			if byteorder == "little" {
				// Reverse the bytes
				for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
					result[i], result[j] = result[j], result[i]
				}
			}

			return &PyBytes{Value: result}, nil
		}}, nil

	case "from_bytes":
		return &PyBuiltinFunc{Name: "int.from_bytes", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return intFromBytesImpl(vm, args, kwargs)
		}}, nil

	// Dunder methods
	case "__abs__":
		return &PyBuiltinFunc{Name: "int.__abs__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if v.BigValue != nil {
				return MakeBigInt(new(big.Int).Abs(v.BigValue)), nil
			}
			val := v.Value
			if val < 0 {
				val = -val
			}
			return MakeInt(val), nil
		}}, nil

	case "__bool__":
		return &PyBuiltinFunc{Name: "int.__bool__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if v.BigValue != nil {
				if v.BigValue.Sign() == 0 {
					return False, nil
				}
				return True, nil
			}
			if v.Value == 0 {
				return False, nil
			}
			return True, nil
		}}, nil

	case "__ceil__":
		return &PyBuiltinFunc{Name: "int.__ceil__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return v, nil
		}}, nil

	case "__floor__":
		return &PyBuiltinFunc{Name: "int.__floor__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return v, nil
		}}, nil

	case "__trunc__":
		return &PyBuiltinFunc{Name: "int.__trunc__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return v, nil
		}}, nil

	case "__round__":
		return &PyBuiltinFunc{Name: "int.__round__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// __round__(ndigits=None)
			if len(args) == 0 {
				return v, nil
			}
			if args[0] == None {
				return v, nil
			}
			ndigits := int(vm.toInt(args[0]))
			if ndigits >= 0 {
				return v, nil
			}
			// Negative ndigits: round to nearest 10^(-ndigits) with banker's rounding
			val := v.Value
			pow := int64(1)
			for i := 0; i < -ndigits; i++ {
				pow *= 10
			}
			remainder := val % pow
			truncated := val - remainder
			half := pow / 2

			if remainder < 0 {
				remainder = -remainder
			}

			if remainder > half {
				if val >= 0 {
					truncated += pow
				} else {
					truncated -= pow
				}
			} else if remainder == half {
				// Banker's rounding: round to even
				quotient := truncated / pow
				if quotient%2 != 0 {
					if val >= 0 {
						truncated += pow
					} else {
						truncated -= pow
					}
				}
			}

			return MakeInt(truncated), nil
		}}, nil

	case "__int__":
		return &PyBuiltinFunc{Name: "int.__int__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return v, nil
		}}, nil

	case "__float__":
		return &PyBuiltinFunc{Name: "int.__float__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if v.BigValue != nil {
				f, _ := v.BigValue.Float64()
				return &PyFloat{Value: f}, nil
			}
			return &PyFloat{Value: float64(v.Value)}, nil
		}}, nil

	case "__index__":
		return &PyBuiltinFunc{Name: "int.__index__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return v, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'int' object has no attribute '%s'", name)
}

// getAttrFloat handles attribute access on *PyFloat values.
func (vm *VM) getAttrFloat(f *PyFloat, name string) (Value, error) {
	switch name {
	// Properties
	case "real":
		return f, nil
	case "imag":
		return &PyFloat{Value: 0.0}, nil

	// Methods
	case "is_integer":
		return &PyBuiltinFunc{Name: "float.is_integer", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
				return False, nil
			}
			if f.Value == math.Trunc(f.Value) {
				return True, nil
			}
			return False, nil
		}}, nil

	case "conjugate":
		return &PyBuiltinFunc{Name: "float.conjugate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return f, nil
		}}, nil

	case "hex":
		return &PyBuiltinFunc{Name: "float.hex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyString{Value: floatHex(f.Value)}, nil
		}}, nil

	case "fromhex":
		return &PyBuiltinFunc{Name: "float.fromhex", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return floatFromHexImpl(args)
		}}, nil

	case "as_integer_ratio":
		return &PyBuiltinFunc{Name: "float.as_integer_ratio", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if math.IsInf(f.Value, 0) {
				return nil, fmt.Errorf("OverflowError: cannot convert Infinity to integer ratio")
			}
			if math.IsNaN(f.Value) {
				return nil, fmt.Errorf("ValueError: cannot convert NaN to integer ratio")
			}
			num, den := floatAsIntegerRatio(f.Value)
			return &PyTuple{Items: []Value{num, den}}, nil
		}}, nil

	// Dunder methods
	case "__abs__":
		return &PyBuiltinFunc{Name: "float.__abs__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyFloat{Value: math.Abs(f.Value)}, nil
		}}, nil

	case "__bool__":
		return &PyBuiltinFunc{Name: "float.__bool__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if f.Value == 0.0 {
				return False, nil
			}
			return True, nil
		}}, nil

	case "__ceil__":
		return &PyBuiltinFunc{Name: "float.__ceil__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
				return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
			}
			return MakeInt(int64(math.Ceil(f.Value))), nil
		}}, nil

	case "__floor__":
		return &PyBuiltinFunc{Name: "float.__floor__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
				return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
			}
			return MakeInt(int64(math.Floor(f.Value))), nil
		}}, nil

	case "__trunc__":
		return &PyBuiltinFunc{Name: "float.__trunc__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
				return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
			}
			return MakeInt(int64(math.Trunc(f.Value))), nil
		}}, nil

	case "__round__":
		return &PyBuiltinFunc{Name: "float.__round__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				// No ndigits: return int with banker's rounding
				if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
					return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
				}
				return MakeInt(int64(math.RoundToEven(f.Value))), nil
			}
			if args[0] == None {
				if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
					return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
				}
				return MakeInt(int64(math.RoundToEven(f.Value))), nil
			}
			ndigits := int(vm.toInt(args[0]))
			pow := math.Pow(10, float64(ndigits))
			return &PyFloat{Value: math.RoundToEven(f.Value*pow) / pow}, nil
		}}, nil

	case "__int__":
		return &PyBuiltinFunc{Name: "float.__int__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if math.IsInf(f.Value, 0) || math.IsNaN(f.Value) {
				return nil, fmt.Errorf("OverflowError: cannot convert float %s to integer", formatSpecialFloat(f.Value))
			}
			return MakeInt(int64(f.Value)), nil
		}}, nil

	case "__float__":
		return &PyBuiltinFunc{Name: "float.__float__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return f, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'float' object has no attribute '%s'", name)
}

// getAttrComplex handles attribute access on *PyComplex values.
func (vm *VM) getAttrComplex(o *PyComplex, name string) (Value, error) {
	switch name {
	case "real":
		return &PyFloat{Value: o.Real}, nil
	case "imag":
		return &PyFloat{Value: o.Imag}, nil
	case "conjugate":
		c := o
		return &PyBuiltinFunc{Name: "complex.conjugate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return MakeComplex(c.Real, -c.Imag), nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'complex' object has no attribute '%s'", name)
}

// intFromBytesImpl implements int.from_bytes(bytes, byteorder, *, signed=False)
func intFromBytesImpl(vm *VM, args []Value, kwargs map[string]Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TypeError: from_bytes() missing required argument: 'bytes'")
	}

	var data []byte
	switch b := args[0].(type) {
	case *PyBytes:
		data = b.Value
	case *PyList:
		data = make([]byte, len(b.Items))
		for i, item := range b.Items {
			data[i] = byte(vm.toInt(item))
		}
	case *PyTuple:
		data = make([]byte, len(b.Items))
		for i, item := range b.Items {
			data[i] = byte(vm.toInt(item))
		}
	default:
		return nil, fmt.Errorf("TypeError: cannot convert '%s' object to bytes", vm.typeName(args[0]))
	}

	byteorder := ""
	if len(args) >= 2 {
		if s, ok := args[1].(*PyString); ok {
			byteorder = s.Value
		} else {
			return nil, fmt.Errorf("TypeError: from_bytes() argument 'byteorder' must be str")
		}
	} else if kv, ok := kwargs["byteorder"]; ok {
		if s, ok := kv.(*PyString); ok {
			byteorder = s.Value
		} else {
			return nil, fmt.Errorf("TypeError: from_bytes() argument 'byteorder' must be str")
		}
	} else {
		return nil, fmt.Errorf("TypeError: from_bytes() missing required argument: 'byteorder'")
	}

	if byteorder != "big" && byteorder != "little" {
		return nil, fmt.Errorf("ValueError: byteorder must be either 'little' or 'big'")
	}

	signed := false
	if kv, ok := kwargs["signed"]; ok {
		signed = vm.Truthy(kv)
	}

	// Make a copy to avoid mutating
	b := make([]byte, len(data))
	copy(b, data)

	if byteorder == "little" {
		// Reverse to big-endian for processing
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}

	// Convert bytes to integer (big-endian)
	var result uint64
	if len(b) <= 8 {
		for _, byteVal := range b {
			result = (result << 8) | uint64(byteVal)
		}
		if signed && len(b) > 0 && b[0]&0x80 != 0 {
			// Sign extend
			for i := len(b); i < 8; i++ {
				result |= 0xff << (uint(i) * 8)
			}
			return MakeInt(int64(result)), nil
		}
		return MakeInt(int64(result)), nil
	}

	// For larger values, use big.Int
	bi := new(big.Int).SetBytes(b)
	if signed && len(b) > 0 && b[0]&0x80 != 0 {
		// Subtract 2^(len*8) for negative
		twoN := new(big.Int).Lsh(big.NewInt(1), uint(len(b)*8))
		bi.Sub(bi, twoN)
	}
	return MakeBigInt(bi), nil
}

// formatSpecialFloat returns "infinity" or "NaN" for error messages.
func formatSpecialFloat(f float64) string {
	if math.IsInf(f, 0) {
		return "infinity"
	}
	return "NaN"
}

// floatHex returns the hex representation of a float64, matching Python's float.hex().
func floatHex(f float64) string {
	if math.IsInf(f, 1) {
		return "inf"
	}
	if math.IsInf(f, -1) {
		return "-inf"
	}
	if math.IsNaN(f) {
		return "nan"
	}

	sign := ""
	if math.Signbit(f) {
		sign = "-"
		f = -f
	}

	if f == 0 {
		return sign + "0x0.0000000000000p+0"
	}

	fbits := math.Float64bits(f)
	mantissa := fbits & ((1 << 52) - 1)
	biasedExp := int((fbits >> 52) & 0x7ff)

	if biasedExp == 0 {
		// Subnormal
		return fmt.Sprintf("%s0x0.%013xp-1022", sign, mantissa)
	}
	// Normal
	exp := biasedExp - 1023
	return fmt.Sprintf("%s0x1.%013xp%+d", sign, mantissa, exp)
}

// floatFromHexImpl implements float.fromhex(string).
func floatFromHexImpl(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TypeError: float.fromhex() requires a string argument")
	}
	s, ok := args[0].(*PyString)
	if !ok {
		return nil, fmt.Errorf("TypeError: float.fromhex() argument must be a string")
	}
	str := strings.TrimSpace(s.Value)
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil, fmt.Errorf("ValueError: could not convert string to float: '%s'", s.Value)
	}
	return &PyFloat{Value: f}, nil
}

// floatAsIntegerRatio returns (numerator, denominator) for a finite float.
func floatAsIntegerRatio(f float64) (Value, Value) {
	if f == 0 {
		if math.Signbit(f) {
			return MakeInt(0), MakeInt(1)
		}
		return MakeInt(0), MakeInt(1)
	}

	// Use math.Frexp: f = frac * 2^exp, 0.5 <= |frac| < 1.0
	frac, exp := math.Frexp(f)

	// Multiply frac by 2^53 to get an exact integer mantissa
	// (float64 has 53 bits of significand)
	mantissa := int64(frac * (1 << 53))
	exp -= 53

	// Remove trailing zeros from mantissa to reduce the fraction
	for mantissa != 0 && mantissa%2 == 0 {
		mantissa /= 2
		exp++
	}

	if exp >= 0 {
		// numerator = mantissa * 2^exp, denominator = 1
		if exp <= 62 {
			return MakeInt(mantissa * (1 << uint(exp))), MakeInt(1)
		}
		// Large shift -- use big.Int
		num := new(big.Int).SetInt64(mantissa)
		num.Lsh(num, uint(exp))
		return MakeBigInt(num), MakeInt(1)
	}

	// numerator = mantissa, denominator = 2^(-exp)
	negExp := uint(-exp)
	if negExp <= 62 {
		return MakeInt(mantissa), MakeInt(1 << negExp)
	}
	den := new(big.Int).Lsh(big.NewInt(1), negExp)
	return MakeInt(mantissa), MakeBigInt(den)
}
