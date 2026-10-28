package runtime

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"strconv"
	"strings"
	"unicode"
)

// Arbitrary-precision int arithmetic. Ints are stored as int64 when they fit
// and as *big.Int (PyInt.BigValue) otherwise; MakeBigInt normalizes results
// back to int64. Operations on two small ints use overflow-checked int64 math
// and only fall back to big.Int when the result does not fit.

// maxBigIntBits bounds the size of ints produced by ** and << so that a
// runaway expression fails with MemoryError instead of exhausting memory.
const maxBigIntBits = 1 << 27

// addInt64 returns a+b and whether it did not overflow.
func addInt64(a, b int64) (int64, bool) {
	s := a + b
	return s, (a^s)&(b^s) >= 0
}

// subInt64 returns a-b and whether it did not overflow.
func subInt64(a, b int64) (int64, bool) {
	s := a - b
	return s, (a^b)&(a^s) >= 0
}

// mulInt64 returns a*b and whether it did not overflow.
func mulInt64(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	p := a * b
	if (a == -1 && b == math.MinInt64) || (b == -1 && a == math.MinInt64) || p/b != a {
		return 0, false
	}
	return p, true
}

// powInt64 returns base**exp (exp >= 0) and whether it did not overflow.
func powInt64(base, exp int64) (int64, bool) {
	result := int64(1)
	for exp > 0 {
		if exp&1 == 1 {
			r, ok := mulInt64(result, base)
			if !ok {
				return 0, false
			}
			result = r
		}
		exp >>= 1
		if exp > 0 {
			b, ok := mulInt64(base, base)
			if !ok {
				return 0, false
			}
			base = b
		}
	}
	return result, true
}

// intToFloat converts an int to float, raising OverflowError when it is out
// of float range (as CPython does).
func intToFloat(i *PyInt) (float64, error) {
	if i.BigValue == nil {
		return float64(i.Value), nil
	}
	f, _ := new(big.Float).SetInt(i.BigValue).Float64()
	if math.IsInf(f, 0) {
		return 0, fmt.Errorf("OverflowError: int too large to convert to float")
	}
	return f, nil
}

// intTrueDivide returns a / b correctly rounded, as CPython does.
func intTrueDivide(a, b *PyInt) (Value, error) {
	if b.BigValue == nil && b.Value == 0 {
		return nil, fmt.Errorf("ZeroDivisionError: division by zero")
	}
	const exact = 1 << 53
	if a.BigValue == nil && b.BigValue == nil && a.Value > -exact && a.Value < exact && b.Value > -exact && b.Value < exact {
		return &PyFloat{Value: float64(a.Value) / float64(b.Value)}, nil
	}
	f, _ := new(big.Rat).SetFrac(a.BigIntValue(), b.BigIntValue()).Float64()
	if math.IsInf(f, 0) {
		return nil, fmt.Errorf("OverflowError: integer division result too large for a float")
	}
	return &PyFloat{Value: f}, nil
}

// bigFloorDivMod returns floor(x/y) and x mod y with the sign of y.
func bigFloorDivMod(x, y *big.Int) (*big.Int, *big.Int) {
	q, r := new(big.Int).QuoRem(x, y, new(big.Int))
	if r.Sign() != 0 && (r.Sign() < 0) != (y.Sign() < 0) {
		q.Sub(q, big.NewInt(1))
		r.Add(r, y)
	}
	return q, r
}

// intPowValue computes a ** b for ints, returning a float for negative b.
func intPowValue(a, b *PyInt) (Value, error) {
	if b.BigValue == nil && b.Value < 0 || b.BigValue != nil && b.BigValue.Sign() < 0 {
		af, err := intToFloat(a)
		if err != nil {
			return nil, err
		}
		bf, err := intToFloat(b)
		if err != nil {
			return nil, err
		}
		return floatPow(af, bf)
	}
	if a.BigValue == nil && b.BigValue == nil {
		if r, ok := powInt64(a.Value, b.Value); ok {
			return MakeInt(r), nil
		}
	}
	x := a.BigIntValue()
	// Bases 0, 1 and -1 stay small for any exponent
	if x.CmpAbs(big.NewInt(1)) <= 0 {
		if x.Sign() >= 0 || b.BigIntValue().Bit(0) == 0 {
			if x.Sign() == 0 && b.BigIntValue().Sign() == 0 {
				return MakeInt(1), nil
			}
			return MakeInt(x.Int64() * x.Int64()), nil
		}
		return MakeInt(-1), nil
	}
	if b.BigValue != nil || b.Value > maxBigIntBits/int64(x.BitLen()) {
		return nil, fmt.Errorf("MemoryError: integer result too large")
	}
	return MakeBigInt(new(big.Int).Exp(x, big.NewInt(b.Value), nil)), nil
}

