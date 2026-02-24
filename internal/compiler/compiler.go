package compiler

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"
)

// CompileError represents a compilation error
type CompileError struct {
	Pos     model.Position
	Message string
}

func (e CompileError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Pos.Filename, e.Pos.Line, e.Pos.Column, e.Message)
}

// Compiler compiles AST to bytecode
type Compiler struct {
	code           *runtime.CodeObject
	symbolTable    *SymbolTable
	errors         []CompileError
	loopStack      []loopInfo
	filename       string
	optimizer      *Optimizer
	currentLine    int // Current source line being compiled
	lineStartOffset int // Bytecode offset where current line started
	finallyDepth   int // Number of enclosing try/finally blocks (for continue/break through finally)
}

type loopInfo struct {
	startOffset   int
	breakJumps    []int
	continueJumps []int
	isForLoop     bool // true for 'for' loops (iterator on stack), false for 'while' loops
}

// newChildCompiler creates a child compiler for compiling nested scopes
// (functions, classes, lambdas, comprehensions).
func (c *Compiler) newChildCompiler(name string, firstLine int, scopeType ScopeType, flags runtime.CodeFlags) *Compiler {
	return &Compiler{
		code: &runtime.CodeObject{
			Name:      name,
			Filename:  c.filename,
			FirstLine: firstLine,
			Flags:     flags,
		},
		symbolTable: NewSymbolTable(scopeType, c.symbolTable),
		filename:    c.filename,
		optimizer:   c.optimizer,
	}
}

// finalizeAndOptimize finishes compilation of a child compiler's code object.
func (c *Compiler) finalizeAndOptimize(child *Compiler) {
	child.finishLineTable()
	child.finalizeCode()
	if c.optimizer != nil {
		c.optimizer.PeepholeOptimize(child.code)
	}
}

// NewCompiler creates a new compiler
func NewCompiler(filename string) *Compiler {
	code := &runtime.CodeObject{
		Name:      "<module>",
		Filename:  filename,
		FirstLine: 1,
	}
	return &Compiler{
		code:        code,
		symbolTable: NewSymbolTable(ScopeModule, nil),
		filename:    filename,
		optimizer:   NewOptimizer(),
	}
}

// Compile compiles a module to bytecode
func (c *Compiler) Compile(module *model.Module) (*runtime.CodeObject, []CompileError) {
	stmts := module.Body

	for _, stmt := range stmts {
		c.compileStmt(stmt)
	}

	// Add implicit return None at end of module
	c.emit(runtime.OpLoadNone) // Use optimized opcode
	c.emit(runtime.OpReturn)

	// Finalize line number table
	c.finishLineTable()

	// Build names and varnames lists
	c.finalizeCode()

	// Apply peephole optimizations
	c.optimizer.PeepholeOptimize(c.code)

	// Validate bytecode indices
	if err := c.code.Validate(); err != nil {
		c.errors = append(c.errors, CompileError{
			Message: err.Error(),
		})
	}

	return c.code, c.errors
}

func (c *Compiler) error(pos model.Position, format string, args ...any) {
	c.errors = append(c.errors, CompileError{
		Pos:     pos,
		Message: fmt.Sprintf(format, args...),
	})
}

// Bytecode emission helpers

func (c *Compiler) emit(op runtime.Opcode) int {
	offset := len(c.code.Code)
	c.code.Code = append(c.code.Code, byte(op))
	return offset
}

func (c *Compiler) emitArg(op runtime.Opcode, arg int) int {
	// Check if argument fits in 16 bits (signed or unsigned)
	// Allow -32768 to 65535 range: negative values wrap to unsigned 16-bit
	// (e.g., -1 becomes 65535/0xFFFF for special sentinel values)
	if arg < -32768 || arg > 65535 {
		// For very large arguments, extended arg support could be added here
		// For now, report an error for overflow cases
		c.errors = append(c.errors, CompileError{
			Message: fmt.Sprintf("bytecode argument %d exceeds 16-bit limit", arg),
		})
		// Clamp to valid range to avoid corrupted bytecode
		if arg < -32768 {
			arg = -32768
		} else {
			arg = 65535
		}
	}
	offset := len(c.code.Code)
	c.code.Code = append(c.code.Code, byte(op), byte(arg), byte(arg>>8))
	return offset
}

func (c *Compiler) emitJump(op runtime.Opcode) int {
	return c.emitArg(op, 0) // Placeholder, will be patched
}

func (c *Compiler) patchJump(offset int, target int) {
	if offset+2 >= len(c.code.Code) {
		c.errors = append(c.errors, CompileError{
			Message: fmt.Sprintf("patchJump: offset %d out of bounds (code length %d)", offset, len(c.code.Code)),
		})
		return
	}
	c.code.Code[offset+1] = byte(target)
	c.code.Code[offset+2] = byte(target >> 8)
}

