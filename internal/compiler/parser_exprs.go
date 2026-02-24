package compiler

import (
	"fmt"
	"strings"

	"github.com/ATSOTECK/rage/internal/model"
)

// Expression parsing using Pratt parser

// parseTupleOrExpr parses a single expression, then checks for trailing commas
// to form an implicit tuple (e.g. `1, 2, 3`). Returns a single Expr or a Tuple.
func (p *Parser) parseTupleOrExpr() model.Expr {
	first := p.parseExpression()
	if first == nil {
		return nil
	}

	if !p.check(model.TK_Comma) {
		return first
	}

	elts := []model.Expr{first}
	for p.match(model.TK_Comma) {
		// Stop if we hit a terminator after the comma (trailing comma).
		if p.check(model.TK_Newline) || p.check(model.TK_EOF) || p.check(model.TK_Comment) ||
			p.check(model.TK_Assign) || p.check(model.TK_RParen) || p.check(model.TK_RBracket) ||
			p.check(model.TK_Colon) {
			break
		}
		next := p.parseExpression()
		if next == nil {
			break
		}
		elts = append(elts, next)
	}

	if len(elts) == 1 {
		return first
	}

	return &model.Tuple{
		Elts:     elts,
		StartPos: first.Pos(),
		EndPos:   elts[len(elts)-1].End(),
	}
}

func (p *Parser) parseExpression() model.Expr {
	return p.parsePrecedence(precLowest)
}

func (p *Parser) parsePrecedence(minPrec int) model.Expr {
	left := p.parseUnaryOrPrimary()
	if left == nil {
		return nil
	}

	for {
		prec := p.infixPrecedence(p.current().Kind)
		if prec < minPrec {
			break
		}

		left = p.parseInfixExpr(left, prec)
	}

	return left
}

func (p *Parser) parseUnaryOrPrimary() model.Expr {
	switch p.current().Kind {
	case model.TK_Not:
		return p.parseNotExpr()
	case model.TK_Minus, model.TK_Plus, model.TK_Tilde:
		return p.parseUnaryExpr()
	case model.TK_Lambda:
		return p.parseLambdaExpr()
	case model.TK_Await:
		return p.parseAwaitExpr()
	case model.TK_Yield:
		return p.parseYieldExpr()
	default:
		return p.parsePrimaryExpr()
	}
}

func (p *Parser) parseNotExpr() model.Expr {
	startPos := p.current().Pos
	p.expect(model.TK_Not)
	operand := p.parsePrecedence(precNot)
	if operand == nil {
		p.addError("expected expression after 'not'")
		return nil
	}
	return &model.UnaryOp{
		Op:       model.TK_Not,
		Operand:  operand,
		StartPos: startPos,
	}
}

func (p *Parser) parseUnaryExpr() model.Expr {
	startPos := p.current().Pos
	op := p.advance().Kind
	operand := p.parsePrecedence(precUnary)
	if operand == nil {
		p.addError("expected expression after unary operator")
		return nil
	}
	return &model.UnaryOp{
		Op:       op,
		Operand:  operand,
		StartPos: startPos,
	}
}

func (p *Parser) parseLambdaExpr() model.Expr {
	startPos := p.current().Pos
	p.expect(model.TK_Lambda)

	args := &model.Arguments{}
	if !p.check(model.TK_Colon) {
		args = p.parseLambdaParameters()
	}

	p.expect(model.TK_Colon)
	body := p.parseExpression()
	if body == nil {
		p.addError("expected expression in lambda body")
		return nil
	}

	return &model.Lambda{
		Args:     args,
		Body:     body,
		StartPos: startPos,
	}
}

func (p *Parser) parseLambdaParameters() *model.Arguments {
	args := &model.Arguments{}

	if p.check(model.TK_Colon) {
		return args
	}

	seenStar := false

	for !p.check(model.TK_Colon) && !p.isAtEnd() {
		if p.match(model.TK_Star) {
			if p.check(model.TK_Identifier) {
				args.VarArg = p.parseArg()
			}
			seenStar = true
			if !p.match(model.TK_Comma) {
				break
			}
			continue
		}

		if p.match(model.TK_DoubleStar) {
			args.KwArg = p.parseArg()
			break
		}

		arg := p.parseSimpleArg()

		var defaultVal model.Expr
		if p.match(model.TK_Assign) {
			defaultVal = p.parseExpression()
		}

		if seenStar {
			args.KwOnlyArgs = append(args.KwOnlyArgs, arg)
			args.KwDefaults = append(args.KwDefaults, defaultVal)
		} else {
			args.Args = append(args.Args, arg)
			if defaultVal != nil {
				args.Defaults = append(args.Defaults, defaultVal)
			}
		}

		if !p.match(model.TK_Comma) {
			break
		}
	}

	return args
}

