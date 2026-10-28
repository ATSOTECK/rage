package runtime

import (
	"math"
	"testing"
)

func TestFloatReprMatchesCPython(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{1.0, "1.0"},
		{math.Copysign(0, -1), "-0.0"},
		{0.1, "0.1"},
		{1e15, "1000000000000000.0"},
		{1e16, "1e+16"},
		{1.5e-7, "1.5e-07"},
		{0.0001, "0.0001"},
		{0.00001, "1e-05"},
		{1e22, "1e+22"},
		{5e-324, "5e-324"},
		{123456789012345678.0, "1.2345678901234568e+17"},
		{math.Inf(1), "inf"},
		{math.Inf(-1), "-inf"},
		{math.NaN(), "nan"},
	}
	for _, tt := range tests {
		if got := floatRepr(tt.in); got != tt.want {
			t.Errorf("floatRepr(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatFloatTypes(t *testing.T) {
	tests := []struct {
		x     float64
		typ   byte
		prec  int
		flags int
		want  string
	}{
		{1234.5678, 'f', 2, 0, "1234.57"},
		{2.675, 'f', 2, 0, "2.67"},
		{1.5, 'f', 0, fmtFloatAlt, "2."},
		{12345.678, 'e', 2, 0, "1.23e+04"},
		{1e-5, 'g', 6, 0, "1e-05"},
		{1.0, 'g', 6, fmtFloatAlt, "1.00000"},
		{100.0, 'g', 3, fmtFloatAddDot0, "1e+02"},
		{10.0, 'g', 3, fmtFloatAddDot0, "10.0"},
		{-1e-10, 'f', 3, fmtFloatNoNegZero, "0.000"},
		{math.NaN(), 'F', 2, 0, "NAN"},
		{1.0, 'r', 0, fmtFloatSign, "+1"},
	}
	for _, tt := range tests {
		if got := formatFloat(tt.x, tt.typ, tt.prec, tt.flags); got != tt.want {
			t.Errorf("formatFloat(%v, %c, %d, %d) = %q, want %q", tt.x, tt.typ, tt.prec, tt.flags, got, tt.want)
		}
	}
}

func TestStrReprQuoting(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"abc", `'abc'`},
		{"it's", `"it's"`},
		{`'"`, `'\'"'`},
		{"a\nb\t\\", `'a\nb\t\\'`},
		{"\x00\x7f", `'\x00\x7f'`},
		{"caf\u00e9", "'caf\u00e9'"},
		{"\u00a0\u2028", `'\xa0\u2028'`},
	}
	for _, tt := range tests {
		if got := strRepr(tt.in); got != tt.want {
			t.Errorf("strRepr(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestFormatSpecs(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		v    Value
		spec string
		want string
	}{
		{MakeInt(1234567), ",", "1,234,567"},
		{MakeInt(1234), "08,", "0,001,234"},
		{MakeInt(255), "#x", "0xff"},
		{MakeInt(-42), "=10", "-       42"},
		{&PyFloat{Value: 0.5}, "%", "50.000000%"},
		{&PyFloat{Value: 1234567.891}, ",.2f", "1,234,567.89"},
		{&PyFloat{Value: 1.0 / 3}, ".3", "0.333"},
		{MakeComplex(1, 2), ".2f", "1.00+2.00j"},
		{&PyString{Value: "ab"}, "*^6", "**ab**"},
		{True, ">5", "    1"},
	}
	for _, tt := range tests {
		got, err := vm.formatValue(tt.v, tt.spec)
		if err != nil {
			t.Errorf("format(%v, %q) error: %v", tt.v, tt.spec, err)
			continue
		}
		if got != tt.want {
			t.Errorf("format(%v, %q) = %q, want %q", tt.v, tt.spec, got, tt.want)
		}
	}
}

func TestReprCycles(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1)}}
	list.Items = append(list.Items, list)
	if got := vm.repr(list); got != "[1, [...]]" {
		t.Errorf("repr(self-referential list) = %q, want %q", got, "[1, [...]]")
	}
}