// floatPow computes a ** b for floats with CPython's special cases: zero to a
// negative power raises, a negative base with a fractional exponent gives a
// complex result, and overflow raises OverflowError.
func floatPow(a, b float64) (Value, error) {
	if a == 0 && b < 0 {
		return nil, fmt.Errorf("ZeroDivisionError: zero to a negative power")
	}
	if a < 0 && b != math.Trunc(b) && !math.IsInf(b, 0) && !math.IsInf(a, 0) {
		r := cmplx.Pow(complex(a, 0), complex(b, 0))
		return MakeComplex(real(r), imag(r)), nil
	}
	r := math.Pow(a, b)
	if math.IsInf(r, 0) && !math.IsInf(a, 0) && !math.IsInf(b, 0) {
		return nil, fmt.Errorf("OverflowError: (34, 'Result too large')")
	}
	return &PyFloat{Value: r}, nil
}

// floatDivMod returns floor(a/b) and a mod b using CPython's algorithm, which
// keeps the two consistent (1 // 0.1 == 9.0, 1 % 0.1 == 0.09999999999999995)
// and gives a zero remainder the sign of b.
func floatDivMod(a, b float64) (float64, float64) {
	mod := math.Mod(a, b)
	div := (a - mod) / b
	if mod != 0 {
		if (b < 0) != (mod < 0) {
			mod += b
			div -= 1.0
		}
	} else {
		mod = math.Copysign(0, b)
	}
	var floordiv float64
	if div != 0 {
		floordiv = math.Floor(div)
		if div-floordiv > 0.5 {
			floordiv += 1.0
		}
	} else {
		floordiv = math.Copysign(0, a/b)
	}
	return floordiv, mod
}

