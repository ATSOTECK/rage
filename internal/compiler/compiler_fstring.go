package compiler

import (
	"fmt"
	"strings"

	"github.com/ATSOTECK/rage/internal/model"
	"github.com/ATSOTECK/rage/internal/runtime"
)

// F-string compilation.
//
// Each replacement field compiles to a call of a formatting builtin:
// {x} -> format(x), {x!r} -> repr(x), {x:spec} -> format(x, spec) and
// {x!r:spec} -> format(repr(x), spec). The parts are then concatenated.

func (c *Compiler) compileFString(e *model.FStringLit) {
	if len(e.Parts) == 0 {
		// Empty f-string
		c.emitLoadConst("")
		return
	}

	for i, part := range e.Parts {
		if part.IsExpr {
			c.compileFStringField(part.Expr, part.Conversion, part.FormatSpec, e.Pos(), true)
		} else {
			c.emitLoadConst(part.Value)
		}
		// Concatenate with previous part
		if i > 0 {
			c.emit(runtime.OpBinaryAdd)
		}
	}
}

// compileFStringField leaves the formatted text of one replacement field on
// the stack. allowNested permits replacement fields inside spec (one level,
// as in CPython).
func (c *Compiler) compileFStringField(expr model.Expr, conversion byte, spec string, pos model.Position, allowNested bool) {
	convName := ""
	switch conversion {
	case 'r':
		convName = "repr"
	case 's':
		convName = "str"
	case 'a':
		convName = "ascii"
	}

	if spec == "" {
		// The conversion result is already a str, so format() would be a no-op
		name := convName
		if name == "" {
			name = "format"
		}
		c.emitArg(runtime.OpLoadGlobal, c.addName(name))
		c.compileExpr(expr)
		c.emitArg(runtime.OpCall, 1)
		return
	}

	c.emitArg(runtime.OpLoadGlobal, c.addName("format"))
	if convName != "" {
		c.emitArg(runtime.OpLoadGlobal, c.addName(convName))
		c.compileExpr(expr)
		c.emitArg(runtime.OpCall, 1)
	} else {
		c.compileExpr(expr)
	}
	c.compileFStringSpec(spec, pos, allowNested)
	c.emitArg(runtime.OpCall, 2)
}

// compileFStringSpec pushes a format spec, evaluating any nested replacement
// fields such as the width in f"{x:>{width}}".
func (c *Compiler) compileFStringSpec(spec string, pos model.Position, allowNested bool) {
	if !strings.ContainsAny(spec, "{}") {
		c.emitLoadConst(spec)
		return
	}
	if !allowNested {
		c.error(pos, "f-string: expressions nested too deeply")
		c.emitLoadConst("")
		return
	}
	pieces, err := parseFStringSpec(spec)
	if err != nil {
		c.error(pos, "%s", err.Error())
		c.emitLoadConst("")
		return
	}
	pushed := 0
	for _, piece := range pieces {
		if piece.literal != "" {
			c.emitLoadConst(piece.literal)
			if pushed++; pushed > 1 {
				c.emit(runtime.OpBinaryAdd)
			}
		}
		if piece.expr != nil {
			c.compileFStringField(piece.expr, piece.conversion, piece.spec, pos, false)
			if pushed++; pushed > 1 {
				c.emit(runtime.OpBinaryAdd)
			}
		}
	}
	if pushed == 0 {
		c.emitLoadConst("")
	}
}

// fstringSpecPiece is literal text in an f-string format spec, optionally
// followed by a nested replacement field.
type fstringSpecPiece struct {
	literal    string
	expr       model.Expr // nil when the piece is only literal text
	conversion byte
	spec       string
}

// parseFStringSpec splits a format spec (kept as raw text by the parser) into
// literal text and nested replacement fields.
func parseFStringSpec(spec string) ([]fstringSpecPiece, error) {
	var pieces []fstringSpecPiece
	var literal strings.Builder
	i := 0
	for i < len(spec) {
		ch := spec[i]
		if ch == '}' {
			return nil, fmt.Errorf("f-string: single '}' is not allowed")
		}
		if ch != '{' {
			literal.WriteByte(ch)
			i++
			continue
		}
		// Find the matching closing brace
		depth := 1
		j := i + 1
		for j < len(spec) {
			if spec[j] == '{' {
				depth++
			} else if spec[j] == '}' {
				depth--
				if depth == 0 {
					break
				}
			}
			j++
		}
		if depth != 0 {
			return nil, fmt.Errorf("f-string: expecting '}'")
		}
		piece, err := parseFStringNestedField(spec[i+1 : j])
		if err != nil {
			return nil, err
		}
		piece.literal = literal.String()
		literal.Reset()
		pieces = append(pieces, piece)
		i = j + 1
	}
	if literal.Len() > 0 {
		pieces = append(pieces, fstringSpecPiece{literal: literal.String()})
	}
	return pieces, nil
}

// fstringSpecExprs returns the expressions of the replacement fields nested
// in an f-string format spec, for scope analysis.
func fstringSpecExprs(spec string) []model.Expr {
	if !strings.Contains(spec, "{") {
		return nil
	}
	pieces, err := parseFStringSpec(spec)
	if err != nil {
		return nil
	}
	var exprs []model.Expr
	for _, piece := range pieces {
		if piece.expr != nil {
			exprs = append(exprs, piece.expr)
		}
	}
	return exprs
}

// parseFStringNestedField parses the text of a nested replacement field:
// an expression with an optional !conversion and :spec.
func parseFStringNestedField(text string) (fstringSpecPiece, error) {
	var piece fstringSpecPiece
	if colon := (&Parser{}).findFormatSpecColon(text); colon >= 0 {
		piece.spec = text[colon+1:]
		text = text[:colon]
	}
	text = strings.TrimSpace(text)
	if len(text) >= 2 {
		switch text[len(text)-2:] {
		case "!r", "!s", "!a":
			piece.conversion = text[len(text)-1]
			text = strings.TrimSpace(text[:len(text)-2])
		}
	}
	if text == "" {
		return piece, fmt.Errorf("f-string: valid expression required before '}'")
	}
	parser := NewParser(text)
	expr := parser.parseExpression()
	if len(parser.errors) > 0 {
		return piece, fmt.Errorf("f-string expression: %s", parser.errors[0].Message)
	}
	if expr == nil {
		return piece, fmt.Errorf("f-string: valid expression required before '}'")
	}
	piece.expr = expr
	return piece, nil
}
