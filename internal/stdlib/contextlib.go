package stdlib

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitContextlibModule registers the contextlib module.
func InitContextlibModule() {
	runtime.RegisterModule("contextlib", func(vm *runtime.VM) *runtime.PyModule {
		mod := runtime.NewModule("contextlib")

		// =========================================
		// AbstractContextManager (ABC base class)
		// =========================================
		abstractCM := &runtime.PyClass{
			Name: "AbstractContextManager",
			Dict: make(map[string]runtime.Value),
		}
		abstractCM.Mro = []*runtime.PyClass{abstractCM}

		abstractCM.Dict["__enter__"] = &runtime.PyBuiltinFunc{
			Name: "AbstractContextManager.__enter__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: __enter__() missing 'self' argument")
				}
				return args[0], nil
			},
		}
		abstractCM.Dict["__exit__"] = &runtime.PyBuiltinFunc{
			Name: "AbstractContextManager.__exit__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				return runtime.None, nil
			},
		}

		mod.Dict["AbstractContextManager"] = abstractCM

		// =========================================
		// _GeneratorContextManager class
		// =========================================
		genCMClass := &runtime.PyClass{
			Name: "_GeneratorContextManager",
			Dict: make(map[string]runtime.Value),
		}
		genCMClass.Mro = []*runtime.PyClass{genCMClass}

		genCMClass.Dict["__enter__"] = &runtime.PyBuiltinFunc{
			Name: "_GeneratorContextManager.__enter__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: __enter__() missing 'self' argument")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected _GeneratorContextManager instance")
				}
				gen, ok := self.Dict["gen"].(*runtime.PyGenerator)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected generator")
				}
				// Advance generator to first yield
				val, done, err := vm.GeneratorSend(gen, runtime.None)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, fmt.Errorf("RuntimeError: generator didn't yield")
				}
				return val, nil
			},
		}

		genCMClass.Dict["__exit__"] = &runtime.PyBuiltinFunc{
			Name: "_GeneratorContextManager.__exit__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 4 {
					return nil, fmt.Errorf("TypeError: __exit__() requires 4 arguments")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected _GeneratorContextManager instance")
				}
				gen, ok := self.Dict["gen"].(*runtime.PyGenerator)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected generator")
				}
				excType := args[1]
				excVal := args[2]

				if runtime.IsNone(excType) {
					// No exception — advance generator, expect StopIteration
					_, done, err := vm.GeneratorSend(gen, runtime.None)
					if err != nil {
						// Check for StopIteration (normal exit)
						if pyExc, ok := err.(*runtime.PyException); ok && pyExc.Type() == "StopIteration" {
							return runtime.False, nil
						}
						return nil, err
					}
					if !done {
						return nil, fmt.Errorf("RuntimeError: generator didn't stop")
					}
					return runtime.False, nil
				}

				// Exception occurred — throw it into the generator
				var exc runtime.Value
				if pyExc, ok := excVal.(*runtime.PyException); ok {
					exc = pyExc
				} else {
					exc = excType
				}

				_, done, err := vm.GeneratorThrow(gen, exc, runtime.None)
				if err != nil {
					// If the generator re-raises the same exception, don't suppress
					if pyExc, ok := err.(*runtime.PyException); ok {
						if pyExc.Type() == "StopIteration" {
							// Generator caught the exception and returned normally
							return runtime.True, nil
						}
						// Check if it's the same exception that was thrown in
						if origExc, ok := excVal.(*runtime.PyException); ok {
							if pyExc == origExc {
								// Same exception re-raised, don't suppress
								return runtime.False, nil
							}
						}
						// Different exception raised inside generator
						return nil, pyExc
					}
					return nil, err
				}
				if done {
					// Generator finished (caught the exception and returned)
					return runtime.True, nil
				}
				// Generator yielded again after receiving exception — that's an error
				return nil, fmt.Errorf("RuntimeError: generator didn't stop after throw()")
			},
		}

		// =========================================
		// contextmanager decorator
		// =========================================
		mod.Dict["contextmanager"] = &runtime.PyBuiltinFunc{
			Name: "contextmanager",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: contextmanager() takes exactly 1 argument (%d given)", len(args))
				}
				genFunc := args[0]

				// Return a wrapper that creates _GeneratorContextManager instances
				wrapper := &runtime.PyBuiltinFunc{
					Name: funcName(genFunc),
					Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
						// Call the generator function
						gen, err := vm.Call(genFunc, args, kwargs)
						if err != nil {
							return nil, err
						}
						if _, ok := gen.(*runtime.PyGenerator); !ok {
							return nil, fmt.Errorf("TypeError: contextmanager expected a generator function, but got %s", vm.TypeNameOf(gen))
						}
						inst := &runtime.PyInstance{
							Class: genCMClass,
							Dict: map[string]runtime.Value{
								"gen": gen,
							},
						}
						return inst, nil
					},
				}
				return wrapper, nil
			},
		}

		// =========================================
		// closing class
		// =========================================
		closingClass := &runtime.PyClass{
			Name: "closing",
			Dict: make(map[string]runtime.Value),
		}
		closingClass.Mro = []*runtime.PyClass{closingClass}

		closingClass.Dict["__enter__"] = &runtime.PyBuiltinFunc{
			Name: "closing.__enter__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: __enter__() missing 'self' argument")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected closing instance")
				}
				return self.Dict["thing"], nil
			},
		}

		closingClass.Dict["__exit__"] = &runtime.PyBuiltinFunc{
			Name: "closing.__exit__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: __exit__() missing 'self' argument")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected closing instance")
				}
				thing := self.Dict["thing"]
				closeMethod, err := vm.GetAttr(thing, "close")
				if err != nil {
					return nil, fmt.Errorf("AttributeError: object has no attribute 'close'")
				}
				_, err = vm.Call(closeMethod, nil, nil)
				if err != nil {
					return nil, err
				}
				return runtime.False, nil
			},
		}

		// closing() constructor
		mod.Dict["closing"] = &runtime.PyBuiltinFunc{
			Name: "closing",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: closing() takes exactly 1 argument (%d given)", len(args))
				}
				inst := &runtime.PyInstance{
					Class: closingClass,
					Dict: map[string]runtime.Value{
						"thing": args[0],
					},
				}
				return inst, nil
			},
		}

		// =========================================
		// suppress class
		// =========================================
		suppressClass := &runtime.PyClass{
			Name: "suppress",
			Dict: make(map[string]runtime.Value),
		}
		suppressClass.Mro = []*runtime.PyClass{suppressClass}

		suppressClass.Dict["__enter__"] = &runtime.PyBuiltinFunc{
			Name: "suppress.__enter__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: __enter__() missing 'self' argument")
				}
				return args[0], nil
			},
		}

		suppressClass.Dict["__exit__"] = &runtime.PyBuiltinFunc{
			Name: "suppress.__exit__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 4 {
					return nil, fmt.Errorf("TypeError: __exit__() requires 4 arguments")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected suppress instance")
				}
				excType := args[1]
				excVal := args[2]

				if runtime.IsNone(excType) {
					return runtime.False, nil
				}

				// Get the exceptions to suppress
				exceptionsVal := self.Dict["exceptions"]
				exceptions, ok := exceptionsVal.(*runtime.PyTuple)
				if !ok {
					return runtime.False, nil
				}

				// Check if the exception matches any of the suppressed types
				pyExc, isExc := excVal.(*runtime.PyException)
				if !isExc {
					return runtime.False, nil
				}

				for _, suppType := range exceptions.Items {
					if vm.ExceptionMatches(pyExc, suppType) {
						return runtime.True, nil
					}
				}

				return runtime.False, nil
			},
		}

		// suppress() constructor
		mod.Dict["suppress"] = &runtime.PyBuiltinFunc{
			Name: "suppress",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				inst := &runtime.PyInstance{
					Class: suppressClass,
					Dict: map[string]runtime.Value{
						"exceptions": runtime.NewTuple(args),
					},
				}
				return inst, nil
			},
		}

		// =========================================
		// nullcontext class
		// =========================================
		nullCtxClass := &runtime.PyClass{
			Name: "nullcontext",
			Dict: make(map[string]runtime.Value),
		}
		nullCtxClass.Mro = []*runtime.PyClass{nullCtxClass}

		nullCtxClass.Dict["__enter__"] = &runtime.PyBuiltinFunc{
			Name: "nullcontext.__enter__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: __enter__() missing 'self' argument")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected nullcontext instance")
				}
				return self.Dict["enter_result"], nil
			},
		}

		nullCtxClass.Dict["__exit__"] = &runtime.PyBuiltinFunc{
			Name: "nullcontext.__exit__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				return runtime.False, nil
			},
		}

		// nullcontext() constructor
		mod.Dict["nullcontext"] = &runtime.PyBuiltinFunc{
			Name: "nullcontext",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				var enterResult runtime.Value = runtime.None
				if len(args) > 0 {
					enterResult = args[0]
				}
				if v, ok := kwargs["enter_result"]; ok {
					enterResult = v
				}
				inst := &runtime.PyInstance{
					Class: nullCtxClass,
					Dict: map[string]runtime.Value{
						"enter_result": enterResult,
					},
				}
				return inst, nil
			},
		}

		// =========================================
		// redirect_stdout / redirect_stderr
		// =========================================
		redirectStdoutClass := makeRedirectClass("redirect_stdout", "stdout")
		redirectStderrClass := makeRedirectClass("redirect_stderr", "stderr")

		mod.Dict["redirect_stdout"] = &runtime.PyBuiltinFunc{
			Name: "redirect_stdout",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: redirect_stdout() takes exactly 1 argument (%d given)", len(args))
				}
				inst := &runtime.PyInstance{
					Class: redirectStdoutClass,
					Dict: map[string]runtime.Value{
						"new_target": args[0],
						"old_target": runtime.None,
					},
				}
				return inst, nil
			},
		}

		mod.Dict["redirect_stderr"] = &runtime.PyBuiltinFunc{
			Name: "redirect_stderr",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: redirect_stderr() takes exactly 1 argument (%d given)", len(args))
				}
				inst := &runtime.PyInstance{
					Class: redirectStderrClass,
					Dict: map[string]runtime.Value{
						"new_target": args[0],
						"old_target": runtime.None,
					},
				}
				return inst, nil
			},
		}

		// =========================================
		// ExitStack class
		// =========================================
		exitStackClass := &runtime.PyClass{
			Name: "ExitStack",
			Dict: make(map[string]runtime.Value),
		}
		exitStackClass.Mro = []*runtime.PyClass{exitStackClass}

		exitStackClass.Dict["__enter__"] = &runtime.PyBuiltinFunc{
			Name: "ExitStack.__enter__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: __enter__() missing 'self' argument")
				}
				return args[0], nil
			},
		}

		exitStackClass.Dict["__exit__"] = &runtime.PyBuiltinFunc{
			Name: "ExitStack.__exit__",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 4 {
					return nil, fmt.Errorf("TypeError: __exit__() requires 4 arguments")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected ExitStack instance")
				}

				excType := args[1]
				excVal := args[2]
				excTb := args[3]

				return exitStackExit(vm, self, excType, excVal, excTb)
			},
		}

		exitStackClass.Dict["enter_context"] = &runtime.PyBuiltinFunc{
			Name: "ExitStack.enter_context",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("TypeError: enter_context() requires 2 arguments")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected ExitStack instance")
				}
				cm := args[1]

				// Get __enter__ and __exit__
				enterMethod, err := vm.GetAttr(cm, "__enter__")
				if err != nil {
					return nil, fmt.Errorf("TypeError: %s does not support the context manager protocol", vm.TypeNameOf(cm))
				}
				exitMethod, err := vm.GetAttr(cm, "__exit__")
				if err != nil {
					return nil, fmt.Errorf("TypeError: %s does not support the context manager protocol", vm.TypeNameOf(cm))
				}

				// Call __enter__
				result, err := vm.Call(enterMethod, nil, nil)
				if err != nil {
					return nil, err
				}

				// Push __exit__ callback onto the stack
				pushExitCallback(self, exitMethod, cm)

				return result, nil
			},
		}

		exitStackClass.Dict["push"] = &runtime.PyBuiltinFunc{
			Name: "ExitStack.push",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("TypeError: push() requires 2 arguments")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected ExitStack instance")
				}
				exitObj := args[1]

				// Check if it has __exit__ (context manager)
				exitMethod, err := vm.GetAttr(exitObj, "__exit__")
				if err == nil {
					pushExitCallback(self, exitMethod, exitObj)
					return exitObj, nil
				}

				// Otherwise treat as a plain callback
				pushPlainCallback(self, exitObj)
				return exitObj, nil
			},
		}

		exitStackClass.Dict["callback"] = &runtime.PyBuiltinFunc{
			Name: "ExitStack.callback",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("TypeError: callback() requires at least 2 arguments")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected ExitStack instance")
				}
				callback := args[1]
				cbArgs := args[2:]

				pushCallbackWithArgs(self, callback, cbArgs, kwargs)
				return callback, nil
			},
		}

		exitStackClass.Dict["pop_all"] = &runtime.PyBuiltinFunc{
			Name: "ExitStack.pop_all",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: pop_all() missing 'self' argument")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected ExitStack instance")
				}

				// Create a new ExitStack with the current callbacks
				newInst := &runtime.PyInstance{
					Class: exitStackClass,
					Dict: map[string]runtime.Value{
						"_callbacks": self.Dict["_callbacks"],
					},
				}
				// Clear the current stack
				self.Dict["_callbacks"] = runtime.NewList(nil)

				return newInst, nil
			},
		}

		exitStackClass.Dict["close"] = &runtime.PyBuiltinFunc{
			Name: "ExitStack.close",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("TypeError: close() missing 'self' argument")
				}
				self, ok := args[0].(*runtime.PyInstance)
				if !ok {
					return nil, fmt.Errorf("TypeError: expected ExitStack instance")
				}
				_, err := exitStackExit(vm, self, runtime.None, runtime.None, runtime.None)
				return runtime.None, err
			},
		}

		// ExitStack() constructor
		mod.Dict["ExitStack"] = &runtime.PyBuiltinFunc{
			Name: "ExitStack",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				inst := &runtime.PyInstance{
					Class: exitStackClass,
					Dict: map[string]runtime.Value{
						"_callbacks": runtime.NewList(nil),
					},
				}
				return inst, nil
			},
		}

		return mod
	})
}