// intBinaryOp applies a binary operator to two ints with arbitrary precision.
// ok is false for operators ints do not support. The VM's inline fast paths
// can call this whenever an operand is big or an int64 result overflows.
func (vm *VM) intBinaryOp(op Opcode, a, b *PyInt) (result Value, ok bool, err error) {
	if a.BigValue == nil && b.BigValue == nil {
		x, y := a.Value, b.Value
		switch op {
		case OpBinaryAdd:
			if r, ok := addInt64(x, y); ok {
				return MakeInt(r), true, nil
			}
		case OpBinarySubtract:
			if r, ok := subInt64(x, y); ok {
				return MakeInt(r), true, nil
			}
		case OpBinaryMultiply:
			if r, ok := mulInt64(x, y); ok {
				return MakeInt(r), true, nil
			}
		case OpBinaryFloorDiv, OpBinaryModulo:
			if y == 0 {
				return nil, true, fmt.Errorf("ZeroDivisionError: division by zero")
			}
			if x == math.MinInt64 && y == -1 {
				break // overflows; use big path
			}
			q, r := x/y, x%y
			if r != 0 && (r < 0) != (y < 0) {
				q--
				r += y
			}
			if op == OpBinaryFloorDiv {
				return MakeInt(q), true, nil
			}
			return MakeInt(r), true, nil
		case OpBinaryLShift:
			if y < 0 {
				return nil, true, fmt.Errorf("ValueError: negative shift count")
			}
			if x == 0 {
				return MakeInt(0), true, nil
			}
			if y < 63 {
				if r := x << uint(y); r>>uint(y) == x {
					return MakeInt(r), true, nil
				}
			}
		case OpBinaryRShift:
			if y < 0 {
				return nil, true, fmt.Errorf("ValueError: negative shift count")
			}
			if y > 63 {
				y = 63
			}
			return MakeInt(x >> uint(y)), true, nil
		case OpBinaryAnd:
			return MakeInt(x & y), true, nil
		case OpBinaryOr:
			return MakeInt(x | y), true, nil
		case OpBinaryXor:
			return MakeInt(x ^ y), true, nil
		}
	}

	switch op {
	case OpBinaryDivide:
		r, err := intTrueDivide(a, b)
		return r, true, err
	case OpBinaryPower:
		r, err := intPowValue(a, b)
		return r, true, err
	}

	x, y := a.BigIntValue(), b.BigIntValue()
	switch op {
	case OpBinaryAdd:
		return MakeBigInt(new(big.Int).Add(x, y)), true, nil
	case OpBinarySubtract:
		return MakeBigInt(new(big.Int).Sub(x, y)), true, nil
	case OpBinaryMultiply:
		return MakeBigInt(new(big.Int).Mul(x, y)), true, nil
	case OpBinaryFloorDiv, OpBinaryModulo:
		if y.Sign() == 0 {
			return nil, true, fmt.Errorf("ZeroDivisionError: division by zero")
		}
		q, r := bigFloorDivMod(x, y)
		if op == OpBinaryFloorDiv {
			return MakeBigInt(q), true, nil
		}
		return MakeBigInt(r), true, nil
	case OpBinaryLShift:
		if y.Sign() < 0 {
			return nil, true, fmt.Errorf("ValueError: negative shift count")
		}
		if x.Sign() == 0 {
			return MakeInt(0), true, nil
		}
		if !y.IsInt64() || int64(x.BitLen())+y.Int64() > maxBigIntBits {
			return nil, true, fmt.Errorf("OverflowError: too many digits in integer")
		}
		return MakeBigInt(new(big.Int).Lsh(x, uint(y.Int64()))), true, nil
	case OpBinaryRShift:
		if y.Sign() < 0 {
			return nil, true, fmt.Errorf("ValueError: negative shift count")
		}
		if !y.IsInt64() || y.Int64() > int64(x.BitLen()) {
			if x.Sign() < 0 {
				return MakeInt(-1), true, nil
			}
			return MakeInt(0), true, nil
		}
		return MakeBigInt(new(big.Int).Rsh(x, uint(y.Int64()))), true, nil
	case OpBinaryAnd:
		return MakeBigInt(new(big.Int).And(x, y)), true, nil
	case OpBinaryOr:
		return MakeBigInt(new(big.Int).Or(x, y)), true, nil
	case OpBinaryXor:
		return MakeBigInt(new(big.Int).Xor(x, y)), true, nil
	}
	return nil, false, nil
}

// IntBinaryOp is the exported form of intBinaryOp for the VM fast paths and
// embedders: it applies op to two ints with arbitrary precision.
func (vm *VM) IntBinaryOp(op Opcode, a, b *PyInt) (Value, error) {
	r, ok, err := vm.intBinaryOp(op, a, b)
	if !ok && err == nil {
		return nil, fmt.Errorf("TypeError: unsupported operand type(s) for %s: 'int' and 'int'", binaryOpSymbol(op))
	}
	return r, err
}

// intNegate returns -i, promoting -MinInt64 to a big int.
func intNegate(i *PyInt) *PyInt {
	if i.BigValue == nil && i.Value != math.MinInt64 {
		return MakeInt(-i.Value)
	}
	return MakeBigInt(new(big.Int).Neg(i.BigIntValue()))
}

// intAbs returns |i|.
func intAbs(i *PyInt) *PyInt {
	if i.BigValue == nil && i.Value >= 0 || i.BigValue != nil && i.BigValue.Sign() >= 0 {
		return i
	}
	return intNegate(i)
}

// intIsZero reports whether i == 0.
func intIsZero(i *PyInt) bool {
	return i.BigValue == nil && i.Value == 0
}

