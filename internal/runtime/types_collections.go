package runtime

import (
	"fmt"
	"sort"
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
	orderedKeys   []Value     // Insertion-ordered keys for Python 3.7+ dict ordering
	instanceOwner *PyInstance // if non-nil, sync mutations back to instance's Dict
}

func (d *PyDict) Type() string { return "dict" }
func (d *PyDict) String() string {
	return fmt.Sprintf("%v", d.Items)
}

// ensureIndex builds the hash buckets for a dict whose entries were stored
// directly in the legacy Items map. Keys already in orderedKeys keep their
// order; any others follow in a deterministic order.
func (d *PyDict) ensureIndex(vm *VM) {
	if d.buckets != nil {
		return
	}
	d.buckets = make(map[uint64][]dictEntry, len(d.Items))
	d.size = 0
	if len(d.Items) == 0 {
		d.orderedKeys = nil
		return
	}
	keys := make([]Value, 0, len(d.Items))
	listed := make(map[Value]bool, len(d.orderedKeys))
	for _, k := range d.orderedKeys {
		if _, ok := d.Items[k]; ok && !listed[k] {
			listed[k] = true
			keys = append(keys, k)
		}
	}
	if len(keys) < len(d.Items) {
		var rest []Value
		for k := range d.Items {
			if !listed[k] {
				rest = append(rest, k)
			}
		}
		sort.SliceStable(rest, func(i, j int) bool {
			return fmt.Sprint(rest[i]) < fmt.Sprint(rest[j])
		})
		keys = append(keys, rest...)
	}
	d.orderedKeys = keys
	for _, k := range keys {
		h := vm.hashValueVM(k)
		d.buckets[h] = append(d.buckets[h], dictEntry{key: k, value: d.Items[k]})
		d.size++
	}
}

