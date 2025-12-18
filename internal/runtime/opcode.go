package runtime

import "fmt"

// Opcode represents a single bytecode instruction
type Opcode byte

const (
	// Stack manipulation
	OpPop  Opcode = iota // Pop top of stack
	OpDup                // Duplicate top of stack
	OpRot2               // Swap top two stack items
	OpRot3               // Rotate top three stack items

	// Constants and variables
	OpLoadConst    // Load constant from constants pool (arg: index)
	OpLoadName     // Load variable by name (arg: name index)
	OpStoreName    // Store to variable by name (arg: name index)
	OpDeleteName   // Delete variable by name (arg: name index)
	OpLoadFast     // Load local variable (arg: local index)
	OpStoreFast    // Store to local variable (arg: local index)
	OpDeleteFast   // Delete local variable (arg: local index)
	OpLoadGlobal   // Load global variable (arg: name index)
	OpStoreGlobal  // Store to global variable (arg: name index)
	OpDeleteGlobal // Delete global variable (arg: name index)
	OpLoadAttr     // Load attribute (arg: name index)
	OpStoreAttr    // Store attribute (arg: name index)
	OpDeleteAttr   // Delete attribute (arg: name index)

	// Subscript operations
	OpBinarySubscr // a[b] -> push a[b]
	OpStoreSubscr  // a[b] = c
	OpDeleteSubscr // del a[b]

	// Arithmetic operations
	OpUnaryPositive // +a
	OpUnaryNegative // -a
	OpUnaryNot      // not a
	OpUnaryInvert   // ~a

	OpBinaryAdd      // a + b
	OpBinarySubtract // a - b
	OpBinaryMultiply // a * b
	OpBinaryDivide   // a / b
	OpBinaryFloorDiv // a // b
	OpBinaryModulo   // a % b
	OpBinaryPower    // a ** b
	OpBinaryMatMul   // a @ b
	OpBinaryLShift   // a << b
	OpBinaryRShift   // a >> b
	OpBinaryAnd      // a & b
	OpBinaryOr       // a | b
	OpBinaryXor      // a ^ b

	// In-place operations
	OpInplaceAdd      // a += b
	OpInplaceSubtract // a -= b
	OpInplaceMultiply // a *= b
	OpInplaceDivide   // a /= b
	OpInplaceFloorDiv // a //= b
	OpInplaceModulo   // a %= b
	OpInplacePower    // a **= b
	OpInplaceMatMul   // a @= b
	OpInplaceLShift   // a <<= b
	OpInplaceRShift   // a >>= b
	OpInplaceAnd      // a &= b
	OpInplaceOr       // a |= b
	OpInplaceXor      // a ^= b

	// Comparison operations
	OpCompareEq    // a == b
	OpCompareNe    // a != b
	OpCompareLt    // a < b
	OpCompareLe    // a <= b
	OpCompareGt    // a > b
	OpCompareGe    // a >= b
	OpCompareIs    // a is b
	OpCompareIsNot // a is not b
	OpCompareIn    // a in b
	OpCompareNotIn // a not in b

	// Boolean operations
	OpJumpIfTrueOrPop  // Jump if true, else pop (arg: offset)
	OpJumpIfFalseOrPop // Jump if false, else pop (arg: offset)

	// Control flow
	OpJump           // Unconditional jump (arg: offset)
	OpJumpIfTrue     // Jump if top is true (arg: offset)
	OpJumpIfFalse    // Jump if top is false (arg: offset)
	OpPopJumpIfTrue  // Pop and jump if true (arg: offset)
	OpPopJumpIfFalse // Pop and jump if false (arg: offset)

	// Iteration
	OpGetIter // Get iterator from iterable
	OpForIter // Get next from iterator or jump (arg: offset)

	// Function operations
	OpMakeFunction // Create function object (arg: flags)
	OpCall         // Call function (arg: arg count)
	OpCallKw       // Call function with keyword args (arg: arg count)
	OpReturn       // Return from function

	// Class operations
	OpLoadBuildClass // Load __build_class__
	OpLoadMethod     // Load method for call optimization
	OpCallMethod     // Call method (arg: arg count)

	// Collection building
	OpBuildTuple  // Build tuple (arg: count)
	OpBuildList   // Build list (arg: count)
	OpBuildSet    // Build set (arg: count)
	OpBuildMap    // Build dict (arg: count of key-value pairs)
	OpBuildString // Build string from parts (arg: count)

	// Unpacking
	OpUnpackSequence // Unpack sequence (arg: count)
	OpUnpackEx       // Unpack with star (arg: counts before/after)

	// List/set/dict comprehensions
	OpListAppend // Append to list for comprehension
	OpSetAdd     // Add to set for comprehension
	OpMapAdd     // Add to map for comprehension

	// Import operations
	OpImportName // Import module (arg: name index)
	OpImportFrom // Import from module (arg: name index)
	OpImportStar // Import * from module

	// Exception handling
	OpSetupExcept  // Setup exception handler (arg: handler offset)
	OpSetupFinally // Setup finally handler (arg: handler offset)
	OpPopExcept    // Pop exception handler
	OpEndFinally   // End finally block
	OpRaiseVarargs // Raise exception (arg: count 0-3)

	// With statement
	OpSetupWith   // Setup with statement (arg: cleanup offset)
	OpWithCleanup // Cleanup with statement

	// Assertion
	OpAssert // Assert with optional message

	// Closure operations
	OpLoadClosure  // Load closure variable
	OpStoreClosure // Store to closure variable
	OpLoadDeref    // Load from cell
	OpStoreDeref   // Store to cell
	OpMakeCell     // Create cell for variable

	// Misc
	OpNop        // No operation
	OpPrintExpr  // Print expression result (REPL)
	OpLoadLocals // Load locals dict

	// ==========================================
	// Specialized/Optimized opcodes (no args)
	// ==========================================

	// Specialized LOAD_FAST for common indices (no arg needed)
	OpLoadFast0 // Load local variable 0
	OpLoadFast1 // Load local variable 1
	OpLoadFast2 // Load local variable 2
	OpLoadFast3 // Load local variable 3

	// Specialized STORE_FAST for common indices (no arg needed)
	OpStoreFast0 // Store to local variable 0
	OpStoreFast1 // Store to local variable 1
	OpStoreFast2 // Store to local variable 2
	OpStoreFast3 // Store to local variable 3

	// Specialized constant loading (no arg needed)
	OpLoadNone  // Load None
	OpLoadTrue  // Load True
	OpLoadFalse // Load False
	OpLoadZero  // Load integer 0
	OpLoadOne   // Load integer 1

	// Specialized arithmetic (no arg needed)
	OpIncrementFast // Increment local by 1 (arg: local index) - has arg!
	OpDecrementFast // Decrement local by 1 (arg: local index) - has arg!

	// Superinstructions - combined operations
	OpLoadFastLoadFast   // Load two locals (arg: packed indices)
	OpLoadFastLoadConst  // Load local then const (arg: packed indices)
	OpStoreFastLoadFast  // Store then load (arg: packed indices)
	OpBinaryAddInt       // Add two ints (optimized path)
	OpBinarySubtractInt  // Subtract two ints (optimized path)
	OpBinaryMultiplyInt  // Multiply two ints (optimized path)
	OpCompareLtInt       // Compare less than for ints
	OpCompareLeInt       // Compare less than or equal for ints
	OpCompareGtInt       // Compare greater than for ints
	OpCompareGeInt       // Compare greater than or equal for ints
	OpCompareEqInt       // Compare equal for ints
	OpCompareNeInt       // Compare not equal for ints

	// Empty collection loading (no arg needed)
	OpLoadEmptyList  // Load empty list []
	OpLoadEmptyTuple // Load empty tuple ()
	OpLoadEmptyDict  // Load empty dict {}

	// Combined compare and jump (superinstructions)
	OpCompareLtJump // Compare < and jump if false (arg: offset)
	OpCompareLeJump // Compare <= and jump if false (arg: offset)
	OpCompareGtJump // Compare > and jump if false (arg: offset)
	OpCompareGeJump // Compare >= and jump if false (arg: offset)
	OpCompareEqJump // Compare == and jump if false (arg: offset)
	OpCompareNeJump // Compare != and jump if false (arg: offset)

	// Inline builtins (no function call overhead)
	OpLenList   // len(list) - inline
	OpLenString // len(string) - inline
	OpLenTuple  // len(tuple) - inline
	OpLenDict   // len(dict) - inline
	OpLenGeneric // len() - generic but optimized path

	// More superinstructions
	OpLoadConstLoadFast  // Load const then local (arg: packed indices)
	OpLoadGlobalLoadFast // Load global then local (arg: packed indices)
)