// funcName extracts a name from a callable for wrapping.
func funcName(fn runtime.Value) string {
	switch f := fn.(type) {
	case *runtime.PyFunction:
		return f.Name
	case *runtime.PyBuiltinFunc:
		return f.Name
	case *runtime.PyGoFunc:
		return f.Name
	default:
		return "<contextmanager>"
	}
}

// makeRedirectClass creates a redirect_stdout/redirect_stderr class.
func makeRedirectClass(name, streamAttr string) *runtime.PyClass {
	cls := &runtime.PyClass{
		Name: name,
		Dict: make(map[string]runtime.Value),
	}
	cls.Mro = []*runtime.PyClass{cls}

	attr := streamAttr // capture for closures

	cls.Dict["__enter__"] = &runtime.PyBuiltinFunc{
		Name: name + ".__enter__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: __enter__() missing 'self' argument")
			}
			self, ok := args[0].(*runtime.PyInstance)
			if !ok {
				return nil, fmt.Errorf("TypeError: expected %s instance", name)
			}
			// Save is a no-op since we don't have real sys.stdout/stderr objects,
			// but we record the intent for compatibility
			self.Dict["old_target"] = runtime.None
			_ = attr
			return self.Dict["new_target"], nil
		},
	}

	cls.Dict["__exit__"] = &runtime.PyBuiltinFunc{
		Name: name + ".__exit__",
		Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
			// Restore is a no-op — see __enter__
			return runtime.False, nil
		},
	}

	return cls
}

