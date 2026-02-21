package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

func (vm *VM) initBuiltins() {
	vm.initBuiltinsTypes()
	vm.initBuiltinsFunctions()
	vm.initBuiltinsClasses()
	vm.initExceptionClasses()
}

// parseComplexString parses a string like "1+2j", "3j", "-1-2j", "1", etc. into a PyComplex
func parseComplexString(s string) (*PyComplex, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
	}

	// Remove surrounding parens if present
	if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		s = s[1 : len(s)-1]
		s = strings.TrimSpace(s)
	}

	// Pure imaginary: ends with 'j' or 'J'
	if s[len(s)-1] == 'j' || s[len(s)-1] == 'J' {
		body := s[:len(s)-1]
		if body == "" || body == "+" {
			return MakeComplex(0, 1), nil
		}
		if body == "-" {
			return MakeComplex(0, -1), nil
		}

		// Find the split point between real and imaginary parts
		// Look for + or - that is NOT after e/E (scientific notation)
		splitIdx := -1
		for i := len(body) - 1; i > 0; i-- {
			if (body[i] == '+' || body[i] == '-') && body[i-1] != 'e' && body[i-1] != 'E' {
				splitIdx = i
				break
			}
		}

		if splitIdx == -1 {
			// Pure imaginary like "3j" or "-3j"
			imag, err := strconv.ParseFloat(body, 64)
			if err != nil {
				return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
			}
			return MakeComplex(0, imag), nil
		}

		// Has both real and imaginary: "1+2j" or "1-2j"
		realStr := body[:splitIdx]
		imagStr := body[splitIdx:]
		if imagStr == "+" {
			imagStr = "+1"
		} else if imagStr == "-" {
			imagStr = "-1"
		}

		real, err := strconv.ParseFloat(realStr, 64)
		if err != nil {
			return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
		}
		imag, err := strconv.ParseFloat(imagStr, 64)
		if err != nil {
			return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
		}
		return MakeComplex(real, imag), nil
	}

	// No 'j' — pure real
	real, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
	}
	return MakeComplex(real, 0), nil
}

// extractSlots checks classDict for __slots__ and returns the slot names.
// Returns nil if __slots__ is not defined. Removes __slots__ from classDict.
func extractSlots(classDict map[string]Value, bases []*PyClass) []string {
	slotsVal, ok := classDict["__slots__"]
	if !ok {
		return nil
	}
	delete(classDict, "__slots__")

	var slotNames []string
	switch s := slotsVal.(type) {
	case *PyList:
		for _, item := range s.Items {
			if str, ok := item.(*PyString); ok {
				slotNames = append(slotNames, str.Value)
			}
		}
	case *PyTuple:
		for _, item := range s.Items {
			if str, ok := item.(*PyString); ok {
				slotNames = append(slotNames, str.Value)
			}
		}
	}

	// Collect slots from base classes that define __slots__
	for _, base := range bases {
		if base.Slots != nil {
			for _, s := range base.Slots {
				slotNames = append(slotNames, s)
			}
		}
	}

	if slotNames == nil {
		slotNames = []string{} // empty __slots__ = () should be non-nil empty slice
	}
	return slotNames
}

// isValidSlot checks whether name is in the class's allowed slots (including base class slots via MRO).
func isValidSlot(cls *PyClass, name string) bool {
	for _, mroClass := range cls.Mro {
		if mroClass.Slots != nil {
			for _, s := range mroClass.Slots {
				if s == name {
					return true
				}
			}
		}
	}
	return false
}
