package runtime

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Iterator

// lazyIterator is implemented by the iterators getIter creates for builtin
// sequences that produce their items on demand instead of copying them into a
// list first. advance returns ok=false once the iterator is exhausted.
type lazyIterator interface {
	advance(vm *VM) (val Value, ok bool, err error)
}

// rangeIterator lazily yields the values of a range object, so that
// `for i in range(n)` doesn't materialize a list of n ints.
type rangeIterator struct {
	next      int64
	step      int64
	remaining int64
}

func (it *rangeIterator) Type() string   { return "range_iterator" }
func (it *rangeIterator) String() string { return "<range_iterator object>" }

func newRangeIterator(r *PyRange) *rangeIterator {
	return &rangeIterator{next: r.Start, step: r.Step, remaining: rangeLen(r)}
}

func (it *rangeIterator) advance(*VM) (Value, bool, error) {
	if it.remaining <= 0 {
		return nil, false, nil
	}
	val := MakeInt(it.next)
	it.remaining--
	it.next += it.step
	return val, true, nil
}

// strIterator lazily yields the characters of a string.
type strIterator struct {
	s   string
	pos int
}

func (it *strIterator) Type() string   { return "str_iterator" }
func (it *strIterator) String() string { return "<str_iterator object>" }

// asciiCharStrings caches the one-character strings for ASCII code points,
// which are immutable and so can be shared, to make iterating over text cheap.
var asciiCharStrings [utf8.RuneSelf]*PyString

func init() {
	for i := range asciiCharStrings {
		asciiCharStrings[i] = &PyString{Value: string(rune(i))}
	}
}

func (it *strIterator) advance(*VM) (Value, bool, error) {
	if it.pos >= len(it.s) {
		return nil, false, nil
	}
	if c := it.s[it.pos]; c < utf8.RuneSelf {
		it.pos++
		return asciiCharStrings[c], true, nil
	}
	// Decode like []rune(s) does: invalid bytes become U+FFFD one at a time.
	r, size := utf8.DecodeRuneInString(it.s[it.pos:])
	it.pos += size
	return &PyString{Value: string(r)}, true, nil
}

// dictKeyIterator yields a dict's keys in insertion order and, like CPython,
// raises RuntimeError if the dict changes size while it is being iterated.
type dictKeyIterator struct {
	d    *PyDict
	keys []Value
	idx  int
	size int
}

func (it *dictKeyIterator) Type() string   { return "dict_keyiterator" }
func (it *dictKeyIterator) String() string { return "<dict_keyiterator object>" }

func (it *dictKeyIterator) advance(*VM) (Value, bool, error) {
	if it.d.DictLen() != it.size {
		it.size = -1 // Keep failing on further calls, as CPython does
		return nil, false, fmt.Errorf("RuntimeError: dictionary changed size during iteration")
	}
	if it.idx >= len(it.keys) {
		return nil, false, nil
	}
	val := it.keys[it.idx]
	it.idx++
	return val, true, nil
}

func (vm *VM) getIter(obj Value) (Value, error) {
	// Generators and coroutines are their own iterators
	switch v := obj.(type) {
	case *PyGenerator:
		return v, nil
	case *PyCoroutine:
		return v, nil
	case *PyIterator:
		return v, nil
	case lazyIterator:
		return v, nil
	}

	// Builtin containers have no __iter__ attribute, so handle them before
	// the attribute lookup below (which formats an AttributeError on failure).
	switch v := obj.(type) {
	case *PyList:
		// Reference the list directly so mutations (append, etc.) are
		// visible during iteration
		return &PyIterator{Source: v, Index: 0}, nil
	case *PyRange:
		return newRangeIterator(v), nil
	case *PyTuple:
		// Tuples are immutable, so their items can be iterated in place.
		return &PyIterator{Items: v.Items, Index: 0}, nil
	case *PyString:
		return &strIterator{s: v.Value}, nil
	case *PyDict:
		return &dictKeyIterator{d: v, keys: v.Keys(vm), size: v.DictLen()}, nil
	case *PySet, *PyFrozenSet, *PyBytes:
		items, err := vm.toList(obj)
		if err != nil {
			return nil, err
		}
		return &PyIterator{Items: items, Index: 0}, nil
	}

	// Try __iter__ method first
	if iterMethod, err := vm.getAttr(obj, "__iter__"); err == nil {
		result, err := vm.call(iterMethod, nil, nil)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	// Fall back to converting to list
	items, err := vm.toList(obj)
	if err != nil {
		return nil, err
	}
	return &PyIterator{Items: items, Index: 0}, nil
}

// iterNext advances an iterator and returns the next value
// Returns (value, done, error) where done is true if the iterator is exhausted
func (vm *VM) iterNext(iter Value) (Value, bool, error) {
	switch it := iter.(type) {
	case lazyIterator:
		val, ok, err := it.advance(vm)
		return val, !ok && err == nil, err

	case *PyIterator:
		items := it.Items
		if it.Source != nil {
			items = it.Source.Items
		}
		if it.Index < len(items) {
			val := items[it.Index]
			it.Index++
			return val, false, nil
		}
		return nil, true, nil

	case *PyGenerator:
		val, done, err := vm.GeneratorSend(it, None)
		if err != nil {
			// StopIteration is not an error for iteration
			if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
				vm.currentException = nil // Clear so it doesn't propagate
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, done, nil

	case *PyCoroutine:
		val, done, err := vm.CoroutineSend(it, None)
		if err != nil {
			if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
				vm.currentException = nil // Clear so it doesn't propagate
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, done, nil

	default:
		// Try __next__ method
		nextMethod, err := vm.getAttr(iter, "__next__")
		if err != nil {
			return nil, false, fmt.Errorf("'%s' object is not an iterator", vm.typeName(iter))
		}
		val, err := vm.call(nextMethod, nil, nil)
		if err != nil {
			if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
				vm.currentException = nil // Clear so it doesn't propagate
				return nil, true, nil
			}
			// Also check for StopIteration from Go function calls (plain error strings)
			errStr := err.Error()
			if strings.HasPrefix(errStr, "StopIteration:") || errStr == "StopIteration" ||
				strings.HasPrefix(errStr, "RuntimeError: StopIteration") {
				vm.currentException = nil // Clear so it doesn't propagate
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, false, nil
	}
}
