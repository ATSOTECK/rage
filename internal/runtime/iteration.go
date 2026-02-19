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

	// For lists, create an iterator that references the list directly
	// so mutations (append, etc.) are visible during iteration
	if lst, ok := obj.(*PyList); ok {
		return &PyIterator{Source: lst, Index: 0}, nil
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
			if strings.HasPrefix(err.Error(), "StopIteration:") {
				vm.currentException = nil // Clear so it doesn't propagate
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, false, nil
	}
}
