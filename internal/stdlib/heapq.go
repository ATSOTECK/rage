package stdlib

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitHeapqModule registers the heapq module.
func InitHeapqModule() {
	runtime.RegisterModule("heapq", func(vm *runtime.VM) *runtime.PyModule {
		mod := runtime.NewModule("heapq")

		mod.Dict["heappush"] = &runtime.PyBuiltinFunc{
			Name: "heappush",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("TypeError: heappush() takes exactly 2 arguments (%d given)", len(args))
				}
				heap, ok := args[0].(*runtime.PyList)
				if !ok {
					return nil, fmt.Errorf("TypeError: heap argument must be a list")
				}
				heap.Items = append(heap.Items, args[1])
				if err := siftDown(vm, heap, 0, len(heap.Items)-1); err != nil {
					return nil, err
				}
				return runtime.None, nil
			},
		}

		mod.Dict["heappop"] = &runtime.PyBuiltinFunc{
			Name: "heappop",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: heappop() takes exactly 1 argument (%d given)", len(args))
				}
				heap, ok := args[0].(*runtime.PyList)
				if !ok {
					return nil, fmt.Errorf("TypeError: heap argument must be a list")
				}
				if len(heap.Items) == 0 {
					return nil, fmt.Errorf("IndexError: index out of range")
				}
				n := len(heap.Items)
				// Swap first and last, pop last, sift first down
				heap.Items[0], heap.Items[n-1] = heap.Items[n-1], heap.Items[0]
				result := heap.Items[n-1]
				heap.Items = heap.Items[:n-1]
				if len(heap.Items) > 0 {
					if err := siftUp(vm, heap, 0); err != nil {
						return nil, err
					}
				}
				return result, nil
			},
		}

		mod.Dict["heapify"] = &runtime.PyBuiltinFunc{
			Name: "heapify",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: heapify() takes exactly 1 argument (%d given)", len(args))
				}
				heap, ok := args[0].(*runtime.PyList)
				if !ok {
					return nil, fmt.Errorf("TypeError: heap argument must be a list")
				}
				n := len(heap.Items)
				for i := n/2 - 1; i >= 0; i-- {
					if err := siftUp(vm, heap, i); err != nil {
						return nil, err
					}
				}
				return runtime.None, nil
			},
		}

		mod.Dict["heapreplace"] = &runtime.PyBuiltinFunc{
			Name: "heapreplace",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("TypeError: heapreplace() takes exactly 2 arguments (%d given)", len(args))
				}
				heap, ok := args[0].(*runtime.PyList)
				if !ok {
					return nil, fmt.Errorf("TypeError: heap argument must be a list")
				}
				if len(heap.Items) == 0 {
					return nil, fmt.Errorf("IndexError: index out of range")
				}
				result := heap.Items[0]
				heap.Items[0] = args[1]
				if err := siftUp(vm, heap, 0); err != nil {
					return nil, err
				}
				return result, nil
			},
		}

		mod.Dict["heappushpop"] = &runtime.PyBuiltinFunc{
			Name: "heappushpop",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 2 {
					return nil, fmt.Errorf("TypeError: heappushpop() takes exactly 2 arguments (%d given)", len(args))
				}
				heap, ok := args[0].(*runtime.PyList)
				if !ok {
					return nil, fmt.Errorf("TypeError: heap argument must be a list")
				}
				item := args[1]
				if len(heap.Items) > 0 {
					lt, err := heapLt(vm, heap.Items[0], item)
					if err != nil {
						return nil, err
					}
					if lt {
						// heap[0] is smaller than item, swap them
						item, heap.Items[0] = heap.Items[0], item
						if err := siftUp(vm, heap, 0); err != nil {
							return nil, err
						}
					}
				}
				return item, nil
			},
		}

		mod.Dict["nlargest"] = &runtime.PyBuiltinFunc{
			Name: "nlargest",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 2 || len(args) > 3 {
					return nil, fmt.Errorf("TypeError: nlargest() requires 2 to 3 positional arguments (%d given)", len(args))
				}
				nVal, ok := args[0].(*runtime.PyInt)
				if !ok {
					return nil, fmt.Errorf("TypeError: n must be an integer")
				}
				n := int(nVal.Value)

				var keyFn runtime.Value
				if len(args) == 3 {
					keyFn = args[2]
				}
				if v, ok := kwargs["key"]; ok {
					keyFn = v
				}

				items, err := vm.ToList(args[1])
				if err != nil {
					return nil, err
				}

				if n <= 0 {
					return runtime.NewList(nil), nil
				}
				if n >= len(items) {
					n = len(items)
				}

				// Build key-decorated list if key function provided
				deco := make([]heapDecorated, len(items))
				for i, item := range items {
					k := item
					if keyFn != nil && !runtime.IsNone(keyFn) {
						k, err = vm.Call(keyFn, []runtime.Value{item}, nil)
						if err != nil {
							return nil, err
						}
					}
					deco[i] = heapDecorated{key: k, val: item}
				}

				// Use a min-heap of size n to find the n largest
				heap := make([]heapDecorated, 0, n)
				for _, d := range deco {
					if len(heap) < n {
						heap = append(heap, d)
						decoSiftDown(vm, heap, 0, len(heap)-1)
					} else {
						lt, err := heapLt(vm, d.key, heap[0].key)
						if err != nil {
							return nil, err
						}
						if !lt {
							heap[0] = d
							decoSiftUp(vm, heap, 0)
						}
					}
				}

				// Sort result in descending order
				result := make([]runtime.Value, len(heap))
				for i := len(heap) - 1; i >= 0; i-- {
					result[i] = heap[0].val
					last := len(heap) - 1
					heap[0] = heap[last]
					heap = heap[:last]
					if len(heap) > 0 {
						decoSiftUp(vm, heap, 0)
					}
				}

				return runtime.NewList(result), nil
			},
		}

		mod.Dict["nsmallest"] = &runtime.PyBuiltinFunc{
			Name: "nsmallest",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 2 || len(args) > 3 {
					return nil, fmt.Errorf("TypeError: nsmallest() requires 2 to 3 positional arguments (%d given)", len(args))
				}
				nVal, ok := args[0].(*runtime.PyInt)
				if !ok {
					return nil, fmt.Errorf("TypeError: n must be an integer")
				}
				n := int(nVal.Value)

				var keyFn runtime.Value
				if len(args) == 3 {
					keyFn = args[2]
				}
				if v, ok := kwargs["key"]; ok {
					keyFn = v
				}

				items, err := vm.ToList(args[1])
				if err != nil {
					return nil, err
				}

				if n <= 0 {
					return runtime.NewList(nil), nil
				}
				if n >= len(items) {
					n = len(items)
				}

				// Build key-decorated list if key function provided
				deco := make([]heapDecorated, len(items))
				for i, item := range items {
					k := item
					if keyFn != nil && !runtime.IsNone(keyFn) {
						k, err = vm.Call(keyFn, []runtime.Value{item}, nil)
						if err != nil {
							return nil, err
						}
					}
					deco[i] = heapDecorated{key: k, val: item}
				}

				// Use a max-heap of size n to find the n smallest
				heap := make([]heapDecorated, 0, n)
				for _, d := range deco {
					if len(heap) < n {
						heap = append(heap, d)
						decoMaxSiftDown(vm, heap, 0, len(heap)-1)
					} else {
						lt, err := heapLt(vm, d.key, heap[0].key)
						if err != nil {
							return nil, err
						}
						if lt {
							heap[0] = d
							decoMaxSiftUp(vm, heap, 0)
						}
					}
				}

				// Sort result in ascending order
				result := make([]runtime.Value, len(heap))
				for i := len(heap) - 1; i >= 0; i-- {
					result[i] = heap[0].val
					last := len(heap) - 1
					heap[0] = heap[last]
					heap = heap[:last]
					if len(heap) > 0 {
						decoMaxSiftUp(vm, heap, 0)
					}
				}

				return runtime.NewList(result), nil
			},
		}

		mod.Dict["merge"] = &runtime.PyBuiltinFunc{
			Name: "merge",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				var keyFn runtime.Value
				reverse := false
				if v, ok := kwargs["key"]; ok {
					keyFn = v
				}
				if v, ok := kwargs["reverse"]; ok {
					reverse = vm.Truthy(v)
				}

				// Collect all iterables into slices
				iterables := make([][]runtime.Value, 0, len(args))
				for _, arg := range args {
					items, err := vm.ToList(arg)
					if err != nil {
						return nil, err
					}
					if len(items) > 0 {
						iterables = append(iterables, items)
					}
				}

				// Use indices to track position in each iterable
				indices := make([]int, len(iterables))

				// Build merged result
				var result []runtime.Value
				for {
					// Find the iterable with the smallest (or largest if reverse) next element
					bestIdx := -1
					var bestKey runtime.Value
					for i, items := range iterables {
						if indices[i] >= len(items) {
							continue
						}
						item := items[indices[i]]
						k := item
						if keyFn != nil && !runtime.IsNone(keyFn) {
							var err error
							k, err = vm.Call(keyFn, []runtime.Value{item}, nil)
							if err != nil {
								return nil, err
							}
						}
						if bestIdx == -1 {
							bestIdx = i
							bestKey = k
						} else {
							lt, err := heapLt(vm, k, bestKey)
							if err != nil {
								return nil, err
							}
							if reverse {
								lt = !lt
							}
							if lt {
								bestIdx = i
								bestKey = k
							}
						}
					}
					if bestIdx == -1 {
						break
					}
					result = append(result, iterables[bestIdx][indices[bestIdx]])
					indices[bestIdx]++
				}

				return runtime.NewList(result), nil
			},
		}

		return mod
	})
}

