package compiler

import (
	"github.com/ATSOTECK/rage/internal/model"
)

// Statement parsing

func (p *Parser) parseStatement() model.Stmt {
	switch p.current().Kind {
	case model.TK_If:
		return p.parseIfStmt()
	case model.TK_While:
		return p.parseWhileStmt()
	case model.TK_For:
		return p.parseForStmt()
	case model.TK_Def:
		return p.parseFunctionDef(false)
	case model.TK_Async:
		return p.parseAsyncStmt()
	case model.TK_Class:
		return p.parseClassDef()
	case model.TK_Return:
		return p.parseReturnStmt()
	case model.TK_Pass:
		return p.parsePassStmt()
	case model.TK_Break:
		return p.parseBreakStmt()
	case model.TK_Continue:
		return p.parseContinueStmt()
	case model.TK_Import:
		return p.parseImportStmt()
	case model.TK_From:
		return p.parseFromImportStmt()
	case model.TK_Raise:
		return p.parseRaiseStmt()
	case model.TK_Try:
		return p.parseTryStmt()
	case model.TK_With:
		return p.parseWithStmt()
	case model.TK_Assert:
		return p.parseAssertStmt()
	case model.TK_Del:
		return p.parseDelStmt()
	case model.TK_Global:
		return p.parseGlobalStmt()
	case model.TK_Nonlocal:
		return p.parseNonlocalStmt()
	case model.TK_Match:
		return p.parseMatchStmt()
	case model.TK_At:
		return p.parseDecorated()
	default:
		return p.parseExpressionStmt()
	}
}

func (p *Parser) parseExpressionStmt() model.Stmt {
	expr := p.parseExpression()
	if expr == nil {
		p.advance() // skip problematic token
		return nil
	}

	// Check for tuple unpacking (a, b, c = ...)
	if p.check(model.TK_Comma) {
		elts := []model.Expr{expr}
		startPos := expr.Pos()
		for p.match(model.TK_Comma) {
			if p.check(model.TK_Assign) || p.check(model.TK_Newline) || p.isAtEnd() {
				break
			}
			next := p.parseExpression()
			if next == nil {
				break
			}
			elts = append(elts, next)
		}
		if len(elts) > 1 {
			expr = &model.Tuple{
				Elts:     elts,
				StartPos: startPos,
				EndPos:   elts[len(elts)-1].End(),
			}
		}
	}

	// Check for assignment
	if p.check(model.TK_Assign) {
		return p.parseAssignment(expr)
	}

	// Check for augmented assignment
	if p.isAugAssignOp(p.current().Kind) {
		return p.parseAugAssignment(expr)
	}

	// Check for annotated assignment
	if p.check(model.TK_Colon) {
		return p.parseAnnotatedAssignment(expr)
	}

	p.match(model.TK_Newline)
	return &model.ExprStmt{Value: expr}
}

func (p *Parser) parseAssignment(first model.Expr) model.Stmt {
	targets := []model.Expr{first}

	for p.match(model.TK_Assign) {
		next := p.parseTupleOrExpr()
		if p.check(model.TK_Assign) {
			targets = append(targets, next)
		} else {
			p.match(model.TK_Newline)
			return &model.Assign{
				Targets: targets,
				Value:   next,
			}
		}
	}

	p.addError("invalid assignment")
	return nil
}

func (p *Parser) parseAugAssignment(target model.Expr) model.Stmt {
	op := p.current().Kind
	p.advance()

	value := p.parseExpression()
	p.match(model.TK_Newline)

	return &model.AugAssign{
		Target: target,
		Op:     op,
		Value:  value,
	}
}

func (p *Parser) parseAnnotatedAssignment(target model.Expr) model.Stmt {
	startPos := target.Pos()
	p.expect(model.TK_Colon)

	annotation := p.parseExpression()

	var value model.Expr
	if p.match(model.TK_Assign) {
		value = p.parseExpression()
	}

	p.match(model.TK_Newline)

	return &model.AnnAssign{
		Target:     target,
		Annotation: annotation,
		Value:      value,
		Simple:     true,
		StartPos:   startPos,
	}
}

func (p *Parser) isAugAssignOp(kind model.TokenKind) bool {
	switch kind {
	case model.TK_PlusAssign, model.TK_MinusAssign, model.TK_StarAssign,
		model.TK_SlashAssign, model.TK_DoubleSlashAssign, model.TK_PercentAssign,
		model.TK_DoubleStarAssign, model.TK_AtAssign, model.TK_AmpersandAssign,
		model.TK_PipeAssign, model.TK_CaretAssign, model.TK_LShiftAssign,
		model.TK_RShiftAssign:
		return true
	}
	return false
}