// OpcodeNames maps opcodes to their string names for debugging
var OpcodeNames = map[Opcode]string{
	OpPop:              "POP",
	OpDup:              "DUP",
	OpRot2:             "ROT_TWO",
	OpRot3:             "ROT_THREE",
	OpLoadConst:        "LOAD_CONST",
	OpLoadName:         "LOAD_NAME",
	OpStoreName:        "STORE_NAME",
	OpDeleteName:       "DELETE_NAME",
	OpLoadFast:         "LOAD_FAST",
	OpStoreFast:        "STORE_FAST",
	OpDeleteFast:       "DELETE_FAST",
	OpLoadGlobal:       "LOAD_GLOBAL",
	OpStoreGlobal:      "STORE_GLOBAL",
	OpDeleteGlobal:     "DELETE_GLOBAL",
	OpLoadAttr:         "LOAD_ATTR",
	OpStoreAttr:        "STORE_ATTR",
	OpDeleteAttr:       "DELETE_ATTR",
	OpBinarySubscr:     "BINARY_SUBSCR",
	OpStoreSubscr:      "STORE_SUBSCR",
	OpDeleteSubscr:     "DELETE_SUBSCR",
	OpUnaryPositive:    "UNARY_POSITIVE",
	OpUnaryNegative:    "UNARY_NEGATIVE",
	OpUnaryNot:         "UNARY_NOT",
	OpUnaryInvert:      "UNARY_INVERT",
	OpBinaryAdd:        "BINARY_ADD",
	OpBinarySubtract:   "BINARY_SUBTRACT",
	OpBinaryMultiply:   "BINARY_MULTIPLY",
	OpBinaryDivide:     "BINARY_TRUE_DIVIDE",
	OpBinaryFloorDiv:   "BINARY_FLOOR_DIVIDE",
	OpBinaryModulo:     "BINARY_MODULO",
	OpBinaryPower:      "BINARY_POWER",
	OpBinaryMatMul:     "BINARY_MATRIX_MULTIPLY",
	OpBinaryLShift:     "BINARY_LSHIFT",
	OpBinaryRShift:     "BINARY_RSHIFT",
	OpBinaryAnd:        "BINARY_AND",
	OpBinaryOr:         "BINARY_OR",
	OpBinaryXor:        "BINARY_XOR",
	OpInplaceAdd:       "INPLACE_ADD",
	OpInplaceSubtract:  "INPLACE_SUBTRACT",
	OpInplaceMultiply:  "INPLACE_MULTIPLY",
	OpInplaceDivide:    "INPLACE_TRUE_DIVIDE",
	OpInplaceFloorDiv:  "INPLACE_FLOOR_DIVIDE",
	OpInplaceModulo:    "INPLACE_MODULO",
	OpInplacePower:     "INPLACE_POWER",
	OpInplaceMatMul:    "INPLACE_MATRIX_MULTIPLY",
	OpInplaceLShift:    "INPLACE_LSHIFT",
	OpInplaceRShift:    "INPLACE_RSHIFT",
	OpInplaceAnd:       "INPLACE_AND",
	OpInplaceOr:        "INPLACE_OR",
	OpInplaceXor:       "INPLACE_XOR",
	OpCompareEq:        "COMPARE_EQ",
	OpCompareNe:        "COMPARE_NE",
	OpCompareLt:        "COMPARE_LT",
	OpCompareLe:        "COMPARE_LE",
	OpCompareGt:        "COMPARE_GT",
	OpCompareGe:        "COMPARE_GE",
	OpCompareIs:        "COMPARE_IS",
	OpCompareIsNot:     "COMPARE_IS_NOT",
	OpCompareIn:        "COMPARE_IN",
	OpCompareNotIn:     "COMPARE_NOT_IN",
	OpJumpIfTrueOrPop:  "JUMP_IF_TRUE_OR_POP",
	OpJumpIfFalseOrPop: "JUMP_IF_FALSE_OR_POP",
	OpJump:             "JUMP",
	OpJumpIfTrue:       "JUMP_IF_TRUE",
	OpJumpIfFalse:      "JUMP_IF_FALSE",
	OpPopJumpIfTrue:    "POP_JUMP_IF_TRUE",
	OpPopJumpIfFalse:   "POP_JUMP_IF_FALSE",
	OpGetIter:          "GET_ITER",
	OpForIter:          "FOR_ITER",
	OpMakeFunction:     "MAKE_FUNCTION",
	OpCall:             "CALL",
	OpCallKw:           "CALL_KW",
	OpReturn:           "RETURN_VALUE",
	OpLoadBuildClass:   "LOAD_BUILD_CLASS",
	OpLoadMethod:       "LOAD_METHOD",
	OpCallMethod:       "CALL_METHOD",
	OpBuildTuple:       "BUILD_TUPLE",
	OpBuildList:        "BUILD_LIST",
	OpBuildSet:         "BUILD_SET",
	OpBuildMap:         "BUILD_MAP",
	OpBuildString:      "BUILD_STRING",
	OpUnpackSequence:   "UNPACK_SEQUENCE",
	OpUnpackEx:         "UNPACK_EX",
	OpListAppend:       "LIST_APPEND",
	OpSetAdd:           "SET_ADD",
	OpMapAdd:           "MAP_ADD",
	OpImportName:       "IMPORT_NAME",
	OpImportFrom:       "IMPORT_FROM",
	OpImportStar:       "IMPORT_STAR",
	OpSetupExcept:      "SETUP_EXCEPT",
	OpSetupFinally:     "SETUP_FINALLY",
	OpPopExcept:        "POP_EXCEPT",
	OpEndFinally:       "END_FINALLY",
	OpRaiseVarargs:     "RAISE_VARARGS",
	OpSetupWith:        "SETUP_WITH",
	OpWithCleanup:      "WITH_CLEANUP",
	OpAssert:           "ASSERT",
	OpLoadClosure:      "LOAD_CLOSURE",
	OpStoreClosure:     "STORE_CLOSURE",
	OpLoadDeref:        "LOAD_DEREF",
	OpStoreDeref:       "STORE_DEREF",
	OpMakeCell:         "MAKE_CELL",
	OpNop:              "NOP",
	OpPrintExpr:        "PRINT_EXPR",
	OpLoadLocals:       "LOAD_LOCALS",
	// Specialized opcodes
	OpLoadFast0:         "LOAD_FAST_0",
	OpLoadFast1:         "LOAD_FAST_1",
	OpLoadFast2:         "LOAD_FAST_2",
	OpLoadFast3:         "LOAD_FAST_3",
	OpStoreFast0:        "STORE_FAST_0",
	OpStoreFast1:        "STORE_FAST_1",
	OpStoreFast2:        "STORE_FAST_2",
	OpStoreFast3:        "STORE_FAST_3",
	OpLoadNone:          "LOAD_NONE",
	OpLoadTrue:          "LOAD_TRUE",
	OpLoadFalse:         "LOAD_FALSE",
	OpLoadZero:          "LOAD_ZERO",
	OpLoadOne:           "LOAD_ONE",
	OpIncrementFast:     "INCREMENT_FAST",
	OpDecrementFast:     "DECREMENT_FAST",
	OpLoadFastLoadFast:  "LOAD_FAST_LOAD_FAST",
	OpLoadFastLoadConst: "LOAD_FAST_LOAD_CONST",
	OpStoreFastLoadFast: "STORE_FAST_LOAD_FAST",
	OpBinaryAddInt:      "BINARY_ADD_INT",
	OpBinarySubtractInt: "BINARY_SUBTRACT_INT",
	OpBinaryMultiplyInt: "BINARY_MULTIPLY_INT",
	OpCompareLtInt:       "COMPARE_LT_INT",
	OpCompareLeInt:       "COMPARE_LE_INT",
	OpCompareGtInt:       "COMPARE_GT_INT",
	OpCompareGeInt:       "COMPARE_GE_INT",
	OpCompareEqInt:       "COMPARE_EQ_INT",
	OpCompareNeInt:       "COMPARE_NE_INT",
	OpLoadEmptyList:      "LOAD_EMPTY_LIST",
	OpLoadEmptyTuple:     "LOAD_EMPTY_TUPLE",
	OpLoadEmptyDict:      "LOAD_EMPTY_DICT",
	OpCompareLtJump:      "COMPARE_LT_JUMP",
	OpCompareLeJump:      "COMPARE_LE_JUMP",
	OpCompareGtJump:      "COMPARE_GT_JUMP",
	OpCompareGeJump:      "COMPARE_GE_JUMP",
	OpCompareEqJump:      "COMPARE_EQ_JUMP",
	OpCompareNeJump:      "COMPARE_NE_JUMP",
	OpLenList:            "LEN_LIST",
	OpLenString:          "LEN_STRING",
	OpLenTuple:           "LEN_TUPLE",
	OpLenDict:            "LEN_DICT",
	OpLenGeneric:         "LEN_GENERIC",
	OpLoadConstLoadFast:  "LOAD_CONST_LOAD_FAST",
	OpLoadGlobalLoadFast: "LOAD_GLOBAL_LOAD_FAST",
}

