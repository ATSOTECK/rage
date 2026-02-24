package compiler

import (
	"github.com/ATSOTECK/rage/internal/runtime"
)

func (o *Optimizer) removeLoadPop(instrs []*instruction) bool {
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}
		// Pattern: LOAD_FAST/LOAD_CONST followed by POP
		if (instrs[i].op == runtime.OpLoadFast || instrs[i].op == runtime.OpLoadConst ||
			instrs[i].op == runtime.OpLoadFast0 || instrs[i].op == runtime.OpLoadFast1 ||
			instrs[i].op == runtime.OpLoadFast2 || instrs[i].op == runtime.OpLoadFast3 ||
			instrs[i].op == runtime.OpLoadNone || instrs[i].op == runtime.OpLoadTrue ||
			instrs[i].op == runtime.OpLoadFalse) &&
			instrs[i+1].op == runtime.OpPop {
			instrs[i].removed = true
			instrs[i+1].removed = true
			changed = true
		}
	}
	return changed
}

func (o *Optimizer) removeDupPop(instrs []*instruction) bool {
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}
		if instrs[i].op == runtime.OpDup && instrs[i+1].op == runtime.OpPop {
			instrs[i].removed = true
			instrs[i+1].removed = true
			changed = true
		}
	}
	return changed
}