// exitStackCallback represents a callback stored in an ExitStack.
type exitStackCallback struct {
	// For context manager __exit__ methods
	exitMethod runtime.Value
	cm         runtime.Value
	// For plain callbacks
	callback runtime.Value
	cbArgs   []runtime.Value
	cbKwargs map[string]runtime.Value
	isPlain  bool
}

func (e *exitStackCallback) Type() string   { return "_exit_callback" }
func (e *exitStackCallback) String() string { return "<exit callback>" }

// pushExitCallback pushes a context manager's __exit__ onto the ExitStack.
func pushExitCallback(self *runtime.PyInstance, exitMethod, cm runtime.Value) {
	callbacks := self.Dict["_callbacks"].(*runtime.PyList)
	cb := &exitStackCallback{
		exitMethod: exitMethod,
		cm:         cm,
	}
	callbacks.Items = append(callbacks.Items, cb)
}

// pushPlainCallback pushes a plain callable onto the ExitStack.
func pushPlainCallback(self *runtime.PyInstance, callback runtime.Value) {
	callbacks := self.Dict["_callbacks"].(*runtime.PyList)
	cb := &exitStackCallback{
		callback: callback,
		isPlain:  true,
	}
	callbacks.Items = append(callbacks.Items, cb)
}

// pushCallbackWithArgs pushes a callback with args/kwargs onto the ExitStack.
func pushCallbackWithArgs(self *runtime.PyInstance, callback runtime.Value, args []runtime.Value, kwargs map[string]runtime.Value) {
	callbacks := self.Dict["_callbacks"].(*runtime.PyList)
	cb := &exitStackCallback{
		callback: callback,
		cbArgs:   args,
		cbKwargs: kwargs,
		isPlain:  true,
	}
	callbacks.Items = append(callbacks.Items, cb)
}