func (op Opcode) String() string {
	if name, ok := OpcodeNames[op]; ok {
		return name
	}
	return "UNKNOWN"
}

// hasArgTable is a lookup table for opcodes that take arguments
// Indexed by opcode value, true = has argument, false = no argument
var hasArgTable [256]bool

func init() {
	// By default, all opcodes have arguments (true)
	for i := range hasArgTable {
		hasArgTable[i] = true
	}
	// Mark opcodes that don't have arguments
	noArgOpcodes := []Opcode{
		OpPop, OpDup, OpRot2, OpRot3,
		OpUnaryPositive, OpUnaryNegative, OpUnaryNot, OpUnaryInvert,
		OpBinaryAdd, OpBinarySubtract, OpBinaryMultiply, OpBinaryDivide,
		OpBinaryFloorDiv, OpBinaryModulo, OpBinaryPower, OpBinaryMatMul,
		OpBinaryLShift, OpBinaryRShift, OpBinaryAnd, OpBinaryOr, OpBinaryXor,
		OpInplaceAdd, OpInplaceSubtract, OpInplaceMultiply, OpInplaceDivide,
		OpInplaceFloorDiv, OpInplaceModulo, OpInplacePower, OpInplaceMatMul,
		OpInplaceLShift, OpInplaceRShift, OpInplaceAnd, OpInplaceOr, OpInplaceXor,
		OpCompareEq, OpCompareNe, OpCompareLt, OpCompareLe,
		OpCompareGt, OpCompareGe, OpCompareIs, OpCompareIsNot,
		OpCompareIn, OpCompareNotIn,
		OpBinarySubscr, OpStoreSubscr, OpDeleteSubscr,
		OpGetIter, OpReturn,
		OpPopExcept, OpEndFinally, OpWithCleanup,
		OpNop, OpPrintExpr, OpLoadLocals, OpLoadBuildClass,
		OpImportStar,
		// Specialized no-arg opcodes
		OpLoadFast0, OpLoadFast1, OpLoadFast2, OpLoadFast3,
		OpStoreFast0, OpStoreFast1, OpStoreFast2, OpStoreFast3,
		OpLoadNone, OpLoadTrue, OpLoadFalse, OpLoadZero, OpLoadOne,
		OpBinaryAddInt, OpBinarySubtractInt, OpBinaryMultiplyInt,
		OpCompareLtInt, OpCompareLeInt, OpCompareGtInt, OpCompareGeInt, OpCompareEqInt, OpCompareNeInt,
		// Empty collection opcodes (no args)
		OpLoadEmptyList, OpLoadEmptyTuple, OpLoadEmptyDict,
		// Inline len opcodes (no args - operate on TOS)
		OpLenList, OpLenString, OpLenTuple, OpLenDict, OpLenGeneric,
	}
	for _, op := range noArgOpcodes {
		hasArgTable[op] = false
	}
}