func (o *Optimizer) optimizeJumps(instrs []*instruction, code *runtime.CodeObject) bool {
	// This is complex because we need to handle jump targets
	// For now, just handle simple cases

	// First, build a set of instruction indices that are jump targets
	// We need to be careful not to optimize patterns where the first instruction
	// is a jump target (because control flow may come from elsewhere with different stack)
	jumpTargets := make(map[int]bool)
	offset := 0
	for i, instr := range instrs {
		if instr.originalHadArg {
			offset += 3
		} else {
			offset++
		}
		_ = i // We'll use the offset to find target instruction index below
	}

	// Map offsets to instruction indices
	offsetToIndex := make(map[int]int)
	offset = 0
	for i, instr := range instrs {
		offsetToIndex[offset] = i
		if instr.originalHadArg {
			offset += 3
		} else {
			offset++
		}
	}

	// Mark jump targets
	for _, instr := range instrs {
		if isJumpOp(instr.op) && instr.arg >= 0 {
			if targetIdx, ok := offsetToIndex[instr.arg]; ok {
				jumpTargets[targetIdx] = true
			}
		}
	}

	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed {
			continue
		}

		// Skip if this instruction is a jump target - can't safely optimize
		// because control flow may come from elsewhere with different stack state
		if jumpTargets[i] {
			continue
		}

		// LOAD_CONST True; POP_JUMP_IF_FALSE -> remove both (never jumps)
		if instrs[i].op == runtime.OpLoadTrue || instrs[i].op == runtime.OpLoadConst {
			if instrs[i].op == runtime.OpLoadConst {
				// Check if it's loading True
				arg := instrs[i].arg
				if arg >= 0 && arg < len(code.Constants) {
					if b, ok := code.Constants[arg].(bool); !ok || !b {
						continue
					}
				}
			}
			if i+1 < len(instrs) && instrs[i+1].op == runtime.OpPopJumpIfFalse {
				instrs[i].removed = true
				instrs[i+1].removed = true
				changed = true
			}
		}
		// LOAD_CONST False; POP_JUMP_IF_FALSE -> JUMP
		if instrs[i].op == runtime.OpLoadFalse || instrs[i].op == runtime.OpLoadConst {
			if instrs[i].op == runtime.OpLoadConst {
				arg := instrs[i].arg
				if arg >= 0 && arg < len(code.Constants) {
					if b, ok := code.Constants[arg].(bool); !ok || b {
						continue
					}
				}
			}
			if i+1 < len(instrs) && instrs[i+1].op == runtime.OpPopJumpIfFalse {
				instrs[i].removed = true
				instrs[i+1].op = runtime.OpJump // Convert to unconditional jump
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) specializeLoadFast(instrs []*instruction) bool {
	changed := false
	for _, instr := range instrs {
		if instr.removed {
			continue
		}
		if instr.op == runtime.OpLoadFast {
			switch instr.arg {
			case 0:
				instr.op = runtime.OpLoadFast0
				instr.arg = -1
				changed = true
			case 1:
				instr.op = runtime.OpLoadFast1
				instr.arg = -1
				changed = true
			case 2:
				instr.op = runtime.OpLoadFast2
				instr.arg = -1
				changed = true
			case 3:
				instr.op = runtime.OpLoadFast3
				instr.arg = -1
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) specializeStoreFast(instrs []*instruction) bool {
	changed := false
	for _, instr := range instrs {
		if instr.removed {
			continue
		}
		if instr.op == runtime.OpStoreFast {
			switch instr.arg {
			case 0:
				instr.op = runtime.OpStoreFast0
				instr.arg = -1
				changed = true
			case 1:
				instr.op = runtime.OpStoreFast1
				instr.arg = -1
				changed = true
			case 2:
				instr.op = runtime.OpStoreFast2
				instr.arg = -1
				changed = true
			case 3:
				instr.op = runtime.OpStoreFast3
				instr.arg = -1
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) specializeLoadConst(instrs []*instruction, code *runtime.CodeObject) bool {
	changed := false
	for _, instr := range instrs {
		if instr.removed || instr.op != runtime.OpLoadConst {
			continue
		}
		if instr.arg >= len(code.Constants) {
			continue
		}
		c := code.Constants[instr.arg]
		switch v := c.(type) {
		case nil:
			instr.op = runtime.OpLoadNone
			instr.arg = -1
			changed = true
		case bool:
			if v {
				instr.op = runtime.OpLoadTrue
			} else {
				instr.op = runtime.OpLoadFalse
			}
			instr.arg = -1
			changed = true
		case int64:
			if v == 0 {
				instr.op = runtime.OpLoadZero
				instr.arg = -1
				changed = true
			} else if v == 1 {
				instr.op = runtime.OpLoadOne
				instr.arg = -1
				changed = true
			}
		case int:
			if v == 0 {
				instr.op = runtime.OpLoadZero
				instr.arg = -1
				changed = true
			} else if v == 1 {
				instr.op = runtime.OpLoadOne
				instr.arg = -1
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) detectIncrementPattern(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_CONST 1, BINARY_ADD, STORE_FAST x -> INCREMENT_FAST x
	changed := false
	for i := 0; i < len(instrs)-3; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed || instrs[i+3].removed {
			continue
		}

		// Check for LOAD_FAST or specialized version
		var localIdx int
		switch instrs[i].op {
		case runtime.OpLoadFast:
			localIdx = instrs[i].arg
		case runtime.OpLoadFast0:
			localIdx = 0
		case runtime.OpLoadFast1:
			localIdx = 1
		case runtime.OpLoadFast2:
			localIdx = 2
		case runtime.OpLoadFast3:
			localIdx = 3
		default:
			continue
		}

		// Check for LOAD_CONST 1 or LOAD_ONE
		isOne := false
		if instrs[i+1].op == runtime.OpLoadOne {
			isOne = true
		} else if instrs[i+1].op == runtime.OpLoadConst {
			arg := instrs[i+1].arg
			if arg >= 0 && arg < len(code.Constants) {
				switch v := code.Constants[arg].(type) {
				case int64:
					isOne = v == 1
				case int:
					isOne = v == 1
				}
			}
		}
		if !isOne {
			continue
		}

		// Check for BINARY_ADD
		if instrs[i+2].op != runtime.OpBinaryAdd && instrs[i+2].op != runtime.OpBinaryAddInt {
			continue
		}

		// Check for STORE_FAST to same variable
		var storeIdx int
		switch instrs[i+3].op {
		case runtime.OpStoreFast:
			storeIdx = instrs[i+3].arg
		case runtime.OpStoreFast0:
			storeIdx = 0
		case runtime.OpStoreFast1:
			storeIdx = 1
		case runtime.OpStoreFast2:
			storeIdx = 2
		case runtime.OpStoreFast3:
			storeIdx = 3
		default:
			continue
		}

		if localIdx != storeIdx {
			continue
		}

		// Convert to INCREMENT_FAST
		instrs[i].op = runtime.OpIncrementFast
		instrs[i].arg = localIdx
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		instrs[i+3].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectDecrementPattern(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_CONST 1, BINARY_SUB, STORE_FAST x -> DECREMENT_FAST x
	changed := false
	for i := 0; i < len(instrs)-3; i++ {
		if instrs[i].removed {
			continue
		}

		// Check for LOAD_FAST or specialized version
		localIdx := o.getLoadFastIndex(instrs[i])
		if localIdx < 0 {
			continue
		}

		// Check for LOAD_CONST 1 or LOAD_ONE
		if !o.isLoadOne(instrs[i+1], code) {
			continue
		}

		// Check for BINARY_SUBTRACT
		if instrs[i+2].op != runtime.OpBinarySubtract && instrs[i+2].op != runtime.OpBinarySubtractInt {
			continue
		}

		// Check for STORE_FAST to same variable
		storeIdx := o.getStoreFastIndex(instrs[i+3])
		if storeIdx < 0 || localIdx != storeIdx {
			continue
		}

		// Convert to DECREMENT_FAST
		instrs[i].op = runtime.OpDecrementFast
		instrs[i].arg = localIdx
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		instrs[i+3].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectNegatePattern(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, UNARY_NEGATIVE, STORE_FAST x -> NEGATE_FAST x
	changed := false
	for i := 0; i < len(instrs)-2; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed {
			continue
		}

		// Check for LOAD_FAST or specialized version
		localIdx := o.getLoadFastIndex(instrs[i])
		if localIdx < 0 {
			continue
		}

		// Check for UNARY_NEGATIVE
		if instrs[i+1].op != runtime.OpUnaryNegative {
			continue
		}

		// Check for STORE_FAST to same variable
		storeIdx := o.getStoreFastIndex(instrs[i+2])
		if storeIdx < 0 || localIdx != storeIdx {
			continue
		}

		// Convert to NEGATE_FAST
		instrs[i].op = runtime.OpNegateFast
		instrs[i].arg = localIdx
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectAddConstPattern(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_CONST c, BINARY_ADD, STORE_FAST x -> ADD_CONST_FAST (x, c)
	// Skip if c == 1 (handled by INCREMENT_FAST)
	changed := false
	for i := 0; i < len(instrs)-3; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed || instrs[i+3].removed {
			continue
		}

		// Check for LOAD_FAST or specialized version
		localIdx := o.getLoadFastIndex(instrs[i])
		if localIdx < 0 {
			continue
		}

		// Check for LOAD_CONST (but not 1, which is handled by INCREMENT_FAST)
		if instrs[i+1].op != runtime.OpLoadConst {
			continue
		}
		constIdx := instrs[i+1].arg
		// Skip if it's loading 1 (use INCREMENT_FAST instead)
		if o.isLoadOne(instrs[i+1], code) {
			continue
		}
		// Only optimize for integer constants
		if constIdx >= len(code.Constants) {
			continue
		}
		switch code.Constants[constIdx].(type) {
		case int64, int:
			// OK, it's an integer
		default:
			continue
		}

		// Check for BINARY_ADD
		if instrs[i+2].op != runtime.OpBinaryAdd && instrs[i+2].op != runtime.OpBinaryAddInt {
			continue
		}

		// Check for STORE_FAST to same variable
		storeIdx := o.getStoreFastIndex(instrs[i+3])
		if storeIdx < 0 || localIdx != storeIdx {
			continue
		}

		// Only encode if indices fit in packed format (8 bits each)
		if localIdx > 255 || constIdx > 255 {
			continue
		}

		// Convert to ADD_CONST_FAST
		instrs[i].op = runtime.OpAddConstFast
		instrs[i].arg = localIdx | (constIdx << 8)
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		instrs[i+3].removed = true
		changed = true
	}
	return changed
}

// Helper to get LOAD_FAST index from instruction (handles specialized versions)
func (o *Optimizer) getLoadFastIndex(instr *instruction) int {
	if instr.removed {
		return -1
	}
	switch instr.op {
	case runtime.OpLoadFast:
		return instr.arg
	case runtime.OpLoadFast0:
		return 0
	case runtime.OpLoadFast1:
		return 1
	case runtime.OpLoadFast2:
		return 2
	case runtime.OpLoadFast3:
		return 3
	}
	return -1
}

// computeJumpTargets returns a map of instruction indices that are jump targets.
// This is used to prevent unsafe optimizations that would merge instructions
// where control flow can jump into the middle of the merged instruction.
func (o *Optimizer) computeJumpTargets(instrs []*instruction) map[int]bool {
	// Map offsets to instruction indices
	offsetToIndex := make(map[int]int)
	offset := 0
	for i, instr := range instrs {
		offsetToIndex[offset] = i
		if instr.originalHadArg {
			offset += 3
		} else {
			offset++
		}
	}

	// Mark jump targets
	jumpTargets := make(map[int]bool)
	for _, instr := range instrs {
		if isJumpOp(instr.op) && instr.arg >= 0 {
			if targetIdx, ok := offsetToIndex[instr.arg]; ok {
				jumpTargets[targetIdx] = true
			}
		}
	}
	return jumpTargets
}

// Helper to get STORE_FAST index from instruction (handles specialized versions)
func (o *Optimizer) getStoreFastIndex(instr *instruction) int {
	if instr.removed {
		return -1
	}
	switch instr.op {
	case runtime.OpStoreFast:
		return instr.arg
	case runtime.OpStoreFast0:
		return 0
	case runtime.OpStoreFast1:
		return 1
	case runtime.OpStoreFast2:
		return 2
	case runtime.OpStoreFast3:
		return 3
	}
	return -1
}

// Helper to check if instruction loads the constant 1
func (o *Optimizer) isLoadOne(instr *instruction, code *runtime.CodeObject) bool {
	if instr.removed {
		return false
	}
	if instr.op == runtime.OpLoadOne {
		return true
	}
	if instr.op == runtime.OpLoadConst && instr.arg >= 0 && instr.arg < len(code.Constants) {
		switch v := code.Constants[instr.arg].(type) {
		case int64:
			return v == 1
		case int:
			return v == 1
		}
	}
	return false
}

func (o *Optimizer) detectLoadFastLoadFast(instrs []*instruction) bool {
	// Pattern: LOAD_FAST x, LOAD_FAST y -> LOAD_FAST_LOAD_FAST (packed)
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		idx1 := o.getLoadFastIndex(instrs[i])
		idx2 := o.getLoadFastIndex(instrs[i+1])

		if idx1 < 0 || idx2 < 0 {
			continue
		}

		// Both indices must fit in 8 bits each (we pack into 16-bit arg)
		if idx1 > 255 || idx2 > 255 {
			continue
		}

		// Convert to superinstruction
		// VM expects: low byte = first, high byte = second
		instrs[i].op = runtime.OpLoadFastLoadFast
		instrs[i].arg = (idx2 << 8) | idx1
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectLoadFastLoadConst(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_CONST y -> LOAD_FAST_LOAD_CONST (packed)
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		fastIdx := o.getLoadFastIndex(instrs[i])
		if fastIdx < 0 || fastIdx > 255 {
			continue
		}

		// Check for LOAD_CONST (but not specialized versions - those are already optimal)
		if instrs[i+1].op != runtime.OpLoadConst {
			continue
		}
		constIdx := instrs[i+1].arg
		if constIdx > 255 {
			continue
		}

		// Convert to superinstruction
		// VM expects: low byte = local, high byte = const
		instrs[i].op = runtime.OpLoadFastLoadConst
		instrs[i].arg = (constIdx << 8) | fastIdx
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectLoadConstLoadFast(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_CONST x, LOAD_FAST y -> LOAD_CONST_LOAD_FAST (packed)
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		// Check for LOAD_CONST (but not specialized versions)
		if instrs[i].op != runtime.OpLoadConst {
			continue
		}
		constIdx := instrs[i].arg
		if constIdx > 255 {
			continue
		}

		fastIdx := o.getLoadFastIndex(instrs[i+1])
		if fastIdx < 0 || fastIdx > 255 {
			continue
		}

		// Convert to superinstruction
		// VM expects: high byte = const, low byte = local
		instrs[i].op = runtime.OpLoadConstLoadFast
		instrs[i].arg = (constIdx << 8) | fastIdx
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) detectStoreFastLoadFast(instrs []*instruction) bool {
	// Pattern: STORE_FAST x, LOAD_FAST y -> STORE_FAST_LOAD_FAST (packed)
	// This is common for chained assignments and expression statements
	// IMPORTANT: Don't merge if LOAD_FAST is a jump target (e.g., start of a while loop)

	// Build map of instruction indices that are jump targets
	jumpTargets := o.computeJumpTargets(instrs)

	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		// Don't merge if the LOAD_FAST is a jump target - control flow may
		// jump directly to it, bypassing the STORE_FAST
		if jumpTargets[i+1] {
			continue
		}

		storeIdx := o.getStoreFastIndex(instrs[i])
		if storeIdx < 0 || storeIdx > 255 {
			continue
		}

		loadIdx := o.getLoadFastIndex(instrs[i+1])
		if loadIdx < 0 || loadIdx > 255 {
			continue
		}

		// Convert to superinstruction
		// VM expects: low byte = store, high byte = load
		instrs[i].op = runtime.OpStoreFastLoadFast
		instrs[i].arg = (loadIdx << 8) | storeIdx
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) optimizeEmptyCollections(instrs []*instruction) bool {
	// Pattern: BUILD_LIST 0 -> LOAD_EMPTY_LIST
	// Pattern: BUILD_TUPLE 0 -> LOAD_EMPTY_TUPLE
	// Pattern: BUILD_MAP 0 -> LOAD_EMPTY_DICT
	changed := false
	for _, instr := range instrs {
		if instr.removed {
			continue
		}
		switch instr.op {
		case runtime.OpBuildList:
			if instr.arg == 0 {
				instr.op = runtime.OpLoadEmptyList
				instr.arg = -1
				changed = true
			}
		case runtime.OpBuildTuple:
			if instr.arg == 0 {
				instr.op = runtime.OpLoadEmptyTuple
				instr.arg = -1
				changed = true
			}
		case runtime.OpBuildMap:
			if instr.arg == 0 {
				instr.op = runtime.OpLoadEmptyDict
				instr.arg = -1
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) optimizeCompareJump(instrs []*instruction) bool {
	// Pattern: COMPARE_xx, POP_JUMP_IF_FALSE -> COMPARE_xx_JUMP
	// IMPORTANT: Don't fuse if the POP_JUMP_IF_FALSE is a jump target,
	// because other instructions (e.g. JUMP_IF_FALSE_OR_POP from 'and'/'or')
	// may jump to it expecting it to pop a value from the stack.
	jumpTargets := o.computeJumpTargets(instrs)

	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		// Check if this is a compare followed by POP_JUMP_IF_FALSE
		if instrs[i+1].op != runtime.OpPopJumpIfFalse {
			continue
		}

		// Don't fuse if the POP_JUMP_IF_FALSE is a jump target
		if jumpTargets[i+1] {
			continue
		}

		var newOp runtime.Opcode
		switch instrs[i].op {
		case runtime.OpCompareLt:
			newOp = runtime.OpCompareLtJump
		case runtime.OpCompareLe:
			newOp = runtime.OpCompareLeJump
		case runtime.OpCompareGt:
			newOp = runtime.OpCompareGtJump
		case runtime.OpCompareGe:
			newOp = runtime.OpCompareGeJump
		case runtime.OpCompareEq:
			newOp = runtime.OpCompareEqJump
		case runtime.OpCompareNe:
			newOp = runtime.OpCompareNeJump
		default:
			continue
		}

		// Convert to combined compare+jump
		instrs[i].op = newOp
		instrs[i].arg = instrs[i+1].arg // Take the jump target
		instrs[i+1].removed = true
		changed = true
	}
	return changed
}

func (o *Optimizer) eliminateStoreLoad(instrs []*instruction) bool {
	// Pattern: STORE_FAST x, LOAD_FAST x -> DUP, STORE_FAST x
	// This keeps the value on the stack and avoids the re-load
	changed := false
	for i := 0; i < len(instrs)-1; i++ {
		if instrs[i].removed || instrs[i+1].removed {
			continue
		}

		storeIdx := o.getStoreFastIndex(instrs[i])
		if storeIdx < 0 {
			continue
		}

		loadIdx := o.getLoadFastIndex(instrs[i+1])
		if loadIdx < 0 {
			continue
		}

		// Only optimize if storing and loading the same variable
		if storeIdx != loadIdx {
			continue
		}

		// Convert to DUP + STORE
		// Insert DUP before STORE, remove LOAD
		instrs[i+1].op = instrs[i].op   // Move store to second position
		instrs[i+1].arg = instrs[i].arg // Copy the argument
		instrs[i].op = runtime.OpDup
		instrs[i].arg = -1
		changed = true
	}
	return changed
}

func (o *Optimizer) threadJumps(instrs []*instruction) bool {
	// Build offset map for finding targets
	// Use originalHadArg since jump args still reference original byte offsets
	offsetToIdx := make(map[int]int)
	offset := 0
	for i, instr := range instrs {
		offsetToIdx[offset] = i
		if instr.originalHadArg {
			offset += 3
		} else {
			offset++
		}
	}

	changed := false
	for _, instr := range instrs {
		if instr.removed {
			continue
		}

		// Only process unconditional jumps
		if instr.op != runtime.OpJump {
			continue
		}

		// Find the target instruction
		targetIdx, ok := offsetToIdx[instr.arg]
		if !ok {
			continue
		}

		// If target is also a jump, thread through it
		if targetIdx < len(instrs) && !instrs[targetIdx].removed {
			target := instrs[targetIdx]
			if target.op == runtime.OpJump {
				// Thread the jump
				instr.arg = target.arg
				changed = true
			}
		}
	}
	return changed
}

func (o *Optimizer) optimizeLenCalls(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_GLOBAL "len", LOAD_xxx, CALL 1 -> LEN_GENERIC
	// This avoids the function call overhead for len()
	changed := false
	for i := 0; i < len(instrs)-2; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed {
			continue
		}

		// Check for LOAD_GLOBAL with "len"
		if instrs[i].op != runtime.OpLoadGlobal {
			continue
		}

		// Verify it's loading "len"
		nameIdx := instrs[i].arg
		if nameIdx >= len(code.Names) || code.Names[nameIdx] != "len" {
			continue
		}

		// Verify i+1 is a simple load (the argument to len), not a CALL or other complex op
		switch instrs[i+1].op {
		case runtime.OpLoadConst, runtime.OpLoadName, runtime.OpLoadFast,
			runtime.OpLoadGlobal, runtime.OpLoadDeref, runtime.OpLoadClosure,
			runtime.OpLoadFast0, runtime.OpLoadFast1, runtime.OpLoadFast2, runtime.OpLoadFast3,
			runtime.OpLoadNone, runtime.OpLoadTrue, runtime.OpLoadFalse,
			runtime.OpLoadZero, runtime.OpLoadOne, runtime.OpLoadAttr:
			// These are safe single-value loads
		default:
			continue
		}

		// Check for CALL with 1 argument
		if instrs[i+2].op != runtime.OpCall || instrs[i+2].arg != 1 {
			continue
		}

		// Replace with inline len
		// Remove LOAD_GLOBAL len
		instrs[i].removed = true

		// Replace CALL with LEN_GENERIC (operates on whatever is on stack)
		instrs[i+2].op = runtime.OpLenGeneric
		instrs[i+2].arg = -1
		changed = true
	}
	return changed
}

// ==========================================
// Binary Operation Specialization
// ==========================================

// Helper to check if an instruction loads an integer constant
func (o *Optimizer) isIntegerLoad(instr *instruction, code *runtime.CodeObject) bool {
	if instr.removed {
		return false
	}
	switch instr.op {
	case runtime.OpLoadZero, runtime.OpLoadOne:
		return true
	case runtime.OpLoadConst:
		if instr.arg >= 0 && instr.arg < len(code.Constants) {
			switch code.Constants[instr.arg].(type) {
			case int64, int:
				return true
			}
		}
	}
	return false
}

// SpecializeBinaryOps converts generic binary ops to specialized int versions
// when we can prove or heuristically determine operands are integers
func (o *Optimizer) SpecializeBinaryOps(instrs []*instruction, code *runtime.CodeObject) bool {
	changed := false
	for i := 0; i < len(instrs); i++ {
		if instrs[i].removed {
			continue
		}

		// Look for patterns where we load two integer constants and operate
		// Pattern: LOAD_CONST int, LOAD_CONST int, BINARY_OP
		if i >= 2 {
			op1 := instrs[i-2]
			op2 := instrs[i-1]
			if o.isIntegerLoad(op1, code) && o.isIntegerLoad(op2, code) {
				switch instrs[i].op {
				case runtime.OpBinaryAdd:
					instrs[i].op = runtime.OpBinaryAddInt
					changed = true
				case runtime.OpBinarySubtract:
					instrs[i].op = runtime.OpBinarySubtractInt
					changed = true
				case runtime.OpBinaryMultiply:
					instrs[i].op = runtime.OpBinaryMultiplyInt
					changed = true
				case runtime.OpCompareLt:
					instrs[i].op = runtime.OpCompareLtInt
					changed = true
				case runtime.OpCompareLe:
					instrs[i].op = runtime.OpCompareLeInt
					changed = true
				case runtime.OpCompareGt:
					instrs[i].op = runtime.OpCompareGtInt
					changed = true
				case runtime.OpCompareGe:
					instrs[i].op = runtime.OpCompareGeInt
					changed = true
				case runtime.OpCompareEq:
					instrs[i].op = runtime.OpCompareEqInt
					changed = true
				case runtime.OpCompareNe:
					instrs[i].op = runtime.OpCompareNeInt
					changed = true
				}
			}
		}

		// Pattern: LOAD_FAST, LOAD_CONST int, BINARY_OP (common in loops)
		// After FOR_ITER, the loop variable is often an int (from range)
		if i >= 2 && o.getLoadFastIndex(instrs[i-2]) >= 0 && o.isIntegerLoad(instrs[i-1], code) {
			switch instrs[i].op {
			case runtime.OpCompareLt:
				instrs[i].op = runtime.OpCompareLtInt
				changed = true
			case runtime.OpCompareLe:
				instrs[i].op = runtime.OpCompareLeInt
				changed = true
			case runtime.OpCompareGt:
				instrs[i].op = runtime.OpCompareGtInt
				changed = true
			case runtime.OpCompareGe:
				instrs[i].op = runtime.OpCompareGeInt
				changed = true
			case runtime.OpCompareEq:
				instrs[i].op = runtime.OpCompareEqInt
				changed = true
			case runtime.OpCompareNe:
				instrs[i].op = runtime.OpCompareNeInt
				changed = true
			case runtime.OpBinaryAdd:
				instrs[i].op = runtime.OpBinaryAddInt
				changed = true
			case runtime.OpBinarySubtract:
				instrs[i].op = runtime.OpBinarySubtractInt
				changed = true
			case runtime.OpBinaryMultiply:
				instrs[i].op = runtime.OpBinaryMultiplyInt
				changed = true
			}
		}

		// Pattern: LOAD_CONST int, LOAD_FAST, BINARY_OP
		if i >= 2 && o.isIntegerLoad(instrs[i-2], code) && o.getLoadFastIndex(instrs[i-1]) >= 0 {
			switch instrs[i].op {
			case runtime.OpCompareLt:
				instrs[i].op = runtime.OpCompareLtInt
				changed = true
			case runtime.OpCompareLe:
				instrs[i].op = runtime.OpCompareLeInt
				changed = true
			case runtime.OpCompareGt:
				instrs[i].op = runtime.OpCompareGtInt
				changed = true
			case runtime.OpCompareGe:
				instrs[i].op = runtime.OpCompareGeInt
				changed = true
			case runtime.OpCompareEq:
				instrs[i].op = runtime.OpCompareEqInt
				changed = true
			case runtime.OpCompareNe:
				instrs[i].op = runtime.OpCompareNeInt
				changed = true
			case runtime.OpBinaryAdd:
				instrs[i].op = runtime.OpBinaryAddInt
				changed = true
			case runtime.OpBinarySubtract:
				instrs[i].op = runtime.OpBinarySubtractInt
				changed = true
			case runtime.OpBinaryMultiply:
				instrs[i].op = runtime.OpBinaryMultiplyInt
				changed = true
			}
		}

		// Always specialize BINARY_DIVIDE to BINARY_DIVIDE_FLOAT
		// (true division in Python always returns float)
		if instrs[i].op == runtime.OpBinaryDivide {
			instrs[i].op = runtime.OpBinaryDivideFloat
			changed = true
		}

		// Specialize float addition when we know one operand is float
		// Pattern: ... BINARY_ADD after float operations
		if i >= 1 && instrs[i].op == runtime.OpBinaryAdd {
			// Check if previous result was from a division (which always produces float)
			for j := i - 1; j >= 0; j-- {
				if instrs[j].removed {
					continue
				}
				if instrs[j].op == runtime.OpBinaryDivideFloat || instrs[j].op == runtime.OpBinaryDivide {
					instrs[i].op = runtime.OpBinaryAddFloat
					changed = true
				}
				break
			}
		}
	}
	return changed
}

func (o *Optimizer) detectCompareLtLocalJump(instrs []*instruction, code *runtime.CodeObject) bool {
	// Pattern: LOAD_FAST x, LOAD_FAST y, COMPARE_LT, POP_JUMP_IF_FALSE target
	// -> COMPARE_LT_LOCAL_JUMP (x, y, target)
	// IMPORTANT: Don't fuse if any inner instruction is a jump target.
	jumpTargets := o.computeJumpTargets(instrs)

	changed := false
	for i := 0; i < len(instrs)-3; i++ {
		if instrs[i].removed || instrs[i+1].removed || instrs[i+2].removed || instrs[i+3].removed {
			continue
		}

		// Don't fuse if any of the inner instructions are jump targets
		if jumpTargets[i+1] || jumpTargets[i+2] || jumpTargets[i+3] {
			continue
		}

		// Get first local index
		local1 := o.getLoadFastIndex(instrs[i])
		if local1 < 0 || local1 > 255 {
			continue
		}

		// Get second local index
		local2 := o.getLoadFastIndex(instrs[i+1])
		if local2 < 0 || local2 > 255 {
			continue
		}

		// Check for COMPARE_LT
		if instrs[i+2].op != runtime.OpCompareLt && instrs[i+2].op != runtime.OpCompareLtInt {
			continue
		}

		// Check for POP_JUMP_IF_FALSE
		if instrs[i+3].op != runtime.OpPopJumpIfFalse {
			continue
		}

		jumpTarget := instrs[i+3].arg
		// Jump target must fit in remaining bits (16 bits after two 8-bit indices)
		if jumpTarget > 0xFFFF {
			continue
		}

		// Convert to COMPARE_LT_LOCAL_JUMP
		// Pack: bits 0-7 = local1, bits 8-15 = local2, bits 16-31 = jump target
		instrs[i].op = runtime.OpCompareLtLocalJump
		instrs[i].arg = local1 | (local2 << 8) | (jumpTarget << 16)
		instrs[i+1].removed = true
		instrs[i+2].removed = true
		instrs[i+3].removed = true
		changed = true
	}
	return changed
}
