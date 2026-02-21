package runtime

import (
	"fmt"
)

// PyList represents a Python list
type PyList struct {
	Items []Value
}

func (l *PyList) Type() string { return "list" }
func (l *PyList) String() string {
	return fmt.Sprintf("%v", l.Items)
}

// PyTuple represents a Python tuple
type PyTuple struct {
	Items []Value
}

func (t *PyTuple) Type() string { return "tuple" }
func (t *PyTuple) String() string {
	return fmt.Sprintf("%v", t.Items)
}

// dictEntry represents a key-value pair in a PyDict
type dictEntry struct {
	key   Value
	value Value
}

// PyDict represents a Python dictionary with hash-based lookups
type PyDict struct {
	Items         map[Value]Value        // Legacy field for compatibility
	buckets       map[uint64][]dictEntry // Hash buckets for O(1) lookup
	size          int
	orderedKeys   []Value // Insertion-ordered keys for Python 3.7+ dict ordering
	instanceOwner *PyInstance            // if non-nil, sync mutations back to instance's Dict
}

func (d *PyDict) Type() string { return "dict" }
func (d *PyDict) String() string {
	return fmt.Sprintf("%v", d.Items)
}

// DictGet retrieves a value by key using hash-based lookup
func (d *PyDict) DictGet(key Value, vm *VM) (Value, bool) {
	if d.buckets == nil {
		// Fall back to legacy Items lookup
		if val, ok := d.Items[key]; ok {
			return val, true
		}
		for k, v := range d.Items {
			if vm.equal(k, key) {
				return v, true
			}
		}
		return nil, false
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for _, e := range entries {
		if vm.equal(e.key, key) {
			return e.value, true
		}
	}
	return nil, false
}

// DictSet sets a key-value pair using hash-based storage
func (d *PyDict) DictSet(key, value Value, vm *VM) {
	if d.buckets == nil {
		d.buckets = make(map[uint64][]dictEntry)
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for i, e := range entries {
		if vm.equal(e.key, key) {
			d.buckets[h][i].value = value
			// Update legacy Items using original key from bucket (value equality)
			if d.Items != nil {
				d.deleteItemByEquality(e.key, vm)
				d.Items[e.key] = value
			}
			// Update value in orderedKeys entry (key stays the same, just update bucket value)
			return
		}
	}
	d.buckets[h] = append(entries, dictEntry{key: key, value: value})
	d.size++
	d.orderedKeys = append(d.orderedKeys, key)
	// Also update legacy Items for compatibility
	if d.Items == nil {
		d.Items = make(map[Value]Value)
	}
	d.deleteItemByEquality(key, vm) // Remove any existing entry with equivalent key
	d.Items[key] = value
	// Sync back to instance dict if this is a __dict__ proxy
	if d.instanceOwner != nil {
		if ks, ok := key.(*PyString); ok {
			d.instanceOwner.Dict[ks.Value] = value
		}
	}
}

// deleteItemByEquality removes a key from legacy Items using value equality
func (d *PyDict) deleteItemByEquality(key Value, vm *VM) {
	for k := range d.Items {
		if vm.equal(k, key) {
			delete(d.Items, k)
			return
		}
	}
}

// removeOrderedKey removes a key from the orderedKeys slice using value equality
func (d *PyDict) removeOrderedKey(key Value, vm *VM) {
	for i, k := range d.orderedKeys {
		if vm.equal(k, key) {
			d.orderedKeys = append(d.orderedKeys[:i], d.orderedKeys[i+1:]...)
			return
		}
	}
}

// Keys returns keys in insertion order
func (d *PyDict) Keys(vm *VM) []Value {
	if len(d.orderedKeys) > 0 {
		return d.orderedKeys
	}
	// Fallback for dicts created without ordered tracking
	var keys []Value
	for k := range d.Items {
		keys = append(keys, k)
	}
	return keys
}

// DictDelete removes a key using hash-based lookup
func (d *PyDict) DictDelete(key Value, vm *VM) bool {
	if d.buckets == nil {
		// Use value equality for legacy Items
		for k := range d.Items {
			if vm.equal(k, key) {
				delete(d.Items, k)
				d.removeOrderedKey(key, vm)
				if d.instanceOwner != nil {
					if ks, ok := key.(*PyString); ok {
						delete(d.instanceOwner.Dict, ks.Value)
					}
				}
				return true
			}
		}
		return false
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for i, e := range entries {
		if vm.equal(e.key, key) {
			// Remove entry by replacing with last and truncating
			d.buckets[h] = append(entries[:i], entries[i+1:]...)
			d.size--
			d.deleteItemByEquality(e.key, vm)
			d.removeOrderedKey(key, vm)
			if d.instanceOwner != nil {
				if ks, ok := key.(*PyString); ok {
					delete(d.instanceOwner.Dict, ks.Value)
				}
			}
			return true
		}
	}
	return false
}

// DictContains checks if a key exists using hash-based lookup
func (d *PyDict) DictContains(key Value, vm *VM) bool {
	_, found := d.DictGet(key, vm)
	return found
}

// DictLen returns the number of items
func (d *PyDict) DictLen() int {
	if d.buckets != nil {
		return d.size
	}
	return len(d.Items)
}

// setEntry represents a value in a PySet
type setEntry struct {
	value Value
}

// PySet represents a Python set with hash-based lookups
type PySet struct {
	Items   map[Value]struct{}    // Legacy field for compatibility
	buckets map[uint64][]setEntry // Hash buckets for O(1) lookup
	size    int
}

func (s *PySet) Type() string { return "set" }
func (s *PySet) String() string {
	return fmt.Sprintf("%v", s.Items)
}

// SetAdd adds a value to the set using hash-based storage
func (s *PySet) SetAdd(value Value, vm *VM) {
	if s.buckets == nil {
		s.buckets = make(map[uint64][]setEntry)
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for _, e := range entries {
		if vm.equal(e.value, value) {
			return // Already exists
		}
	}
	s.buckets[h] = append(entries, setEntry{value: value})
	s.size++
	// Also update legacy Items for compatibility
	if s.Items == nil {
		s.Items = make(map[Value]struct{})
	}
	s.Items[value] = struct{}{}
}

// SetContains checks if a value exists using hash-based lookup
func (s *PySet) SetContains(value Value, vm *VM) bool {
	if s.buckets == nil {
		// Fall back to legacy Items lookup
		if _, ok := s.Items[value]; ok {
			return true
		}
		for k := range s.Items {
			if vm.equal(k, value) {
				return true
			}
		}
		return false
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for _, e := range entries {
		if vm.equal(e.value, value) {
			return true
		}
	}
	return false
}

// SetRemove removes a value from the set
func (s *PySet) SetRemove(value Value, vm *VM) bool {
	if s.buckets == nil {
		delete(s.Items, value)
		return true
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for i, e := range entries {
		if vm.equal(e.value, value) {
			s.buckets[h] = append(entries[:i], entries[i+1:]...)
			s.size--
			delete(s.Items, value)
			return true
		}
	}
	return false
}

// SetLen returns the number of items
func (s *PySet) SetLen() int {
	if s.buckets != nil {
		return s.size
	}
	return len(s.Items)
}

// PyFrozenSet represents an immutable Python frozenset
type PyFrozenSet struct {
	Items   map[Value]struct{}    // Legacy field for compatibility
	buckets map[uint64][]setEntry // Hash buckets for O(1) lookup
	size    int
}

func (s *PyFrozenSet) Type() string { return "frozenset" }
func (s *PyFrozenSet) String() string {
	if len(s.Items) == 0 {
		return "frozenset()"
	}
	return fmt.Sprintf("frozenset(%v)", s.Items)
}

// FrozenSetAdd adds a value to the frozenset (used during construction)
func (s *PyFrozenSet) FrozenSetAdd(value Value, vm *VM) {
	if s.buckets == nil {
		s.buckets = make(map[uint64][]setEntry)
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for _, e := range entries {
		if vm.equal(e.value, value) {
			return // Already exists
		}
	}
	s.buckets[h] = append(entries, setEntry{value: value})
	s.size++
	if s.Items == nil {
		s.Items = make(map[Value]struct{})
	}
	s.Items[value] = struct{}{}
}

// FrozenSetContains checks if a value exists using hash-based lookup
func (s *PyFrozenSet) FrozenSetContains(value Value, vm *VM) bool {
	if s.buckets == nil {
		if _, ok := s.Items[value]; ok {
			return true
		}
		for k := range s.Items {
			if vm.equal(k, value) {
				return true
			}
		}
		return false
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for _, e := range entries {
		if vm.equal(e.value, value) {
			return true
		}
	}
	return false
}

// FrozenSetLen returns the number of items
func (s *PyFrozenSet) FrozenSetLen() int {
	if s.buckets != nil {
		return s.size
	}
	return len(s.Items)
}