// heapLt returns true if a < b using Python comparison.
func heapLt(vm *runtime.VM, a, b runtime.Value) (bool, error) {
	result := vm.CompareOp(runtime.OpCompareLt, a, b)
	if result == nil {
		return false, fmt.Errorf("TypeError: '<' not supported between instances of '%s' and '%s'",
			vm.TypeNameOf(a), vm.TypeNameOf(b))
	}
	return vm.Truthy(result), nil
}

// siftDown is the "sift-down" operation (CPython's _siftdown).
// Restores heap invariant by moving heap[pos] up toward root.
func siftDown(vm *runtime.VM, heap *runtime.PyList, startPos, pos int) error {
	newItem := heap.Items[pos]
	for pos > startPos {
		parentPos := (pos - 1) >> 1
		parent := heap.Items[parentPos]
		lt, err := heapLt(vm, newItem, parent)
		if err != nil {
			return err
		}
		if lt {
			heap.Items[pos] = parent
			pos = parentPos
		} else {
			break
		}
	}
	heap.Items[pos] = newItem
	return nil
}

// siftUp is the "sift-up" operation (CPython's _siftup).
// Restores heap invariant by moving heap[pos] down toward leaves.
func siftUp(vm *runtime.VM, heap *runtime.PyList, pos int) error {
	endPos := len(heap.Items)
	startPos := pos
	newItem := heap.Items[pos]
	// Bubble up the smaller child until hitting a leaf.
	childPos := 2*pos + 1
	for childPos < endPos {
		// Set childPos to the smaller child.
		rightPos := childPos + 1
		if rightPos < endPos {
			lt, err := heapLt(vm, heap.Items[childPos], heap.Items[rightPos])
			if err != nil {
				return err
			}
			if !lt {
				childPos = rightPos
			}
		}
		// Move the smaller child up.
		heap.Items[pos] = heap.Items[childPos]
		pos = childPos
		childPos = 2*pos + 1
	}
	// The leaf at pos is empty now. Put newItem there, and sift it up.
	heap.Items[pos] = newItem
	return siftDown(vm, heap, startPos, pos)
}

