package runtime

import (
	"fmt"
	"strings"
)

// Iterator

func (vm *VM) getIter(obj Value) (Value, error) {
	// Generators and coroutines are their own iterators
	switch v := obj.(type) {
	case *PyGenerator:
		return v, nil
	case *PyCoroutine:
		return v, nil
	case *PyIterator:
		return v, nil
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
	case *PyIterator:
		if it.Index < len(it.Items) {
			val := it.Items[it.Index]
			it.Index++
			return val, false, nil
		}
		return nil, true, nil

	case *PyGenerator:
		val, done, err := vm.GeneratorSend(it, None)
		if err != nil {
			// StopIteration is not an error for iteration
			if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, done, nil

	case *PyCoroutine:
		val, done, err := vm.CoroutineSend(it, None)
		if err != nil {
			if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
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
				return nil, true, nil
			}
			// Also check for StopIteration from Go function calls (plain error strings)
			if strings.HasPrefix(err.Error(), "StopIteration:") {
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, false, nil
	}
}