func (p *Parser) parseIfStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_If)

	test := p.parseExpression()
	p.expect(model.TK_Colon)
	p.expect(model.TK_Newline)

	body := p.parseBlock()

	var orElse []model.Stmt
	for p.check(model.TK_Elif) {
		elifPos := p.current().Pos
		p.advance()
		elifTest := p.parseExpression()
		p.expect(model.TK_Colon)
		p.expect(model.TK_Newline)
		elifBody := p.parseBlock()

		// Wrap elif as nested If
		var elifEndPos model.Position
		if len(elifBody) > 0 {
			elifEndPos = elifBody[len(elifBody)-1].End()
		} else {
			elifEndPos = elifPos
		}
		orElse = []model.Stmt{&model.If{
			Test:     elifTest,
			Body:     elifBody,
			StartPos: elifPos,
			EndPos:   elifEndPos,
		}}
	}

	if p.check(model.TK_Else) {
		p.advance()
		p.expect(model.TK_Colon)
		p.expect(model.TK_Newline)
		if len(orElse) > 0 {
			// Attach else to the last elif
			lastElif := orElse[0].(*model.If)
			lastElif.OrElse = p.parseBlock()
		} else {
			orElse = p.parseBlock()
		}
	}

	var endPos model.Position
	if len(orElse) > 0 {
		endPos = orElse[len(orElse)-1].End()
	} else if len(body) > 0 {
		endPos = body[len(body)-1].End()
	} else {
		endPos = startPos
	}

	return &model.If{
		Test:     test,
		Body:     body,
		OrElse:   orElse,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseWhileStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_While)

	test := p.parseExpression()
	p.expect(model.TK_Colon)
	p.expect(model.TK_Newline)

	body := p.parseBlock()

	var orElse []model.Stmt
	if p.check(model.TK_Else) {
		p.advance()
		p.expect(model.TK_Colon)
		p.expect(model.TK_Newline)
		orElse = p.parseBlock()
	}

	var endPos model.Position
	if len(orElse) > 0 {
		endPos = orElse[len(orElse)-1].End()
	} else if len(body) > 0 {
		endPos = body[len(body)-1].End()
	} else {
		endPos = startPos
	}

	return &model.While{
		Test:     test,
		Body:     body,
		OrElse:   orElse,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseForStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_For)

	target := p.parseTargetList()
	p.expect(model.TK_In)
	iter := p.parseExpression()
	p.expect(model.TK_Colon)
	p.expect(model.TK_Newline)

	body := p.parseBlock()

	var orElse []model.Stmt
	if p.check(model.TK_Else) {
		p.advance()
		p.expect(model.TK_Colon)
		p.expect(model.TK_Newline)
		orElse = p.parseBlock()
	}

	var endPos model.Position
	if len(orElse) > 0 {
		endPos = orElse[len(orElse)-1].End()
	} else if len(body) > 0 {
		endPos = body[len(body)-1].End()
	} else {
		endPos = startPos
	}

	return &model.For{
		Target:   target,
		Iter:     iter,
		Body:     body,
		OrElse:   orElse,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseTargetList() model.Expr {
	first := p.parsePrimaryExpr()
	if first == nil {
		p.addError("expected target expression")
		return nil
	}

	if p.check(model.TK_Comma) {
		elts := []model.Expr{first}
		startPos := first.Pos()
		for p.match(model.TK_Comma) {
			if p.check(model.TK_In) {
				break
			}
			next := p.parsePrimaryExpr()
			if next == nil {
				break
			}
			elts = append(elts, next)
		}
		return &model.Tuple{
			Elts:     elts,
			StartPos: startPos,
			EndPos:   elts[len(elts)-1].End(),
		}
	}

	return first
}

func (p *Parser) parseFunctionDef(isAsync bool) model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Def)

	name := p.parseIdentifier()
	p.expect(model.TK_LParen)
	args := p.parseParameters()
	p.expect(model.TK_RParen)

	var returns model.Expr
	if p.match(model.TK_Arrow) {
		returns = p.parseExpression()
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

	return &model.FunctionDef{
		Name:     name,
		Args:     args,
		Body:     body,
		Returns:  returns,
		IsAsync:  isAsync,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseParameters() *model.Arguments {
	args := &model.Arguments{}

	if p.check(model.TK_RParen) {
		return args
	}

	seenSlash := false
	seenStar := false

	for !p.check(model.TK_RParen) && !p.isAtEnd() {
		if p.match(model.TK_Slash) {
			// Position-only parameters marker
			args.PosOnlyArgs = args.Args
			args.Args = nil
			seenSlash = true
			p.match(model.TK_Comma)
			continue
		}

		if p.match(model.TK_Star) {
			if p.check(model.TK_Identifier) {
				// *args
				args.VarArg = p.parseArg()
			}
			seenStar = true
			p.match(model.TK_Comma)
			continue
		}

		if p.match(model.TK_DoubleStar) {
			// **kwargs
			args.KwArg = p.parseArg()
			p.match(model.TK_Comma)
			continue
		}

		arg := p.parseArg()

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
	_ = seenSlash

	return args
}

func (p *Parser) parseArg() *model.Arg {
	startPos := p.current().Pos
	name := p.parseIdentifier()

	var annotation model.Expr
	if p.match(model.TK_Colon) {
		annotation = p.parseExpression()
	}

	endPos := p.current().Pos
	if annotation != nil {
		endPos = annotation.End()
	}

	return &model.Arg{
		Arg:        name,
		Annotation: annotation,
		StartPos:   startPos,
		EndPos:     endPos,
	}
}

func (p *Parser) parseAsyncStmt() model.Stmt {
	p.expect(model.TK_Async)

	switch p.current().Kind {
	case model.TK_Def:
		return p.parseFunctionDef(true)
	case model.TK_For:
		stmt := p.parseForStmt().(*model.For)
		stmt.IsAsync = true
		return stmt
	case model.TK_With:
		stmt := p.parseWithStmt().(*model.With)
		stmt.IsAsync = true
		return stmt
	default:
		p.addError("expected 'def', 'for', or 'with' after 'async'")
		return nil
	}
}

func (p *Parser) parseClassDef() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Class)

	name := p.parseIdentifier()

	var bases []model.Expr
	var keywords []*model.Keyword

	if p.match(model.TK_LParen) {
		for !p.check(model.TK_RParen) && !p.isAtEnd() {
			if p.check(model.TK_Identifier) && p.peek().Kind == model.TK_Assign {
				// Keyword argument
				kwName := p.parseIdentifier()
				p.expect(model.TK_Assign)
				kwValue := p.parseExpression()
				keywords = append(keywords, &model.Keyword{
					Arg:      kwName,
					Value:    kwValue,
					StartPos: kwName.Pos(),
				})
			} else {
				bases = append(bases, p.parseExpression())
			}

			if !p.match(model.TK_Comma) {
				break
			}
		}
		p.expect(model.TK_RParen)
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

	return &model.ClassDef{
		Name:     name,
		Bases:    bases,
		Keywords: keywords,
		Body:     body,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseReturnStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Return)

	var value model.Expr
	endPos := startPos

	if !p.check(model.TK_Newline) && !p.check(model.TK_EOF) {
		value = p.parseTupleOrExpr()
		if value != nil {
			endPos = value.End()
		}
	}

	p.match(model.TK_Newline)

	return &model.Return{
		Value:    value,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parsePassStmt() model.Stmt {
	tok := p.expect(model.TK_Pass)
	p.match(model.TK_Newline)
	return &model.Pass{
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseBreakStmt() model.Stmt {
	tok := p.expect(model.TK_Break)
	p.match(model.TK_Newline)
	return &model.Break{
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseContinueStmt() model.Stmt {
	tok := p.expect(model.TK_Continue)
	p.match(model.TK_Newline)
	return &model.Continue{
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}

func (p *Parser) parseImportStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Import)

	var names []*model.Alias
	for {
		alias := p.parseImportAlias()
		names = append(names, alias)

		if !p.match(model.TK_Comma) {
			break
		}
	}

	endPos := names[len(names)-1].End()
	p.match(model.TK_Newline)

	return &model.Import{
		Names:    names,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseFromImportStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_From)

	level := 0
	for p.match(model.TK_Dot) {
		level++
	}

	var module *model.Identifier
	if p.check(model.TK_Identifier) {
		module = p.parseDottedName()
	}

	p.expect(model.TK_Import)

	var names []*model.Alias
	if p.match(model.TK_Star) {
		names = []*model.Alias{{
			Name: &model.Identifier{
				Name:     "*",
				StartPos: p.tokens[p.pos-1].Pos,
				EndPos:   p.tokens[p.pos-1].EndPos,
			},
			StartPos: p.tokens[p.pos-1].Pos,
			EndPos:   p.tokens[p.pos-1].EndPos,
		}}
	} else {
		inParens := p.match(model.TK_LParen)
		for {
			alias := p.parseImportAlias()
			names = append(names, alias)

			if !p.match(model.TK_Comma) {
				break
			}
			p.skipNewlines()
		}
		if inParens {
			p.expect(model.TK_RParen)
		}
	}

	endPos := p.current().Pos
	p.match(model.TK_Newline)

	return &model.ImportFrom{
		Module:   module,
		Names:    names,
		Level:    level,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseImportAlias() *model.Alias {
	startPos := p.current().Pos
	name := p.parseDottedName()

	var asName *model.Identifier
	if p.match(model.TK_As) {
		asName = p.parseIdentifier()
	}

	endPos := p.current().Pos
	if asName != nil {
		endPos = asName.End()
	}

	return &model.Alias{
		Name:     name,
		AsName:   asName,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseDottedName() *model.Identifier {
	startPos := p.current().Pos
	name := p.expect(model.TK_Identifier).Literal

	for p.match(model.TK_Dot) {
		name += "." + p.expect(model.TK_Identifier).Literal
	}

	return &model.Identifier{
		Name:     name,
		StartPos: startPos,
		EndPos:   p.tokens[p.pos-1].EndPos,
	}
}

func (p *Parser) parseRaiseStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Raise)

	var exc, cause model.Expr
	endPos := startPos

	if !p.check(model.TK_Newline) && !p.check(model.TK_Comment) && !p.check(model.TK_EOF) {
		exc = p.parseExpression()
		if exc != nil {
			endPos = exc.End()
		}

		if p.match(model.TK_From) {
			cause = p.parseExpression()
			if cause != nil {
				endPos = cause.End()
			}
		}
	}

	p.match(model.TK_Comment)
	p.match(model.TK_Newline)

	return &model.Raise{
		Exc:      exc,
		Cause:    cause,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseTryStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Try)
	p.expect(model.TK_Colon)
	p.expect(model.TK_Newline)

	body := p.parseBlock()

	var handlers []*model.ExceptHandler
	for p.check(model.TK_Except) {
		handlers = append(handlers, p.parseExceptHandler())
	}

	// Validate: no mixing except and except* in the same try block
	if len(handlers) > 1 {
		hasStar := false
		hasPlain := false
		for _, h := range handlers {
			if h.IsStar {
				hasStar = true
			} else {
				hasPlain = true
			}
		}
		if hasStar && hasPlain {
			p.addError("cannot mix except and except* in the same try block")
		}
	}

	var orElse []model.Stmt
	if p.check(model.TK_Else) {
		p.advance()
		p.expect(model.TK_Colon)
		p.expect(model.TK_Newline)
		orElse = p.parseBlock()
	}

	var finalBody []model.Stmt
	if p.check(model.TK_Finally) {
		p.advance()
		p.expect(model.TK_Colon)
		p.expect(model.TK_Newline)
		finalBody = p.parseBlock()
	}

	var endPos model.Position
	if len(finalBody) > 0 {
		endPos = finalBody[len(finalBody)-1].End()
	} else if len(orElse) > 0 {
		endPos = orElse[len(orElse)-1].End()
	} else if len(handlers) > 0 {
		endPos = handlers[len(handlers)-1].End()
	} else if len(body) > 0 {
		endPos = body[len(body)-1].End()
	} else {
		endPos = startPos
	}

	return &model.Try{
		Body:      body,
		Handlers:  handlers,
		OrElse:    orElse,
		FinalBody: finalBody,
		StartPos:  startPos,
		EndPos:    endPos,
	}
}

func (p *Parser) parseExceptHandler() *model.ExceptHandler {
	startPos := p.current().Pos
	p.expect(model.TK_Except)

	var typ model.Expr
	var name *model.Identifier
	isStar := false

	// Check for except* syntax
	if p.check(model.TK_Star) {
		p.advance()
		isStar = true
	}

	if !p.check(model.TK_Colon) {
		typ = p.parseExpression()

		if p.match(model.TK_As) {
			name = p.parseIdentifier()
		}
	} else if isStar {
		p.addError("except* requires an exception type")
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

	return &model.ExceptHandler{
		Type:     typ,
		Name:     name,
		Body:     body,
		IsStar:   isStar,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseWithStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_With)

	var items []*model.WithItem
	for {
		contextExpr := p.parseExpression()
		var optionalVar model.Expr
		if p.match(model.TK_As) {
			optionalVar = p.parseTargetList()
		}
		items = append(items, &model.WithItem{
			ContextExpr: contextExpr,
			OptionalVar: optionalVar,
		})

		if !p.match(model.TK_Comma) {
			break
		}
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

	return &model.With{
		Items:    items,
		Body:     body,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseAssertStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Assert)

	test := p.parseExpression()
	if test == nil {
		p.addError("expected expression after 'assert'")
		return nil
	}
	endPos := test.End()

	var msg model.Expr
	if p.match(model.TK_Comma) {
		msg = p.parseExpression()
		if msg != nil {
			endPos = msg.End()
		}
	}

	p.match(model.TK_Newline)

	return &model.Assert{
		Test:     test,
		Msg:      msg,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseDelStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Del)

	var targets []model.Expr
	for {
		target := p.parsePrimaryExpr()
		if target == nil {
			p.addError("expected expression after 'del'")
			break
		}
		targets = append(targets, target)
		if !p.match(model.TK_Comma) {
			break
		}
	}

	p.match(model.TK_Newline)

	return &model.Delete{
		Targets:  targets,
		StartPos: startPos,
	}
}

func (p *Parser) parseGlobalStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Global)

	var names []*model.Identifier
	for {
		names = append(names, p.parseIdentifier())
		if !p.match(model.TK_Comma) {
			break
		}
	}

	endPos := names[len(names)-1].End()
	p.match(model.TK_Newline)

	return &model.Global{
		Names:    names,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseNonlocalStmt() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Nonlocal)

	var names []*model.Identifier
	for {
		names = append(names, p.parseIdentifier())
		if !p.match(model.TK_Comma) {
			break
		}
	}

	endPos := names[len(names)-1].End()
	p.match(model.TK_Newline)

	return &model.Nonlocal{
		Names:    names,
		StartPos: startPos,
		EndPos:   endPos,
	}
}

func (p *Parser) parseTypeAlias() model.Stmt {
	startPos := p.current().Pos
	p.expect(model.TK_Type)

	name := p.parseIdentifier()

	var typeParams []*model.TypeParam
	if p.match(model.TK_LBracket) {
		for !p.check(model.TK_RBracket) && !p.isAtEnd() {
			paramStart := p.current().Pos
			paramName := p.parseIdentifier()
			var bound model.Expr
			if p.match(model.TK_Colon) {
				bound = p.parseExpression()
			}
			typeParams = append(typeParams, &model.TypeParam{
				Name:     paramName,
				Bound:    bound,
				StartPos: paramStart,
				EndPos:   p.current().Pos,
			})
			if !p.match(model.TK_Comma) {
				break
			}
		}
		p.expect(model.TK_RBracket)
	}

	p.expect(model.TK_Assign)
	value := p.parseExpression()

	p.match(model.TK_Newline)

	return &model.TypeAlias{
		Name:       name,
		TypeParams: typeParams,
		Value:      value,
		StartPos:   startPos,
		EndPos:     value.End(),
	}
}

func (p *Parser) parseDecorated() model.Stmt {
	var decorators []model.Expr

	for p.check(model.TK_At) {
		p.advance()
		decorators = append(decorators, p.parseExpression())
		p.expect(model.TK_Newline)
	}

	switch p.current().Kind {
	case model.TK_Def:
		stmt := p.parseFunctionDef(false).(*model.FunctionDef)
		stmt.Decorators = decorators
		return stmt
	case model.TK_Async:
		p.advance()
		if p.check(model.TK_Def) {
			stmt := p.parseFunctionDef(true).(*model.FunctionDef)
			stmt.Decorators = decorators
			return stmt
		}
		p.addError("expected 'def' after 'async' in decorated statement")
		return nil
	case model.TK_Class:
		stmt := p.parseClassDef().(*model.ClassDef)
		stmt.Decorators = decorators
		return stmt
	default:
		p.addError("expected function or class definition after decorator")
		return nil
	}
}