// decorated type for nlargest/nsmallest helpers
type heapDecorated struct {
	key runtime.Value
	val runtime.Value
}

// decoSiftDown for min-heap of decorated values (used by nlargest)
func decoSiftDown(vm *runtime.VM, heap []heapDecorated, startPos, pos int) {
	newItem := heap[pos]
	for pos > startPos {
		parentPos := (pos - 1) >> 1
		parent := heap[parentPos]
		lt, _ := heapLt(vm, newItem.key, parent.key)
		if lt {
			heap[pos] = parent
			pos = parentPos
		} else {
			break
		}
	}
	heap[pos] = newItem
}

// decoSiftUp for min-heap of decorated values (used by nlargest)
func decoSiftUp(vm *runtime.VM, heap []heapDecorated, pos int) {
	endPos := len(heap)
	startPos := pos
	newItem := heap[pos]
	childPos := 2*pos + 1
	for childPos < endPos {
		rightPos := childPos + 1
		if rightPos < endPos {
			lt, _ := heapLt(vm, heap[childPos].key, heap[rightPos].key)
			if !lt {
				childPos = rightPos
			}
		}
		heap[pos] = heap[childPos]
		pos = childPos
		childPos = 2*pos + 1
	}
	heap[pos] = newItem
	decoSiftDown(vm, heap, startPos, pos)
}

// decoMaxSiftDown for max-heap of decorated values (used by nsmallest)
func decoMaxSiftDown(vm *runtime.VM, heap []heapDecorated, startPos, pos int) {
	newItem := heap[pos]
	for pos > startPos {
		parentPos := (pos - 1) >> 1
		parent := heap[parentPos]
		lt, _ := heapLt(vm, parent.key, newItem.key)
		if lt {
			heap[pos] = parent
			pos = parentPos
		} else {
			break
		}
	}
	heap[pos] = newItem
}

// decoMaxSiftUp for max-heap of decorated values (used by nsmallest)
func decoMaxSiftUp(vm *runtime.VM, heap []heapDecorated, pos int) {
	endPos := len(heap)
	startPos := pos
	newItem := heap[pos]
	childPos := 2*pos + 1
	for childPos < endPos {
		rightPos := childPos + 1
		if rightPos < endPos {
			lt, _ := heapLt(vm, heap[childPos].key, heap[rightPos].key)
			if lt {
				childPos = rightPos
			}
		}
		heap[pos] = heap[childPos]
		pos = childPos
		childPos = 2*pos + 1
	}
	heap[pos] = newItem
	decoMaxSiftDown(vm, heap, startPos, pos)
}