func (c *Compiler) currentOffset() int {
	return len(c.code.Code)
}

func (c *Compiler) addConstant(value any) int {
	// Skip deduplication for slice types (they can't be compared with ==)
	switch value.(type) {
	case []string, []any, []int, []float64, []byte:
		// Don't deduplicate slices, just add them
	case *runtime.PyInt:
		// Don't deduplicate PyInt (may contain big.Int which panics on ==)
	default:
		for i, v := range c.code.Constants {
			if v == value {
				return i
			}
		}
	}
	c.code.Constants = append(c.code.Constants, value)
	return len(c.code.Constants) - 1
}

func (c *Compiler) addName(name string) int {
	for i, n := range c.code.Names {
		if n == name {
			return i
		}
	}
	c.code.Names = append(c.code.Names, name)
	return len(c.code.Names) - 1
}

func (c *Compiler) emitLoadConst(value any) {
	idx := c.addConstant(value)
	c.emitArg(runtime.OpLoadConst, idx)
}

// setLine updates the current line number being compiled.
// When the line changes, it records the previous line's bytecode range.
func (c *Compiler) setLine(line int) {
	if line <= 0 {
		return
	}
	if line != c.currentLine {
		// Finish the previous line's entry if there was one
		if c.currentLine > 0 {
			currentOffset := len(c.code.Code)
			if currentOffset > c.lineStartOffset {
				c.code.LineNoTab = append(c.code.LineNoTab, runtime.LineEntry{
					StartOffset: c.lineStartOffset,
					EndOffset:   currentOffset,
					Line:        c.currentLine,
				})
			}
		}
		c.currentLine = line
		c.lineStartOffset = len(c.code.Code)
	}
}

// finishLineTable finalizes the line number table with the last entry
func (c *Compiler) finishLineTable() {
	if c.currentLine > 0 {
		currentOffset := len(c.code.Code)
		if currentOffset > c.lineStartOffset {
			c.code.LineNoTab = append(c.code.LineNoTab, runtime.LineEntry{
				StartOffset: c.lineStartOffset,
				EndOffset:   currentOffset,
				Line:        c.currentLine,
			})
		}
	}
}

func (c *Compiler) finalizeCode() {
	// Build VarNames list for local variables
	for name, sym := range c.symbolTable.symbols {
		if sym.Scope == ScopeLocal && sym.Index >= 0 {
			// Ensure VarNames has enough capacity
			for len(c.code.VarNames) <= sym.Index {
				c.code.VarNames = append(c.code.VarNames, "")
			}
			c.code.VarNames[sym.Index] = name
		}
	}

	// Build CellVars list (variables captured by inner functions)
	// Also add cell variables to VarNames at their original positions
	// so the VM can match parameter names to cells
	for _, sym := range c.symbolTable.cellSyms {
		c.code.CellVars = append(c.code.CellVars, sym.Name)
		// Add to VarNames at the original position (before it became a cell)
		if sym.OriginalIndex >= 0 {
			for len(c.code.VarNames) <= sym.OriginalIndex {
				c.code.VarNames = append(c.code.VarNames, "")
			}
			c.code.VarNames[sym.OriginalIndex] = sym.Name
		}
	}

	// Build FreeVars list
	for _, sym := range c.symbolTable.freeSyms {
		c.code.FreeVars = append(c.code.FreeVars, sym.Name)
	}

	// Calculate stack size (simplified estimate)
	c.code.StackSize = c.estimateStackSize()
}

func (c *Compiler) estimateStackSize() int {
	// Conservative estimate based on code length
	maxStack := 10
	code := c.code.Code
	for i := 0; i < len(code); {
		op := runtime.Opcode(code[i])
		var arg int
		if op.HasArg() {
			if i+2 < len(code) {
				arg = int(code[i+1]) | int(code[i+2])<<8
			}
			i += 3
		} else {
			i++
		}
		// Certain ops increase stack needs
		switch op {
		case runtime.OpBuildList, runtime.OpBuildTuple, runtime.OpBuildSet, runtime.OpBuildMap:
			if arg > maxStack {
				maxStack = arg + 10
			}
		}
	}
	return maxStack
}

// CompileSource compiles Python source code to a code object
func CompileSource(source, filename string) (*runtime.CodeObject, []error) {
	parser := NewParser(source)
	module, parseErrors := parser.Parse()

	if len(parseErrors) > 0 {
		var errs []error
		for _, e := range parseErrors {
			errs = append(errs, e)
		}
		return nil, errs
	}

	compiler := NewCompiler(filename)
	code, compileErrors := compiler.Compile(module)

	if len(compileErrors) > 0 {
		var errs []error
		for _, e := range compileErrors {
			errs = append(errs, e)
		}
		return nil, errs
	}

	return code, nil
}

// indexByte returns the index of the first occurrence of c in s, or -1 if not present.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