// HasArg returns true if the opcode takes an argument
// Uses a lookup table for O(1) performance
func (op Opcode) HasArg() bool {
	return hasArgTable[op]
}

// Instruction represents a single bytecode instruction with its argument
type Instruction struct {
	Op     Opcode
	Arg    int // Argument value (if HasArg)
	Line   int // Source line number
	Offset int // Byte offset in code
}

// CodeObject represents compiled Python code
type CodeObject struct {
	Name           string        // Function/class/module name
	Filename       string        // Source filename
	FirstLine      int           // First line number in source
	Code           []byte        // Bytecode instructions
	Constants      []interface{} // Constant pool
	Names          []string      // Names used in code
	VarNames       []string      // Local variable names
	FreeVars       []string      // Free variables (closures)
	CellVars       []string      // Variables captured in closures
	ArgCount       int           // Number of positional arguments
	KwOnlyArgCount int           // Number of keyword-only arguments
	Flags          CodeFlags     // Code flags
	StackSize      int           // Maximum stack size needed
	LineNoTab      []LineEntry   // Line number table
}

// CodeFlags represents flags for code objects
type CodeFlags int

const (
	FlagOptimized         CodeFlags = 1 << iota // Locals are optimized
	FlagNewLocals                               // Use new dict for locals
	FlagVarArgs                                 // *args parameter
	FlagVarKeywords                             // **kwargs parameter
	FlagNested                                  // Nested function
	FlagGenerator                               // Generator function
	FlagNoFree                                  // No free variables
	FlagCoroutine                               // Coroutine function
	FlagIterableCoroutine                       // Iterable coroutine
	FlagAsyncGenerator                          // Async generator
)