// complexQuot divides complex numbers with Smith's algorithm, as CPython does.
func complexQuot(a, b complex128) (complex128, error) {
	ar, ai, br, bi := real(a), imag(a), real(b), imag(b)
	absBr, absBi := math.Abs(br), math.Abs(bi)
	switch {
	case absBr >= absBi:
		if absBr == 0 {
			return 0, fmt.Errorf("ZeroDivisionError: division by zero")
		}
		ratio := bi / br
		denom := br + bi*ratio
		return complex((ar+ai*ratio)/denom, (ai-ar*ratio)/denom), nil
	case absBi >= absBr:
		ratio := br / bi
		denom := br*ratio + bi
		return complex((ar*ratio+ai)/denom, (ai*ratio-ar)/denom), nil
	}
	return complex(math.NaN(), math.NaN()), nil
}

// complexPow computes a ** b like CPython: small integral exponents use
// repeated multiplication (so 1j**2 == -1+0j exactly), others use polar form.
func complexPow(a, b complex128) (complex128, error) {
	if imag(b) == 0 && real(b) == math.Floor(real(b)) && math.Abs(real(b)) <= 100 {
		n := int64(real(b))
		if n < 0 && a == 0 {
			return 0, fmt.Errorf("ZeroDivisionError: zero to a negative or complex power")
		}
		r := complex(1, 0)
		p := a
		for m := absInt64(n); m > 0; m >>= 1 {
			if m&1 == 1 {
				r = complexProd(r, p)
			}
			p = complexProd(p, p)
		}
		if n < 0 {
			return complexQuot(1, r)
		}
		return r, nil
	}
	if b == 0 {
		return 1, nil
	}
	if a == 0 {
		if imag(b) != 0 || real(b) < 0 {
			return 0, fmt.Errorf("ZeroDivisionError: zero to a negative or complex power")
		}
		return 0, nil
	}
	vabs := math.Hypot(real(a), imag(a))
	length := math.Pow(vabs, real(b))
	at := math.Atan2(imag(a), real(a))
	phase := at * real(b)
	if imag(b) != 0 {
		length /= math.Exp(at * imag(b))
		phase += imag(b) * math.Log(vabs)
	}
	return complex(length*math.Cos(phase), length*math.Sin(phase)), nil
}

func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// complexProd multiplies complex numbers with the textbook formula, rounding
// each product separately (explicit conversions prevent FMA fusion).
func complexProd(a, b complex128) complex128 {
	ar, ai, br, bi := real(a), imag(a), real(b), imag(b)
	return complex(float64(ar*br)-float64(ai*bi), float64(ar*bi)+float64(ai*br))
}