// DictGet retrieves a value by key using hash-based lookup
func (d *PyDict) DictGet(key Value, vm *VM) (Value, bool) {
	if d.buckets == nil {
		if len(d.Items) == 0 {
			return nil, false
		}
		d.ensureIndex(vm)
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for _, e := range entries {
		if e.key == key || vm.equal(e.key, key) {
			return e.value, true
		}
	}
	return nil, false
}

// DictSet sets a key-value pair using hash-based storage
func (d *PyDict) DictSet(key, value Value, vm *VM) {
	d.ensureIndex(vm)
	if d.Items == nil {
		d.Items = make(map[Value]Value)
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for i, e := range entries {
		if e.key == key || vm.equal(e.key, key) {
			// Existing key: keep the original key object, replace the value
			entries[i].value = value
			d.Items[e.key] = value
			d.syncOwner(e.key, value, true)
			return
		}
	}
	d.buckets[h] = append(entries, dictEntry{key: key, value: value})
	d.size++
	d.orderedKeys = append(d.orderedKeys, key)
	d.Items[key] = value
	d.syncOwner(key, value, true)
}

// syncOwner mirrors a mutation back to the instance whose __dict__ this is.
func (d *PyDict) syncOwner(key, value Value, set bool) {
	if d.instanceOwner == nil {
		return
	}
	if ks, ok := key.(*PyString); ok {
		if set {
			d.instanceOwner.Dict[ks.Value] = value
		} else {
			delete(d.instanceOwner.Dict, ks.Value)
		}
	}
}

// removeOrderedKey removes a key (by identity) from the orderedKeys slice
func (d *PyDict) removeOrderedKey(key Value) {
	for i, k := range d.orderedKeys {
		if k == key {
			copy(d.orderedKeys[i:], d.orderedKeys[i+1:])
			d.orderedKeys[len(d.orderedKeys)-1] = nil // release moved-out tail ref for GC
			d.orderedKeys = d.orderedKeys[:len(d.orderedKeys)-1]
			return
		}
	}
}

// Keys returns keys in insertion order
func (d *PyDict) Keys(vm *VM) []Value {
	if d.buckets == nil && len(d.Items) > 0 {
		d.ensureIndex(vm)
	}
	keys := make([]Value, len(d.orderedKeys))
	copy(keys, d.orderedKeys)
	return keys
}

// DictDelete removes a key using hash-based lookup
func (d *PyDict) DictDelete(key Value, vm *VM) bool {
	if d.buckets == nil {
		if len(d.Items) == 0 {
			return false
		}
		d.ensureIndex(vm)
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for i, e := range entries {
		if e.key == key || vm.equal(e.key, key) {
			// Remove entry, releasing the moved-out tail slot's refs for GC.
			copy(entries[i:], entries[i+1:])
			entries[len(entries)-1] = dictEntry{}
			d.buckets[h] = entries[:len(entries)-1]
			if len(d.buckets[h]) == 0 {
				delete(d.buckets, h)
			}
			d.size--
			delete(d.Items, e.key)
			d.removeOrderedKey(e.key)
			d.syncOwner(e.key, nil, false)
			return true
		}
	}
	return false
}

// DictClear removes all entries
func (d *PyDict) DictClear() {
	d.Items = make(map[Value]Value)
	d.buckets = make(map[uint64][]dictEntry)
	d.orderedKeys = nil
	d.size = 0
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

// Set hash table parameters, mirroring CPython's Objects/setobject.c.
const (
	setMinSize      = 8
	setLinearProbes = 9
	setPerturbShift = 5
)

// setSlot is one slot of a setTable. An empty slot has a nil key and dummy
// unset; a slot whose key was removed keeps dummy set so probe chains through
// it stay intact.
type setSlot struct {
	key   Value
	hash  uint64
	dummy bool
}

// setTable is an open-addressing hash table modeled on CPython's set
// implementation. It uses the same probing, resizing, and dummy-slot rules so
// that iteration order is deterministic and, combined with CPython-compatible
// hashes for numbers and tuples, matches CPython's order for those values.
type setTable struct {
	slots  []setSlot
	fill   int // active + dummy slots
	used   int // active slots
	finger int // search start for pop()
}

func (t *setTable) ensureTable() {
	if t.slots == nil {
		t.slots = make([]setSlot, setMinSize)
	}
}

// Len returns the number of elements.
func (t *setTable) Len() int { return t.used }

// Values returns the elements in iteration order.
func (t *setTable) Values() []Value {
	out := make([]Value, 0, t.used)
	for i := range t.slots {
		if t.slots[i].key != nil {
			out = append(out, t.slots[i].key)
		}
	}
	return out
}

// lookup returns the slot index holding key, or -1 if absent.
func (t *setTable) lookup(key Value, hash uint64, vm *VM) int {
	if t.used == 0 {
		return -1
	}
	mask := uint64(len(t.slots) - 1)
	i := hash & mask
	perturb := hash
	for {
		probes := uint64(0)
		if i+setLinearProbes <= mask {
			probes = setLinearProbes
		}
		for j := i; j <= i+probes; j++ {
			e := &t.slots[j]
			if e.key == nil {
				if !e.dummy {
					return -1
				}
				continue
			}
			if e.hash == hash && (e.key == key || vm.equal(e.key, key)) {
				return int(j)
			}
		}
		perturb >>= setPerturbShift
		i = (i*5 + 1 + perturb) & mask
	}
}

// add inserts key unless an equal key is present. Returns true if inserted.
func (t *setTable) add(key Value, hash uint64, vm *VM) bool {
	t.ensureTable()
	mask := uint64(len(t.slots) - 1)
	i := hash & mask
	perturb := hash
	freeslot := -1
	for {
		probes := uint64(0)
		if i+setLinearProbes <= mask {
			probes = setLinearProbes
		}
		for j := i; j <= i+probes; j++ {
			e := &t.slots[j]
			if e.key == nil {
				if e.dummy {
					// CPython remembers the last dummy on the probe chain
					freeslot = int(j)
					continue
				}
				if freeslot >= 0 {
					// Reuse a dummy slot seen earlier in the probe chain
					t.slots[freeslot] = setSlot{key: key, hash: hash}
					t.used++
					return true
				}
				*e = setSlot{key: key, hash: hash}
				t.fill++
				t.used++
				if uint64(t.fill)*5 >= mask*3 {
					t.resize(t.growSize())
				}
				return true
			}
			if e.hash == hash && (e.key == key || vm.equal(e.key, key)) {
				return false
			}
		}
		perturb >>= setPerturbShift
		i = (i*5 + 1 + perturb) & mask
	}
}

func (t *setTable) growSize() int {
	if t.used > 50000 {
		return t.used * 2
	}
	return t.used * 4
}

// insertClean places a key known to be absent into a table without dummies.
func (t *setTable) insertClean(key Value, hash uint64) {
	mask := uint64(len(t.slots) - 1)
	i := hash & mask
	perturb := hash
	for {
		if t.slots[i].key == nil {
			t.slots[i] = setSlot{key: key, hash: hash}
			return
		}
		if i+setLinearProbes <= mask {
			for j := i + 1; j <= i+setLinearProbes; j++ {
				if t.slots[j].key == nil {
					t.slots[j] = setSlot{key: key, hash: hash}
					return
				}
			}
		}
		perturb >>= setPerturbShift
		i = (i*5 + 1 + perturb) & mask
	}
}

// resize rebuilds the table with the smallest power-of-two size > minused,
// dropping dummy slots.
func (t *setTable) resize(minused int) {
	newsize := setMinSize
	for newsize <= minused {
		newsize <<= 1
	}
	old := t.slots
	t.slots = make([]setSlot, newsize)
	t.fill = t.used
	for i := range old {
		if old[i].key != nil {
			t.insertClean(old[i].key, old[i].hash)
		}
	}
}

// discard removes key if present, leaving a dummy slot. Returns true if removed.
func (t *setTable) discard(key Value, hash uint64, vm *VM) bool {
	idx := t.lookup(key, hash, vm)
	if idx < 0 {
		return false
	}
	t.slots[idx] = setSlot{dummy: true}
	t.used--
	return true
}

// pop removes and returns an arbitrary element (the first after the finger).
func (t *setTable) pop() (Value, bool) {
	if t.used == 0 {
		return nil, false
	}
	mask := len(t.slots) - 1
	i := t.finger & mask
	for t.slots[i].key == nil {
		i++
		if i > mask {
			i = 0
		}
	}
	key := t.slots[i].key
	t.slots[i] = setSlot{dummy: true}
	t.used--
	t.finger = i + 1
	return key, true
}

// clear empties the table.
func (t *setTable) clear() {
	t.slots = nil
	t.fill = 0
	t.used = 0
}

// merge adds every element of other (CPython's set_merge).
func (t *setTable) merge(other *setTable, vm *VM) {
	if other == t || other.used == 0 {
		return
	}
	t.ensureTable()
	if (t.fill+other.used)*5 >= (len(t.slots)-1)*3 {
		t.resize((t.used + other.used) * 2)
	}
	if t.fill == 0 && len(t.slots) == len(other.slots) && other.fill == other.used {
		copy(t.slots, other.slots)
		t.fill = other.fill
		t.used = other.used
		return
	}
	if t.fill == 0 {
		t.fill = other.used
		t.used = other.used
		for i := range other.slots {
			if other.slots[i].key != nil {
				t.insertClean(other.slots[i].key, other.slots[i].hash)
			}
		}
		return
	}
	for i := range other.slots {
		if other.slots[i].key != nil {
			t.add(other.slots[i].key, other.slots[i].hash, vm)
		}
	}
}

// copyTable returns an independent copy (CPython's set_copy).
func (t *setTable) copyTable() setTable {
	var cp setTable
	if t.used == 0 {
		return cp
	}
	cp.merge(t, nil)
	return cp
}

// hashOf returns the CPython-compatible hash of a set element.
func (t *setTable) hashOf(value Value, vm *VM) uint64 {
	if vm == nil {
		return hashValue(value)
	}
	return vm.hashValueVM(value)
}

// PySet represents a Python set
type PySet struct {
	setTable
}

func (s *PySet) Type() string { return "set" }
func (s *PySet) String() string {
	return fmt.Sprintf("%v", s.Values())
}

// SetAdd adds a value to the set
func (s *PySet) SetAdd(value Value, vm *VM) {
	s.add(value, s.hashOf(value, vm), vm)
}

// SetContains checks if a value exists in the set
func (s *PySet) SetContains(value Value, vm *VM) bool {
	return s.lookup(value, s.hashOf(value, vm), vm) >= 0
}

// SetRemove removes a value from the set
func (s *PySet) SetRemove(value Value, vm *VM) bool {
	return s.discard(value, s.hashOf(value, vm), vm)
}

// SetLen returns the number of items
func (s *PySet) SetLen() int {
	return s.used
}

// Copy returns a shallow copy of the set
func (s *PySet) Copy() *PySet {
	return &PySet{setTable: s.copyTable()}
}

// PyFrozenSet represents an immutable Python frozenset
type PyFrozenSet struct {
	setTable
}

func (s *PyFrozenSet) Type() string { return "frozenset" }
func (s *PyFrozenSet) String() string {
	if s.used == 0 {
		return "frozenset()"
	}
	return fmt.Sprintf("frozenset(%v)", s.Values())
}

// FrozenSetAdd adds a value to the frozenset (used during construction)
func (s *PyFrozenSet) FrozenSetAdd(value Value, vm *VM) {
	s.add(value, s.hashOf(value, vm), vm)
}

// FrozenSetContains checks if a value exists in the frozenset
func (s *PyFrozenSet) FrozenSetContains(value Value, vm *VM) bool {
	return s.lookup(value, s.hashOf(value, vm), vm) >= 0
}

// FrozenSetLen returns the number of items
func (s *PyFrozenSet) FrozenSetLen() int {
	return s.used
}

// frozenSetHash computes CPython's frozenset hash from the element hashes.
func frozenSetHash(t *setTable) uint64 {
	shuffle := func(h uint64) uint64 {
		return ((h ^ 89869747) ^ (h << 16)) * 3644798167
	}
	var hash uint64
	for i := range t.slots {
		if t.slots[i].key != nil {
			hash ^= shuffle(t.slots[i].hash)
		}
	}
	hash ^= (uint64(t.used) + 1) * 1927868237
	hash ^= (hash >> 11) ^ (hash >> 25)
	hash = hash*69069 + 907133923
	if hash == ^uint64(0) {
		hash = 590923713
	}
	return hash
}