// LineEntry maps bytecode offsets to source lines
type LineEntry struct {
	StartOffset int
	EndOffset   int
	Line        int
}

// Disassemble returns a human-readable disassembly of the code object
func (co *CodeObject) Disassemble() string {
	var result string
	result += "Disassembly of " + co.Name + ":\n"

	offset := 0
	for offset < len(co.Code) {
		op := Opcode(co.Code[offset])
		line := co.LineForOffset(offset)

		if op.HasArg() && offset+2 < len(co.Code) {
			arg := int(co.Code[offset+1]) | int(co.Code[offset+2])<<8
			argStr := formatArg(co, op, arg)
			result += formatInstruction(offset, line, op, arg, argStr)
			offset += 3
		} else {
			result += formatInstruction(offset, line, op, -1, "")
			offset += 1
		}
	}

	return result
}

// LineForOffset returns the source line for a bytecode offset
func (co *CodeObject) LineForOffset(offset int) int {
	for _, entry := range co.LineNoTab {
		if offset >= entry.StartOffset && offset < entry.EndOffset {
			return entry.Line
		}
	}
	return co.FirstLine
}

func formatInstruction(offset, line int, op Opcode, arg int, argStr string) string {
	if arg >= 0 {
		if argStr != "" {
			return fmt.Sprintf("%4d %4d %-20s %d (%s)\n", line, offset, op.String(), arg, argStr)
		}
		return fmt.Sprintf("%4d %4d %-20s %d\n", line, offset, op.String(), arg)
	}
	return fmt.Sprintf("%4d %4d %-20s\n", line, offset, op.String())
}