// sumValues adds the items produced by next to start the way CPython's sum()
// does: exactly while everything is an int, with compensated summation once
// the total is a float, and with + otherwise. Items are pulled one at a time
// so summing a generator or a range never materializes it.
func (vm *VM) sumValues(start Value, next func() (Value, bool, error)) (Value, error) {
	result := start
	var item Value
	more := func() (bool, error) {
		v, done, err := next()
		if err != nil || done {
			return false, err
		}
		item = v
		return true, nil
	}
	ok, err := more()
	if err != nil {
		return nil, err
	}
	// Exact int phase
	if _, isInt := result.(*PyInt); isInt {
		for ok {
			var it *PyInt
			switch v := item.(type) {
			case *PyInt:
				it = v
			case *PyBool:
				it = boolToInt(v)
			}
			if it == nil {
				break
			}
			r, _, err := vm.intBinaryOp(OpBinaryAdd, result.(*PyInt), it)
			if err != nil {
				return nil, err
			}
			result = r
			if ok, err = more(); err != nil {
				return nil, err
			}
		}
		if ok {
			r, err := vm.binaryOp(OpBinaryAdd, result, item)
			if err != nil {
				return nil, err
			}
			result = r
			if ok, err = more(); err != nil {
				return nil, err
			}
		}
	}
	// Compensated float phase
	if rf, isFloat := result.(*PyFloat); isFloat {
		f, c := rf.Value, 0.0
		add := func(x float64) {
			t := f + x
			if math.Abs(f) >= math.Abs(x) {
				c += (f - t) + x
			} else {
				c += (x - t) + f
			}
			f = t
		}
		for ok {
			if v, isF := item.(*PyFloat); isF {
				add(v.Value)
			} else if v, isI := item.(*PyInt); isI && v.BigValue == nil {
				add(float64(v.Value))
			} else {
				break
			}
			if ok, err = more(); err != nil {
				return nil, err
			}
		}
		if c != 0 && !math.IsInf(c, 0) && !math.IsNaN(c) {
			f += c
		}
		result = &PyFloat{Value: f}
	}
	// Generic phase
	for ok {
		r, err := vm.binaryOp(OpBinaryAdd, result, item)
		if err != nil {
			return nil, err
		}
		result = r
		if ok, err = more(); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// indexInt converts a value to an int via the __index__ protocol, keeping
// arbitrary precision.
func (vm *VM) indexInt(v Value) (*PyInt, error) {
	switch n := v.(type) {
	case *PyInt:
		return n, nil
	case *PyBool:
		return boolToInt(n), nil
	case *PyInstance:
		if result, found, err := vm.callDunder(n, "__index__"); found {
			if err != nil {
				return nil, err
			}
			if i, ok := result.(*PyInt); ok {
				return i, nil
			}
			return nil, fmt.Errorf("TypeError: __index__ returned non-int (type %s)", vm.typeName(result))
		}
	}
	return nil, fmt.Errorf("TypeError: '%s' object cannot be interpreted as an integer", vm.typeName(v))
}

// modPow implements three-argument pow(base, exp, mod) for ints, including
// negative exponents via the modular inverse.
func (vm *VM) modPow(base, exp, mod Value) (Value, error) {
	toInt := func(v Value) (*big.Int, bool) {
		switch n := v.(type) {
		case *PyInt:
			return n.BigIntValue(), true
		case *PyBool:
			return boolToInt(n).BigIntValue(), true
		}
		return nil, false
	}
	b, ok1 := toInt(base)
	e, ok2 := toInt(exp)
	m, ok3 := toInt(mod)
	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("TypeError: pow() 3rd argument not allowed unless all arguments are integers")
	}
	if m.Sign() == 0 {
		return nil, fmt.Errorf("ValueError: pow() 3rd argument cannot be 0")
	}
	absM := new(big.Int).Abs(m)
	if absM.Cmp(big.NewInt(1)) == 0 {
		return MakeInt(0), nil
	}
	b = new(big.Int).Mod(b, absM) // Euclidean: 0 <= b < |m|
	if e.Sign() < 0 {
		inv := new(big.Int).ModInverse(b, absM)
		if inv == nil {
			return nil, fmt.Errorf("ValueError: base is not invertible for the given modulus")
		}
		b = inv
		e = new(big.Int).Neg(e)
	}
	r := new(big.Int).Exp(b, e, absM)
	// The result takes the sign of the modulus, like %
	if m.Sign() < 0 && r.Sign() != 0 {
		r.Sub(r, absM)
	}
	return MakeBigInt(r), nil
}

// divmod implements divmod(a, b).
func (vm *VM) divmod(a, b Value) (Value, error) {
	if inst, ok := a.(*PyInstance); ok {
		if result, found, err := vm.callDunder(inst, "__divmod__", b); found && result != NotImplemented {
			return result, err
		}
	}
	if inst, ok := b.(*PyInstance); ok {
		if result, found, err := vm.callDunder(inst, "__rdivmod__", a); found && result != NotImplemented {
			return result, err
		}
	}
	if ab, ok := a.(*PyBool); ok {
		a = boolToInt(ab)
	}
	if bb, ok := b.(*PyBool); ok {
		b = boolToInt(bb)
	}
	ai, aIsInt := a.(*PyInt)
	bi, bIsInt := b.(*PyInt)
	if aIsInt && bIsInt {
		q, _, err := vm.intBinaryOp(OpBinaryFloorDiv, ai, bi)
		if err != nil {
			return nil, err
		}
		r, _, err := vm.intBinaryOp(OpBinaryModulo, ai, bi)
		if err != nil {
			return nil, err
		}
		return &PyTuple{Items: []Value{q, r}}, nil
	}
	toFloat := func(v Value) (float64, bool, error) {
		switch n := v.(type) {
		case *PyFloat:
			return n.Value, true, nil
		case *PyInt:
			f, err := intToFloat(n)
			return f, true, err
		}
		return 0, false, nil
	}
	af, aOK, err := toFloat(a)
	if err != nil {
		return nil, err
	}
	bf, bOK, err := toFloat(b)
	if err != nil {
		return nil, err
	}
	if !aOK || !bOK {
		return nil, fmt.Errorf("TypeError: unsupported operand type(s) for divmod(): '%s' and '%s'", vm.typeName(a), vm.typeName(b))
	}
	if bf == 0 {
		return nil, fmt.Errorf("ZeroDivisionError: division by zero")
	}
	q, r := floatDivMod(af, bf)
	return &PyTuple{Items: []Value{&PyFloat{Value: q}, &PyFloat{Value: r}}}, nil
}

// floatToInt converts a float to an int (truncating), with CPython's errors
// for infinities and NaN.
func floatToInt(f float64) (*PyInt, error) {
	if math.IsInf(f, 0) {
		return nil, fmt.Errorf("OverflowError: cannot convert float infinity to integer")
	}
	if math.IsNaN(f) {
		return nil, fmt.Errorf("ValueError: cannot convert float NaN to integer")
	}
	if f >= -9.2e18 && f <= 9.2e18 {
		return MakeInt(int64(f)), nil
	}
	i, _ := new(big.Float).SetFloat64(math.Trunc(f)).Int(nil)
	return MakeBigInt(i), nil
}

// roundRatHalfEven rounds q to the nearest integer, ties to even.
func roundRatHalfEven(q *big.Rat) *big.Int {
	n := new(big.Int).Div(q.Num(), q.Denom()) // floor, since Denom > 0
	rem := new(big.Rat).Sub(q, new(big.Rat).SetInt(n))
	if c := rem.Cmp(big.NewRat(1, 2)); c > 0 || (c == 0 && n.Bit(0) == 1) {
		n.Add(n, big.NewInt(1))
	}
	return n
}

// round implements round(x) and round(x, ndigits) for ints and floats.
func (vm *VM) round(x Value, ndigits Value) (Value, error) {
	if b, ok := x.(*PyBool); ok {
		x = boolToInt(b)
	}
	var nd *PyInt
	if ndigits != nil {
		n, err := vm.indexInt(ndigits)
		if err != nil {
			return nil, err
		}
		nd = n
	}
	switch v := x.(type) {
	case *PyInt:
		if nd == nil || nd.BigValue == nil && nd.Value >= 0 || nd.BigValue != nil && nd.BigValue.Sign() > 0 {
			return v, nil
		}
		// Rounding away more digits than the number has gives 0
		if nd.BigValue != nil || -nd.Value > int64(len(v.BigIntValue().String()))+1 {
			return MakeInt(0), nil
		}
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(-nd.Value), nil)
		n := roundRatHalfEven(new(big.Rat).SetFrac(v.BigIntValue(), pow))
		return MakeBigInt(n.Mul(n, pow)), nil
	case *PyFloat:
		f := v.Value
		if nd == nil {
			return floatToInt(math.RoundToEven(f))
		}
		if math.IsInf(f, 0) || math.IsNaN(f) || f == 0 {
			return v, nil
		}
		// Beyond these bounds the result is x itself, or a signed zero
		if nd.BigValue != nil && nd.BigValue.Sign() > 0 || nd.BigValue == nil && nd.Value > 323 {
			return v, nil
		}
		if nd.BigValue != nil || nd.Value < -308 {
			return &PyFloat{Value: math.Copysign(0, f)}, nil
		}
		var r float64
		if nd.Value >= 0 {
			// strconv rounds the exact binary value correctly, like CPython's dtoa
			r, _ = strconv.ParseFloat(strconv.FormatFloat(f, 'f', int(nd.Value), 64), 64)
		} else {
			pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(-nd.Value), nil)
			exact := new(big.Rat).SetFloat64(f)
			n := roundRatHalfEven(exact.Quo(exact, new(big.Rat).SetInt(pow)))
			r, _ = new(big.Float).SetInt(n.Mul(n, pow)).Float64()
		}
		if math.IsInf(r, 0) {
			return nil, fmt.Errorf("OverflowError: rounded value too large to represent")
		}
		if r == 0 {
			r = math.Copysign(0, f)
		}
		return &PyFloat{Value: r}, nil
	}
	return nil, fmt.Errorf("TypeError: type %s doesn't define __round__ method", vm.typeName(x))
}

