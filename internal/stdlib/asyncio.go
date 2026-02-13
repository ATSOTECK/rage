package stdlib

import (
	"time"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitAsyncioModule registers the asyncio module
func InitAsyncioModule() {
	runtime.NewModuleBuilder("asyncio").
		Doc("Asynchronous I/O, event loop, and coroutines.").
		Func("run", asyncioRun).
		Func("sleep", asyncioSleep).
		Func("gather", asyncioGather).
		Func("create_task", asyncioCreateTask).
		Register()
}

// asyncio.run(coro) - Run a coroutine to completion
func asyncioRun(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.Push(runtime.None)
		return 1
	}

	arg := vm.Get(1)

	switch coro := arg.(type) {
	case *runtime.PyCoroutine:
		result, err := runCoroutineToCompletion(vm, coro)
		if err != nil {
			// Extract value from StopIteration
			if pyErr, ok := err.(*runtime.PyException); ok && pyErr.Type() == "StopIteration" {
				if pyErr.Args != nil && len(pyErr.Args.Items) > 0 {
					vm.Push(pyErr.Args.Items[0])
					return 1
				}
				vm.Push(runtime.None)
				return 1
			}
			vm.RaiseError("%s", err.Error())
			return 0
		}
		vm.Push(result)
		return 1

	case *runtime.PyGenerator:
		// Allow running generators too (for testing)
		result, err := runGeneratorToCompletion(vm, coro)
		if err != nil {
			if pyErr, ok := err.(*runtime.PyException); ok && pyErr.Type() == "StopIteration" {
				if pyErr.Args != nil && len(pyErr.Args.Items) > 0 {
					vm.Push(pyErr.Args.Items[0])
					return 1
				}
				vm.Push(runtime.None)
				return 1
			}
			vm.RaiseError("%s", err.Error())
			return 0
		}
		vm.Push(result)
		return 1

	default:
		vm.RaiseError("asyncio.run() requires a coroutine")
		return 0
	}
}

// runCoroutineToCompletion runs a coroutine until it returns
func runCoroutineToCompletion(vm *runtime.VM, coro *runtime.PyCoroutine) (runtime.Value, error) {
	for {
		val, done, err := vm.CoroutineSend(coro, runtime.None)
		if err != nil {
			return nil, err
		}
		if done {
			return val, nil
		}
		// If the coroutine yielded, it might be waiting on something
		// For now, just continue (this handles simple await chains)
	}
}

// runGeneratorToCompletion runs a generator until it returns
func runGeneratorToCompletion(vm *runtime.VM, gen *runtime.PyGenerator) (runtime.Value, error) {
	var lastVal runtime.Value = runtime.None
	for {
		val, done, err := vm.GeneratorSend(gen, runtime.None)
		if err != nil {
			return nil, err
		}
		if done {
			return lastVal, nil
		}
		lastVal = val
	}
}

// asyncio.sleep(seconds) - Async sleep (returns an awaitable)
func asyncioSleep(vm *runtime.VM) int {
	seconds := 0.0
	if vm.GetTop() >= 1 {
		switch v := vm.Get(1).(type) {
		case *runtime.PyInt:
			seconds = float64(v.Value)
		case *runtime.PyFloat:
			seconds = v.Value
		}
	}

	// For simplicity, we do a blocking sleep
	// A real implementation would use an event loop
	if seconds > 0 {
		time.Sleep(time.Duration(seconds * float64(time.Second)))
	}

	vm.Push(runtime.None)
	return 1
}

// asyncio.gather(*coros) - Run multiple coroutines concurrently
func asyncioGather(vm *runtime.VM) int {
	top := vm.GetTop()
	if top == 0 {
		vm.Push(&runtime.PyList{Items: []runtime.Value{}})
		return 1
	}

	results := make([]runtime.Value, top)

	// For simplicity, run coroutines sequentially
	// A real implementation would use an event loop for true concurrency
	for i := 1; i <= top; i++ {
		arg := vm.Get(i)
		switch coro := arg.(type) {
		case *runtime.PyCoroutine:
			result, err := runCoroutineToCompletion(vm, coro)
			if err != nil {
				if pyErr, ok := err.(*runtime.PyException); ok && pyErr.Type() == "StopIteration" {
					if pyErr.Args != nil && len(pyErr.Args.Items) > 0 {
						results[i-1] = pyErr.Args.Items[0]
					} else {
						results[i-1] = runtime.None
					}
					continue
				}
				vm.RaiseError("%s", err.Error())
				return 0
			}
			results[i-1] = result

		case *runtime.PyGenerator:
			result, err := runGeneratorToCompletion(vm, coro)
			if err != nil {
				if pyErr, ok := err.(*runtime.PyException); ok && pyErr.Type() == "StopIteration" {
					if pyErr.Args != nil && len(pyErr.Args.Items) > 0 {
						results[i-1] = pyErr.Args.Items[0]
					} else {
						results[i-1] = runtime.None
					}
					continue
				}
				vm.RaiseError("%s", err.Error())
				return 0
			}
			results[i-1] = result

		default:
			// Not a coroutine, just use the value directly
			results[i-1] = arg
		}
	}

	vm.Push(&runtime.PyList{Items: results})
	return 1
}

// asyncio.create_task(coro) - Schedule a coroutine as a task
// For now, this just returns the coroutine itself
func asyncioCreateTask(vm *runtime.VM) int {
	if !vm.RequireArgs("create_task", 1) {
		return 0
	}

	coro := vm.Get(1)
	// In a full implementation, this would wrap the coroutine in a Task object
	// For now, just return the coroutine
	vm.Push(coro)
	return 1
}