// exitStackExit processes the ExitStack's __exit__, calling callbacks in LIFO order.
func exitStackExit(vm *runtime.VM, self *runtime.PyInstance, excType, excVal, excTb runtime.Value) (runtime.Value, error) {
	callbacks := self.Dict["_callbacks"].(*runtime.PyList)
	suppressed := false

	// Process callbacks in reverse order (LIFO)
	for i := len(callbacks.Items) - 1; i >= 0; i-- {
		cb, ok := callbacks.Items[i].(*exitStackCallback)
		if !ok {
			continue
		}

		if cb.isPlain {
			// Plain callback — call with stored args
			args := cb.cbArgs
			if args == nil {
				args = []runtime.Value{}
			}
			_, err := vm.Call(cb.callback, args, cb.cbKwargs)
			if err != nil {
				// New exception from callback replaces existing
				excType = runtime.None
				excVal = runtime.None
				excTb = runtime.None
				if pyExc, ok := err.(*runtime.PyException); ok {
					if pyExc.ExcType != nil {
						excType = pyExc.ExcType
					} else {
						excType = runtime.NewString(pyExc.Type())
					}
					excVal = pyExc
				}
				suppressed = false
			}
		} else {
			// Context manager __exit__
			var result runtime.Value
			var err error

			result, err = vm.Call(cb.exitMethod, []runtime.Value{excType, excVal, excTb}, nil)
			if err != nil {
				// New exception from __exit__ replaces existing
				excType = runtime.None
				excVal = runtime.None
				excTb = runtime.None
				if pyExc, ok := err.(*runtime.PyException); ok {
					if pyExc.ExcType != nil {
						excType = pyExc.ExcType
					} else {
						excType = runtime.NewString(pyExc.Type())
					}
					excVal = pyExc
				}
				suppressed = false
				continue
			}

			if vm.Truthy(result) {
				suppressed = true
				excType = runtime.None
				excVal = runtime.None
				excTb = runtime.None
			}
		}
	}

	// Clear the callback stack
	callbacks.Items = nil

	if suppressed {
		return runtime.True, nil
	}
	return runtime.False, nil
}