// Overflow-safe int helpers for the VM's inline fast paths. Each takes the
// cheap int64 route when both operands are small and the result fits, and
// otherwise produces a big int; results are always correct.

// AddInts returns a + b.
func AddInts(a, b *PyInt) *PyInt {
	if a.BigValue == nil && b.BigValue == nil {
		s := a.Value + b.Value
		if (a.Value^s)&(b.Value^s) >= 0 {
			return MakeInt(s)
		}
	}
	return MakeBigInt(new(big.Int).Add(a.BigIntValue(), b.BigIntValue()))
}

// SubInts returns a - b.
func SubInts(a, b *PyInt) *PyInt {
	if a.BigValue == nil && b.BigValue == nil {
		s := a.Value - b.Value
		if (a.Value^b.Value)&(a.Value^s) >= 0 {
			return MakeInt(s)
		}
	}
	return MakeBigInt(new(big.Int).Sub(a.BigIntValue(), b.BigIntValue()))
}

// MulInts returns a * b.
func MulInts(a, b *PyInt) *PyInt {
	if a.BigValue == nil && b.BigValue == nil {
		if p, ok := mulInt64(a.Value, b.Value); ok {
			return MakeInt(p)
		}
	}
	return MakeBigInt(new(big.Int).Mul(a.BigIntValue(), b.BigIntValue()))
}