func (p *Parser) parseSimpleArg() *model.Arg {
	startPos := p.current().Pos
	name := p.parseIdentifier()

	return &model.Arg{
		Arg:      name,
		StartPos: startPos,
		EndPos:   name.End(),
	}
}

func (p *Parser) parseAwaitExpr() model.Expr {
	startPos := p.current().Pos
	p.expect(model.TK_Await)
	value := p.parsePrecedence(precAwait)
	if value == nil {
		p.addError("expected expression after 'await'")
		return nil
	}
	return &model.Await{
		Value:    value,
		StartPos: startPos,
	}
}

func (p *Parser) parseYieldExpr() model.Expr {
	startPos := p.current().Pos
	p.expect(model.TK_Yield)

	if p.match(model.TK_From) {
		value := p.parseExpression()
		if value == nil {
			p.addError("expected expression after 'yield from'")
			return nil
		}
		return &model.YieldFrom{
			Value:    value,
			StartPos: startPos,
		}
	}

	var value model.Expr
	endPos := startPos
	if !p.check(model.TK_Newline) && !p.check(model.TK_Comment) && !p.check(model.TK_RParen) && !p.check(model.TK_EOF) {
		value = p.parseTupleOrExpr()
		if value != nil {
			endPos = value.End()
		}
	}

	return &model.Yield{
		Value:    value,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parsePrimaryExpr() model.Expr {
	var expr model.Expr

	switch p.current().Kind {
	case model.TK_Identifier:
		expr = p.parseIdentifier()
	case model.TK_IntLit:
		expr = p.parseIntLit()
	case model.TK_FloatLit:
		expr = p.parseFloatLit()
	case model.TK_ImaginaryLit:
		expr = p.parseImaginaryLit()
	case model.TK_StringLit:
		expr = p.parseStringLit()
	case model.TK_FStringLit:
		expr = p.parseFStringLit()
	case model.TK_BytesLit:
		expr = p.parseBytesLit()
	case model.TK_True, model.TK_False:
		expr = p.parseBoolLit()
	case model.TK_None:
		expr = p.parseNoneLit()
	case model.TK_Ellipsis:
		expr = p.parseEllipsis()
	case model.TK_LParen:
		expr = p.parseParenExpr()
	case model.TK_LBracket:
		expr = p.parseListExpr()
	case model.TK_LBrace:
		expr = p.parseDictOrSetExpr()
	case model.TK_Star:
		expr = p.parseStarredExpr()
	default:
		return nil
	}

	// Handle postfix operations
	for {
		switch p.current().Kind {
		case model.TK_LParen:
			expr = p.parseCallExpr(expr)
		case model.TK_LBracket:
			expr = p.parseSubscriptExpr(expr)
		case model.TK_Dot:
			expr = p.parseAttributeExpr(expr)
		default:
			return expr
		}
	}
}

func (p *Parser) parseIntLit() model.Expr {
	tok := p.advance()
	return &model.IntLit{
		Value:    tok.Literal,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseFloatLit() model.Expr {
	tok := p.advance()
	return &model.FloatLit{
		Value:    tok.Literal,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseImaginaryLit() model.Expr {
	tok := p.advance()
	return &model.ImaginaryLit{
		Value:    tok.Literal,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseStringLit() model.Expr {
	tok := p.advance()
	return &model.StringLit{
		Value:    tok.Literal,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseFStringLit() model.Expr {
	tok := p.advance()
	content := tok.Literal

	var parts []model.FStringPart
	i := 0

	for i < len(content) {
		// Look for start of expression
		braceIdx := -1
		for j := i; j < len(content); j++ {
			if content[j] == '{' {
				if j+1 < len(content) && content[j+1] == '{' {
					// Escaped brace, skip
					j++
					continue
				}
				braceIdx = j
				break
			}
		}

		if braceIdx == -1 {
			// No more expressions, rest is literal
			literal := p.unescapeFStringBraces(content[i:])
			if literal != "" {
				parts = append(parts, model.FStringPart{
					IsExpr: false,
					Value:  literal,
				})
			}
			break
		}

		// Add literal part before the expression
		if braceIdx > i {
			literal := p.unescapeFStringBraces(content[i:braceIdx])
			if literal != "" {
				parts = append(parts, model.FStringPart{
					IsExpr: false,
					Value:  literal,
				})
			}
		}

		// Find matching closing brace
		depth := 1
		exprStart := braceIdx + 1
		exprEnd := exprStart
		for exprEnd < len(content) && depth > 0 {
			ch := content[exprEnd]
			if ch == '{' {
				depth++
			} else if ch == '}' {
				depth--
			}
			if depth > 0 {
				exprEnd++
			}
		}

		if depth != 0 {
			p.addError("unmatched '{' in f-string")
			return &model.StringLit{Value: content, StartPos: tok.Pos, EndPos: tok.EndPos}
		}

		// Extract expression text and optional format spec
		exprText := content[exprStart:exprEnd]
		formatSpec := ""

		// Check for format spec (: not inside nested braces/brackets/parens)
		colonIdx := p.findFormatSpecColon(exprText)
		if colonIdx != -1 {
			formatSpec = exprText[colonIdx+1:]
			exprText = exprText[:colonIdx]
		}

		// Check for conversion (!r, !s, !a) at end of expression
		var conversion byte
		exprText = strings.TrimSpace(exprText)
		if len(exprText) >= 2 {
			suffix := exprText[len(exprText)-2:]
			if suffix == "!r" || suffix == "!s" || suffix == "!a" {
				conversion = suffix[1]
				exprText = strings.TrimSpace(exprText[:len(exprText)-2])
			}
		}

		// Parse the expression
		if exprText != "" {
			// Create a new parser for the expression
			exprParser := NewParser(exprText)
			expr := exprParser.parseExpression()

			// Propagate any parse errors from the f-string expression
			for _, err := range exprParser.errors {
				p.errors = append(p.errors, ParseError{
					Pos:     tok.Pos,
					Message: fmt.Sprintf("f-string expression: %s", err.Message),
				})
			}

			if expr != nil && len(exprParser.errors) == 0 {
				parts = append(parts, model.FStringPart{
					IsExpr:     true,
					Expr:       expr,
					FormatSpec: formatSpec,
					Conversion: conversion,
				})
			} else if len(exprParser.errors) == 0 {
				p.addError("f-string: empty expression not allowed")
			}
		} else {
			p.addError("f-string: empty expression not allowed")
		}

		i = exprEnd + 1 // Skip past the closing brace
	}

	return &model.FStringLit{
		Parts:    parts,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

// unescapeFStringBraces converts {{ to { and }} to } in f-string literals
func (p *Parser) unescapeFStringBraces(s string) string {
	s = strings.ReplaceAll(s, "{{", "{")
	s = strings.ReplaceAll(s, "}}", "}")
	return s
}

// findFormatSpecColon finds the colon that separates expression from format spec
// It must not be inside nested brackets, parens, or braces
func (p *Parser) findFormatSpecColon(s string) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ':':
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func (p *Parser) parseBytesLit() model.Expr {
	tok := p.advance()
	return &model.BytesLit{
		Value:    tok.Literal,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseBoolLit() model.Expr {
	tok := p.advance()
	return &model.BoolLit{
		Value:    tok.Kind == model.TK_True,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseNoneLit() model.Expr {
	tok := p.advance()
	return &model.NoneLit{
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseEllipsis() model.Expr {
	tok := p.advance()
	return &model.Ellipsis{
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseParenExpr() model.Expr {
	startPos := p.current().Pos
	p.expect(model.TK_LParen)

	// Empty tuple
	if p.check(model.TK_RParen) {
		endTok := p.advance()
		return &model.Tuple{
			StartPos: startPos,
			EndPos:   endTok.EndPos,
		}
	}

	// Generator expression or tuple/grouped expression
	first := p.parseExpression()

	// Check for generator expression
	if p.check(model.TK_For) {
		return p.parseGeneratorExpr(first, startPos)
	}

	// Check for tuple
	if p.check(model.TK_Comma) {
		elts := []model.Expr{first}
		for p.match(model.TK_Comma) {
			if p.check(model.TK_RParen) {
				break
			}
			elts = append(elts, p.parseExpression())
		}
		endTok := p.expect(model.TK_RParen)
		return &model.Tuple{
			Elts:     elts,
			StartPos: startPos,
			EndPos:   endTok.EndPos,
		}
	}

	p.expect(model.TK_RParen)
	return first
}

func (p *Parser) parseGeneratorExpr(first model.Expr, startPos model.Position) model.Expr {
	generators := p.parseComprehensionClauses()
	endTok := p.expect(model.TK_RParen)

	return &model.GeneratorExpr{
		Elt:        first,
		Generators: generators,
		StartPos:   startPos,
		EndPos:     endTok.EndPos,
	}
}

func (p *Parser) parseListExpr() model.Expr {
	startPos := p.current().Pos
	p.expect(model.TK_LBracket)

	if p.check(model.TK_RBracket) {
		endTok := p.advance()
		return &model.List{
			StartPos: startPos,
			EndPos:   endTok.EndPos,
		}
	}

	first := p.parseExpression()

	// List comprehension
	if p.check(model.TK_For) {
		generators := p.parseComprehensionClauses()
		endTok := p.expect(model.TK_RBracket)
		return &model.ListComp{
			Elt:        first,
			Generators: generators,
			StartPos:   startPos,
			EndPos:     endTok.EndPos,
		}
	}

	// Regular list
	elts := []model.Expr{first}
	for p.match(model.TK_Comma) {
		if p.check(model.TK_RBracket) {
			break
		}
		elts = append(elts, p.parseExpression())
	}

	endTok := p.expect(model.TK_RBracket)
	return &model.List{
		Elts:     elts,
		StartPos: startPos,
		EndPos:   endTok.EndPos,
	}
}

func (p *Parser) parseDictOrSetExpr() model.Expr {
	startPos := p.current().Pos
	p.expect(model.TK_LBrace)

	// Empty dict
	if p.check(model.TK_RBrace) {
		endTok := p.advance()
		return &model.Dict{
			StartPos: startPos,
			EndPos:   endTok.EndPos,
		}
	}

	// Check for dict unpacking
	if p.check(model.TK_DoubleStar) {
		return p.parseDictExpr(startPos, nil, nil)
	}

	first := p.parseExpression()

	// Dict
	if p.check(model.TK_Colon) {
		p.advance()
		value := p.parseExpression()

		// Dict comprehension
		if p.check(model.TK_For) {
			generators := p.parseComprehensionClauses()
			endTok := p.expect(model.TK_RBrace)
			return &model.DictComp{
				Key:        first,
				Value:      value,
				Generators: generators,
				StartPos:   startPos,
				EndPos:     endTok.EndPos,
			}
		}

		return p.parseDictExpr(startPos, first, value)
	}

	// Set comprehension
	if p.check(model.TK_For) {
		generators := p.parseComprehensionClauses()
		endTok := p.expect(model.TK_RBrace)
		return &model.SetComp{
			Elt:        first,
			Generators: generators,
			StartPos:   startPos,
			EndPos:     endTok.EndPos,
		}
	}

	// Set
	elts := []model.Expr{first}
	for p.match(model.TK_Comma) {
		if p.check(model.TK_RBrace) {
			break
		}
		elts = append(elts, p.parseExpression())
	}

	endTok := p.expect(model.TK_RBrace)
	return &model.Set{
		Elts:     elts,
		StartPos: startPos,
		EndPos:   endTok.EndPos,
	}
}

func (p *Parser) parseDictExpr(startPos model.Position, firstKey, firstValue model.Expr) model.Expr {
	keys := []model.Expr{firstKey}
	values := []model.Expr{firstValue}

	for p.match(model.TK_Comma) {
		if p.check(model.TK_RBrace) {
			break
		}

		if p.match(model.TK_DoubleStar) {
			keys = append(keys, nil)
			values = append(values, p.parseExpression())
		} else {
			keys = append(keys, p.parseExpression())
			p.expect(model.TK_Colon)
			values = append(values, p.parseExpression())
		}
	}

	endTok := p.expect(model.TK_RBrace)
	return &model.Dict{
		Keys:     keys,
		Values:   values,
		StartPos: startPos,
		EndPos:   endTok.EndPos,
	}
}

func (p *Parser) parseComprehensionClauses() []*model.Comprehension {
	var generators []*model.Comprehension

	for p.check(model.TK_For) || p.check(model.TK_Async) {
		isAsync := p.match(model.TK_Async)
		p.expect(model.TK_For)

		target := p.parseTargetList()
		p.expect(model.TK_In)
		iter := p.parsePrecedence(precOr) // Don't parse ternary in comprehension

		var ifs []model.Expr
		for p.match(model.TK_If) {
			ifs = append(ifs, p.parsePrecedence(precOr))
		}

		generators = append(generators, &model.Comprehension{
			Target:  target,
			Iter:    iter,
			Ifs:     ifs,
			IsAsync: isAsync,
		})
	}

	return generators
}

func (p *Parser) parseStarredExpr() model.Expr {
	startPos := p.current().Pos
	p.expect(model.TK_Star)
	value := p.parsePrimaryExpr()
	if value == nil {
		p.addError("expected expression after '*'")
		return nil
	}
	return &model.Starred{
		Value:    value,
		StartPos: startPos,
	}
}

func (p *Parser) parseCallExpr(fn model.Expr) model.Expr {
	startTok := p.expect(model.TK_LParen)

	var args []model.Expr
	var keywords []*model.Keyword

	for !p.check(model.TK_RParen) && !p.isAtEnd() {
		if p.check(model.TK_DoubleStar) {
			// **kwargs
			startPos := p.current().Pos
			p.advance()
			keywords = append(keywords, &model.Keyword{
				Value:    p.parseExpression(),
				StartPos: startPos,
			})
		} else if p.check(model.TK_Identifier) && p.peek().Kind == model.TK_Assign {
			// keyword=value
			name := p.parseIdentifier()
			p.expect(model.TK_Assign)
			keywords = append(keywords, &model.Keyword{
				Arg:      name,
				Value:    p.parseExpression(),
				StartPos: name.Pos(),
			})
		} else {
			expr := p.parseExpression()
			// Check for generator expression: f(x for x in ...)
			if p.check(model.TK_For) {
				generators := p.parseComprehensionClauses()
				genExpr := &model.GeneratorExpr{
					Elt:        expr,
					Generators: generators,
					StartPos:   startTok.Pos,
				}
				args = append(args, genExpr)
			} else {
				args = append(args, expr)
			}
		}

		if !p.match(model.TK_Comma) {
			break
		}
	}

	endTok := p.expect(model.TK_RParen)

	return &model.Call{
		Func:     fn,
		Args:     args,
		Keywords: keywords,
		EndPos:   endTok.EndPos,
	}
}

func (p *Parser) parseSubscriptExpr(value model.Expr) model.Expr {
	p.expect(model.TK_LBracket)

	slice := p.parseSlice()

	endTok := p.expect(model.TK_RBracket)

	return &model.Subscript{
		Value:  value,
		Slice:  slice,
		EndPos: endTok.EndPos,
	}
}

func (p *Parser) parseSlice() model.Expr {
	startPos := p.current().Pos

	var lower, upper, step model.Expr

	// Check for slice vs regular index
	if !p.check(model.TK_Colon) {
		lower = p.parseExpression()
		if !p.check(model.TK_Colon) {
			// Check for comma-separated tuple index: obj[a, b, c]
			if p.check(model.TK_Comma) {
				elts := []model.Expr{lower}
				for p.match(model.TK_Comma) {
					if p.check(model.TK_RBracket) {
						break
					}
					next := p.parseExpression()
					if next == nil {
						break
					}
					elts = append(elts, next)
				}
				if len(elts) == 1 {
					return lower
				}
				return &model.Tuple{
					Elts:     elts,
					StartPos: lower.Pos(),
					EndPos:   elts[len(elts)-1].End(),
				}
			}
			return lower // Regular index
		}
	}

	if p.match(model.TK_Colon) {
		if !p.check(model.TK_Colon) && !p.check(model.TK_RBracket) {
			upper = p.parseExpression()
		}

		if p.match(model.TK_Colon) {
			if !p.check(model.TK_RBracket) {
				step = p.parseExpression()
			}
		}

		return &model.Slice{
			Lower:    lower,
			Upper:    upper,
			Step:     step,
			StartPos: startPos,
			EndPos:   p.current().Pos,
		}
	}

	return lower
}

func (p *Parser) parseAttributeExpr(value model.Expr) model.Expr {
	p.expect(model.TK_Dot)
	attr := p.parseIdentifier()

	return &model.Attribute{
		Value: value,
		Attr:  attr,
	}
}

func (p *Parser) parseInfixExpr(left model.Expr, prec int) model.Expr {
	switch p.current().Kind {
	case model.TK_If:
		return p.parseTernaryExpr(left)
	case model.TK_And, model.TK_Or:
		return p.parseBoolOpExpr(left, prec)
	case model.TK_Less, model.TK_Greater, model.TK_LessEqual, model.TK_GreaterEqual,
		model.TK_Equal, model.TK_NotEqual, model.TK_In, model.TK_Is, model.TK_Not:
		return p.parseCompareExpr(left)
	case model.TK_Walrus:
		return p.parseWalrusExpr(left)
	default:
		return p.parseBinaryOpExpr(left, prec)
	}
}

func (p *Parser) parseTernaryExpr(body model.Expr) model.Expr {
	p.expect(model.TK_If)
	test := p.parsePrecedence(precOr)
	if test == nil {
		p.addError("expected condition in ternary expression")
		return body
	}
	p.expect(model.TK_Else)
	orElse := p.parsePrecedence(precTernary)
	if orElse == nil {
		p.addError("expected expression after 'else' in ternary expression")
		return body
	}

	return &model.IfExpr{
		Test:   test,
		Body:   body,
		OrElse: orElse,
	}
}

func (p *Parser) parseBoolOpExpr(left model.Expr, prec int) model.Expr {
	op := p.advance().Kind
	values := []model.Expr{left}

	for {
		right := p.parsePrecedence(prec + 1)
		if right == nil {
			p.addError("expected expression after boolean operator")
			break
		}
		values = append(values, right)

		if p.current().Kind != op {
			break
		}
		p.advance()
	}

	return &model.BoolOp{
		Op:     op,
		Values: values,
	}
}

func (p *Parser) parseCompareExpr(left model.Expr) model.Expr {
	var ops []model.TokenKind
	var comparators []model.Expr

	for p.isComparisonOp(p.current().Kind) {
		op := p.advance().Kind

		// Handle 'not in' and 'is not'
		if op == model.TK_Not && p.check(model.TK_In) {
			p.advance()
			op = model.TK_NotIn
		} else if op == model.TK_Is && p.check(model.TK_Not) {
			p.advance()
			op = model.TK_IsNot
		}

		comp := p.parsePrecedence(precBitOr)
		if comp == nil {
			p.addError("expected expression after comparison operator")
			break
		}
		ops = append(ops, op)
		comparators = append(comparators, comp)
	}

	if len(comparators) == 0 {
		return left
	}

	return &model.Compare{
		Left:        left,
		Ops:         ops,
		Comparators: comparators,
	}
}

func (p *Parser) parseWalrusExpr(left model.Expr) model.Expr {
	p.expect(model.TK_Walrus)
	value := p.parseExpression()

	target, ok := left.(*model.Identifier)
	if !ok {
		p.addError("walrus operator target must be an identifier")
		return left
	}

	if value == nil {
		p.addError("expected expression after ':='")
		return left
	}

	return &model.NamedExpr{
		Target: target,
		Value:  value,
	}
}

func (p *Parser) parseBinaryOpExpr(left model.Expr, prec int) model.Expr {
	op := p.advance().Kind

	// Power operator is right-associative
	assoc := 1
	if op == model.TK_DoubleStar {
		assoc = 0
	}

	right := p.parsePrecedence(prec + assoc)
	if right == nil {
		p.addError("expected expression after binary operator")
		return left
	}

	return &model.BinaryOp{
		Left:  left,
		Op:    op,
		Right: right,
	}
}

func (p *Parser) infixPrecedence(kind model.TokenKind) int {
	switch kind {
	case model.TK_Walrus:
		return precWalrus
	case model.TK_If:
		return precTernary
	case model.TK_Or:
		return precOr
	case model.TK_And:
		return precAnd
	case model.TK_Less, model.TK_Greater, model.TK_LessEqual, model.TK_GreaterEqual,
		model.TK_Equal, model.TK_NotEqual, model.TK_In, model.TK_Is, model.TK_Not:
		return precComparison
	case model.TK_Pipe:
		return precBitOr
	case model.TK_Caret:
		return precBitXor
	case model.TK_Ampersand:
		return precBitAnd
	case model.TK_LShift, model.TK_RShift:
		return precShift
	case model.TK_Plus, model.TK_Minus:
		return precAddSub
	case model.TK_Star, model.TK_Slash, model.TK_DoubleSlash, model.TK_Percent, model.TK_At:
		return precMulDiv
	case model.TK_DoubleStar:
		return precPower
	default:
		return precNone
	}
}

func (p *Parser) isComparisonOp(kind model.TokenKind) bool {
	switch kind {
	case model.TK_Less, model.TK_Greater, model.TK_LessEqual, model.TK_GreaterEqual,
		model.TK_Equal, model.TK_NotEqual, model.TK_In, model.TK_Is, model.TK_Not:
		return true
	}
	return false
}
