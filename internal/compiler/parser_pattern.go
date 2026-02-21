package compiler

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/model"
)

func (p *Parser) parseMatchStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Match)

	subject := p.parseExpression()
	p.expect(model.TK_Colon)
	p.expect(model.TK_Newline)
	p.expect(model.TK_Indent)

	var cases []*model.MatchCase
	for p.check(model.TK_Case) {
		cases = append(cases, p.parseMatchCase())
	}

	p.expect(model.TK_Dedent)

	var endPos model.Position
	if len(cases) > 0 {
		endPos = cases[len(cases)-1].End()
	} else {
		endPos = startPos
	}

	return &model.Match{
		Subject:  subject,
		Cases:    cases,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseMatchCase() *model.MatchCase {
	startPos := p.current().Pos
	p.expect(model.TK_Case)

	pattern := p.parsePattern()

	var guard model.Expr
	if p.match(model.TK_If) {
		guard = p.parseExpression()
	}

	p.expect(model.TK_Colon)
	p.expect(model.TK_Newline)

	body := p.parseBlock()

	var endPos model.Position
	if len(body) > 0 {
		endPos = body[len(body)-1].End()
	} else {
		endPos = startPos
	}

	return &model.MatchCase{
		Pattern:  pattern,
		Guard:    guard,
		Body:     body,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parsePattern() model.Pattern {
	// Parse or pattern: pattern | pattern | ...
	return p.parseOrPattern()
}

func (p *Parser) parseOrPattern() model.Pattern {
	startPos := p.current().Pos
	pattern := p.parseAsPattern()

	if p.check(model.TK_Pipe) {
		patterns := []model.Pattern{pattern}
		for p.match(model.TK_Pipe) {
			patterns = append(patterns, p.parseAsPattern())
		}
		endPos := patterns[len(patterns)-1].End()
		return &model.MatchOr{
			Patterns: patterns,
			StartPos: startPos,
			EndPos:   endPos,
		}
	}

	return pattern
}

func (p *Parser) parseAsPattern() model.Pattern {
	startPos := p.current().Pos
	pattern := p.parseClosedPattern()

	if p.match(model.TK_As) {
		name := p.parseIdentifier()
		return &model.MatchAs{
			Pattern:  pattern,
			Name:     name,
			StartPos: startPos,
			EndPos:   name.End(),
		}
	}

	return pattern
}

func (p *Parser) parseClosedPattern() model.Pattern {
	startPos := p.current().Pos

	switch p.current().Kind {
	case model.TK_Identifier:
		return p.parseCaptureOrClassPattern()

	case model.TK_IntLit:
		// Integer literal pattern
		expr := p.parseIntLit()
		return &model.MatchValue{
			Value:    expr,
			StartPos: startPos,
			EndPos:   expr.End(),
		}

	case model.TK_FloatLit:
		// Float literal pattern
		expr := p.parseFloatLit()
		return &model.MatchValue{
			Value:    expr,
			StartPos: startPos,
			EndPos:   expr.End(),
		}

	case model.TK_Minus:
		// Negative number pattern
		p.advance()
		var expr model.Expr
		if p.current().Kind == model.TK_IntLit {
			expr = p.parseIntLit()
		} else if p.current().Kind == model.TK_FloatLit {
			expr = p.parseFloatLit()
		} else {
			p.addError("expected number after '-' in pattern")
			return nil
		}
		negExpr := &model.UnaryOp{
			Op:       model.TK_Minus,
			Operand:  expr,
			StartPos: startPos,
		}
		return &model.MatchValue{
			Value:    negExpr,
			StartPos: startPos,
			EndPos:   expr.End(),
		}

	case model.TK_StringLit:
		expr := p.parseStringLit()
		return &model.MatchValue{
			Value:    expr,
			StartPos: startPos,
			EndPos:   expr.End(),
		}

	case model.TK_True, model.TK_False, model.TK_None:
		// Singleton patterns (matched with 'is')
		expr := p.parsePrimaryExpr()
		return &model.MatchSingleton{
			Value:    expr,
			StartPos: startPos,
			EndPos:   expr.End(),
		}

	case model.TK_LBracket:
		// Sequence pattern [a, b, *rest, c]
		return p.parseSequencePattern()

	case model.TK_LParen:
		// Either a grouped pattern or tuple pattern
		return p.parseGroupOrTuplePattern()

	case model.TK_LBrace:
		// Mapping pattern {"key": pattern, **rest}
		return p.parseMappingPattern()

	case model.TK_Star:
		// Starred pattern (only valid inside sequence)
		return p.parseStarPattern()

	default:
		p.addError(fmt.Sprintf("unexpected token in pattern: %s", p.current().Kind))
		return nil
	}
}

func (p *Parser) parseCaptureOrClassPattern() model.Pattern {
	startPos := p.current().Pos
	name := p.parseIdentifier()

	// Wildcard pattern
	if name.Name == "_" {
		return &model.MatchAs{
			StartPos: startPos,
			EndPos:   name.End(),
		}
	}

	// Check for dotted name (for class patterns like mod.ClassName)
	var expr model.Expr = name
	for p.match(model.TK_Dot) {
		attrName := p.parseIdentifier()
		expr = &model.Attribute{
			Value: expr,
			Attr:  attrName,
		}
	}

	// Class pattern: ClassName(patterns)
	if p.check(model.TK_LParen) {
		return p.parseClassPattern(expr, startPos)
	}

	// If we have a dotted name without call, it's a value pattern
	if _, isIdent := expr.(*model.Identifier); !isIdent {
		return &model.MatchValue{
			Value:    expr,
			StartPos: startPos,
			EndPos:   expr.End(),
		}
	}

	// Simple capture pattern (just a name)
	return &model.MatchAs{
		Name:     name,
		StartPos: startPos,
		EndPos:   name.End(),
	}
}

func (p *Parser) parseClassPattern(cls model.Expr, startPos model.Position) model.Pattern {
	p.expect(model.TK_LParen)

	var patterns []model.Pattern
	var kwdAttrs []*model.Identifier
	var kwdPatterns []model.Pattern
	inKeywords := false

	for !p.check(model.TK_RParen) && !p.isAtEnd() {
		patternStart := p.current().Pos

		// Check for keyword argument: name=pattern
		if p.check(model.TK_Identifier) && p.peek().Kind == model.TK_Assign {
			inKeywords = true
			attrName := p.parseIdentifier()
			p.expect(model.TK_Assign)
			pattern := p.parsePattern()
			kwdAttrs = append(kwdAttrs, attrName)
			kwdPatterns = append(kwdPatterns, pattern)
		} else {
			if inKeywords {
				p.errors = append(p.errors, ParseError{
					Pos:     patternStart,
					Message: "positional pattern follows keyword pattern",
				})
			}
			pattern := p.parsePattern()
			patterns = append(patterns, pattern)
		}

		if !p.match(model.TK_Comma) {
			break
		}
	}

	endPos := p.current().Pos
	p.expect(model.TK_RParen)

	return &model.MatchClass{
		Cls:         cls,
		Patterns:    patterns,
		KwdAttrs:    kwdAttrs,
		KwdPatterns: kwdPatterns,
		StartPos:    startPos,
		EndPos:      endPos,
	}
}

func (p *Parser) parseSequencePattern() model.Pattern {
	startPos := p.current().Pos
	p.expect(model.TK_LBracket)

	var patterns []model.Pattern
	for !p.check(model.TK_RBracket) && !p.isAtEnd() {
		pattern := p.parsePattern()
		patterns = append(patterns, pattern)

		if !p.match(model.TK_Comma) {
			break
		}
	}

	endPos := p.current().Pos
	p.expect(model.TK_RBracket)

	return &model.MatchSequence{
		Patterns: patterns,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseGroupOrTuplePattern() model.Pattern {
	startPos := p.current().Pos
	p.expect(model.TK_LParen)

	// Empty tuple
	if p.check(model.TK_RParen) {
		endPos := p.current().Pos
		p.advance()
		return &model.MatchSequence{
			Patterns: nil,
			StartPos: startPos,
			EndPos:   endPos,
		}
	}

	// Parse first pattern
	first := p.parsePattern()

	// Check if it's a tuple (has comma) or just grouped pattern
	if p.match(model.TK_Comma) {
		// It's a tuple pattern
		patterns := []model.Pattern{first}

		for !p.check(model.TK_RParen) && !p.isAtEnd() {
			pattern := p.parsePattern()
			patterns = append(patterns, pattern)

			if !p.match(model.TK_Comma) {
				break
			}
		}

		endPos := p.current().Pos
		p.expect(model.TK_RParen)

		return &model.MatchSequence{
			Patterns: patterns,
			StartPos: startPos,
			EndPos:   endPos,
		}
	}

	// Just a grouped pattern
	p.expect(model.TK_RParen)
	return first
}

func (p *Parser) parseMappingPattern() model.Pattern {
	startPos := p.current().Pos
	p.expect(model.TK_LBrace)

	var keys []model.Expr
	var patterns []model.Pattern
	var rest *model.Identifier

	for !p.check(model.TK_RBrace) && !p.isAtEnd() {
		// Check for **rest
		if p.match(model.TK_DoubleStar) {
			rest = p.parseIdentifier()
			// **rest must be last
			p.match(model.TK_Comma) // optional trailing comma
			break
		}

		// Parse key (must be a literal or simple value)
		var key model.Expr
		switch p.current().Kind {
		case model.TK_StringLit:
			key = p.parseStringLit()
		case model.TK_IntLit:
			key = p.parseIntLit()
		case model.TK_True, model.TK_False, model.TK_None:
			key = p.parsePrimaryExpr()
		default:
			p.addError("mapping pattern keys must be literals")
			return nil
		}

		p.expect(model.TK_Colon)
		pattern := p.parsePattern()

		keys = append(keys, key)
		patterns = append(patterns, pattern)

		if !p.match(model.TK_Comma) {
			break
		}
	}

	endPos := p.current().Pos
	p.expect(model.TK_RBrace)

	return &model.MatchMapping{
		Keys:     keys,
		Patterns: patterns,
		Rest:     rest,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseStarPattern() model.Pattern {
	startPos := p.current().Pos
	p.expect(model.TK_Star)

	var name *model.Identifier
	if p.check(model.TK_Identifier) {
		ident := p.parseIdentifier()
		if ident.Name != "_" {
			name = ident
		}
	}

	endPos := p.current().Pos
	return &model.MatchStar{
		Name:     name,
		StartPos: startPos,
		EndPos:   endPos,
	}
}