// CompareInts returns -1, 0 or 1 as a is less than, equal to, or greater than b.
func CompareInts(a, b *PyInt) int {
	if a.BigValue == nil && b.BigValue == nil {
		switch {
		case a.Value < b.Value:
			return -1
		case a.Value > b.Value:
			return 1
		}
		return 0
	}
	return a.BigIntValue().Cmp(b.BigIntValue())
}

// NegateInt returns -a.
func NegateInt(a *PyInt) *PyInt {
	return intNegate(a)
}

// IntToFloat64 converts an int to float64; out-of-range big ints give ±Inf.
func IntToFloat64(a *PyInt) float64 {
	if a.BigValue == nil {
		return float64(a.Value)
	}
	f, _ := new(big.Float).SetInt(a.BigValue).Float64()
	return f
}

// unicodeDigitValue returns the value of a Unicode decimal digit (category
// Nd), relying on Nd digits being allocated in contiguous runs of 0-9.
func unicodeDigitValue(r rune) (int, bool) {
	if r >= '0' && r <= '9' {
		return int(r - '0'), true
	}
	if !unicode.IsDigit(r) {
		return 0, false
	}
	n := 0
	for unicode.IsDigit(r - rune(n) - 1) {
		n++
	}
	return n % 10, true
}

// parseIntString converts a string to an int following the rules of
// CPython's int(s, base): surrounding whitespace, an optional sign, a base
// prefix for bases 0/2/8/16, and single underscores between digits.
func (vm *VM) parseIntString(s string, base int64) (*PyInt, error) {
	invalid := func() error {
		return fmt.Errorf("ValueError: invalid literal for int() with base %d: %s", base, vm.repr(&PyString{Value: s}))
	}
	if base != 0 && (base < 2 || base > 36) {
		return nil, fmt.Errorf("ValueError: int() base must be >= 2 and <= 36, or 0")
	}
	// Map Unicode decimal digits to ASCII, as CPython does before parsing
	var sb strings.Builder
	for _, r := range strings.TrimSpace(s) {
		if v, ok := unicodeDigitValue(r); ok {
			sb.WriteByte(byte('0' + v))
		} else {
			sb.WriteRune(r)
		}
	}
	str := sb.String()
	negative := false
	if str != "" && (str[0] == '+' || str[0] == '-') {
		negative = str[0] == '-'
		str = str[1:]
	}
	effBase := base
	mustBeZero := false
	hasPrefix := func(p byte) bool {
		return len(str) >= 2 && str[0] == '0' && (str[1] == p || str[1] == p-'a'+'A')
	}
	if base == 0 {
		switch {
		case hasPrefix('x'):
			effBase = 16
		case hasPrefix('o'):
			effBase = 8
		case hasPrefix('b'):
			effBase = 2
		default:
			effBase = 10
			// Old C-style octal ("010") is invalid, but zeros are fine
			mustBeZero = len(str) > 0 && str[0] == '0'
		}
	}
	if (effBase == 16 && hasPrefix('x')) || (effBase == 8 && hasPrefix('o')) || (effBase == 2 && hasPrefix('b')) {
		str = str[2:]
		// One underscore is allowed right after the prefix
		if str != "" && str[0] == '_' {
			str = str[1:]
		}
	}
	if str == "" || str[0] == '_' || str[len(str)-1] == '_' || strings.Contains(str, "__") {
		return nil, invalid()
	}
	digits := strings.ReplaceAll(str, "_", "")
	for i := 0; i < len(digits); i++ {
		c := digits[i]
		var d int64
		switch {
		case c >= '0' && c <= '9':
			d = int64(c - '0')
		case c >= 'a' && c <= 'z':
			d = int64(c-'a') + 10
		case c >= 'A' && c <= 'Z':
			d = int64(c-'A') + 10
		default:
			return nil, invalid()
		}
		if d >= effBase {
			return nil, invalid()
		}
	}
	if effBase&(effBase-1) != 0 && len(digits) > maxIntStringDigits {
		return nil, fmt.Errorf("ValueError: Exceeds the limit (%d digits) for integer string conversion: value has %d digits; use sys.set_int_max_str_digits() to increase the limit", maxIntStringDigits, len(digits))
	}
	var result *PyInt
	if u, err := strconv.ParseUint(digits, int(effBase), 63); err == nil {
		result = MakeInt(int64(u))
	} else {
		bi, _ := new(big.Int).SetString(digits, int(effBase))
		result = MakeBigInt(bi)
	}
	if mustBeZero && !intIsZero(result) {
		return nil, invalid()
	}
	if negative {
		result = intNegate(result)
	}
	return result, nil
}