func formatArg(co *CodeObject, op Opcode, arg int) string {
	switch op {
	case OpLoadConst:
		if arg < len(co.Constants) {
			return fmt.Sprintf("%v", co.Constants[arg])
		}
	case OpLoadName, OpStoreName, OpDeleteName,
		OpLoadGlobal, OpStoreGlobal, OpDeleteGlobal,
		OpLoadAttr, OpStoreAttr, OpDeleteAttr,
		OpImportName, OpImportFrom, OpLoadMethod:
		if arg < len(co.Names) {
			return co.Names[arg]
		}
	case OpLoadFast, OpStoreFast, OpDeleteFast:
		if arg < len(co.VarNames) {
			return co.VarNames[arg]
		}
	case OpLoadDeref, OpStoreDeref, OpLoadClosure, OpStoreClosure:
		idx := arg
		if idx < len(co.CellVars) {
			return co.CellVars[idx]
		}
		idx -= len(co.CellVars)
		if idx < len(co.FreeVars) {
			return co.FreeVars[idx]
		}
	case OpJump, OpJumpIfTrue, OpJumpIfFalse,
		OpPopJumpIfTrue, OpPopJumpIfFalse,
		OpJumpIfTrueOrPop, OpJumpIfFalseOrPop,
		OpForIter, OpSetupExcept, OpSetupFinally, OpSetupWith:
		return fmt.Sprintf("to %d", arg)
	}
	return ""
}