// sortKwargs reads the key= and reverse= keyword arguments of sort/sorted.
func (vm *VM) sortKwargs(name string, kwargs map[string]Value) (Value, bool, error) {
	var keyFn Value
	reverse := false
	for k, v := range kwargs {
		switch k {
		case "key":
			if v != None {
				keyFn = v
			}
		case "reverse":
			reverse = vm.truthy(v)
		default:
			return nil, false, fmt.Errorf("TypeError: %s() got an unexpected keyword argument '%s'", name, k)
		}
	}
	return keyFn, reverse, nil
}

// minMax implements min() and max(); less selects which extreme wins.
func (vm *VM) minMax(name string, args []Value, kwargs map[string]Value) (Value, error) {
	var keyFn, defaultVal Value
	hasDefault := false
	for k, v := range kwargs {
		switch k {
		case "key":
			if v != None {
				keyFn = v
			}
		case "default":
			defaultVal, hasDefault = v, true
		default:
			return nil, fmt.Errorf("TypeError: %s() got an unexpected keyword argument '%s'", name, k)
		}
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("TypeError: %s expected at least 1 argument, got 0", name)
	}
	items := args
	if len(args) == 1 {
		var err error
		if items, err = vm.toList(args[0]); err != nil {
			return nil, err
		}
	} else if hasDefault {
		return nil, fmt.Errorf("TypeError: Cannot specify a default for %s() with multiple positional arguments", name)
	}
	if len(items) == 0 {
		if hasDefault {
			return defaultVal, nil
		}
		return nil, fmt.Errorf("ValueError: %s() iterable argument is empty", name)
	}
	op := OpCompareGt
	if name == "min" {
		op = OpCompareLt
	}
	best := items[0]
	bestKey := best
	if keyFn != nil {
		k, err := vm.call1(keyFn, best)
		if err != nil {
			return nil, err
		}
		bestKey = k
	}
	for _, item := range items[1:] {
		key := item
		if keyFn != nil {
			k, err := vm.call1(keyFn, item)
			if err != nil {
				return nil, err
			}
			key = k
		}
		r, err := vm.compareChecked(op, key, bestKey)
		if err != nil {
			return nil, err
		}
		if vm.truthy(r) {
			best, bestKey = item, key
		}
	}
	return best, nil
}
