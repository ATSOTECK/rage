package compiler

import (
	"fmt"

	"github.com/ATSOTECK/RAGE/internal/model"
)

// Precedence levels for Pratt parsing.
const (
	precNone       = 0 // No precedence (used for non-infix tokens)
	precLowest     = 1
	precWalrus     = 2  // :=
	precLambda     = 3  // lambda
	precTernary    = 4  // if else
	precOr         = 5  // or
	precAnd        = 6  // and
	precNot        = 7  // not
	precComparison = 8  // <, >, ==, !=, <=, >=, in, is, not in, is not
	precBitOr      = 9  // |
	precBitXor     = 10 // ^
	precBitAnd     = 11 // &
	precShift      = 12 // <<, >>
	precAddSub     = 13 // +, -
	precMulDiv     = 14 // *, /, //, %, @
	precUnary      = 15 // +x, -x, ~x
	precPower      = 16 // **
	precAwait      = 17 // await
	precCall       = 18 // x(), x[], x.attr
)

// ParseError represents a parsing error.
type ParseError struct {
	Pos     model.Position
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// Parser parses Python source code into an AST.
type Parser struct {
	lexer  *Lexer
	tokens []model.Token
	pos    int
	errors []ParseError
}

// NewParser creates a new parser for the given source.
func NewParser(source string) *Parser {
	return NewParserWithFilename(source, "")
}

// NewParserWithFilename creates a new parser with a filename for error messages.
func NewParserWithFilename(source, filename string) *Parser {
	lexer := NewLexerWithFilename(source, filename)
	tokens, _ := lexer.Tokenize()
	return &Parser{
		lexer:  lexer,
		tokens: tokens,
		pos:    0,
	}
}

// Parse parses the source and returns the AST module.
func (p *Parser) Parse() (*model.Module, []ParseError) {
	module := &model.Module{}

	if len(p.tokens) > 0 {
		module.StartPos = p.tokens[0].Pos
	}

	for !p.isAtEnd() {
		if p.check(model.TK_Newline) || p.check(model.TK_Comment) {
			p.advance()
			continue
		}
		if p.check(model.TK_EOF) {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			module.Body = append(module.Body, stmt)
		}
	}

	if len(p.tokens) > 0 {
		module.EndPos = p.tokens[len(p.tokens)-1].Pos
	}

	return module, p.errors
}

// Errors returns the parsing errors.
func (p *Parser) Errors() []ParseError {
	return p.errors
}

// Helper methods

func (p *Parser) current() model.Token {
	if p.pos >= len(p.tokens) {
		return model.Token{Kind: model.TK_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peek() model.Token {
	if p.pos+1 >= len(p.tokens) {
		return model.Token{Kind: model.TK_EOF}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) advance() model.Token {
	tok := p.current()
	if !p.isAtEnd() {
		p.pos++
	}
	return tok
}

func (p *Parser) isAtEnd() bool {
	return p.current().Kind == model.TK_EOF
}

func (p *Parser) check(kind model.TokenKind) bool {
	return p.current().Kind == kind
}

func (p *Parser) match(kinds ...model.TokenKind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) expect(kind model.TokenKind) model.Token {
	if p.check(kind) {
		return p.advance()
	}
	p.addError(fmt.Sprintf("expected %s, got %s", kind, p.current().Kind))
	return p.current()
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, ParseError{
		Pos:     p.current().Pos,
		Message: msg,
	})
}

func (p *Parser) skipNewlines() {
	for p.check(model.TK_Newline) || p.check(model.TK_Comment) {
		p.advance()
	}
}

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
	case model.TK_Type:
		return p.parseTypeAlias()
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
			elts = append(elts, p.parseExpression())
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
		next := p.parseExpression()
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

	if p.check(model.TK_Comma) {
		elts := []model.Expr{first}
		startPos := first.Pos()
		for p.match(model.TK_Comma) {
			if p.check(model.TK_In) {
				break
			}
			elts = append(elts, p.parsePrimaryExpr())
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
		value = p.parseExpression()
		endPos = value.End()
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

	if !p.check(model.TK_Newline) && !p.check(model.TK_EOF) {
		exc = p.parseExpression()
		endPos = exc.End()

		if p.match(model.TK_From) {
			cause = p.parseExpression()
			endPos = cause.End()
		}
	}

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

	if !p.check(model.TK_Colon) {
		typ = p.parseExpression()

		if p.match(model.TK_As) {
			name = p.parseIdentifier()
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

	return &model.ExceptHandler{
		Type:     typ,
		Name:     name,
		Body:     body,
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
	endPos := test.End()

	var msg model.Expr
	if p.match(model.TK_Comma) {
		msg = p.parseExpression()
		endPos = msg.End()
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
		targets = append(targets, p.parsePrimaryExpr())
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
	// Simplified pattern parsing - handles basic patterns
	startPos := p.current().Pos

	if p.check(model.TK_Identifier) {
		name := p.parseIdentifier()
		if name.Name == "_" {
			return &model.MatchAs{
				StartPos: startPos,
				EndPos:   name.End(),
			}
		}
		if p.match(model.TK_As) {
			asName := p.parseIdentifier()
			return &model.MatchAs{
				Pattern: &model.MatchValue{
					Value:    name,
					StartPos: startPos,
					EndPos:   name.End(),
				},
				Name:     asName,
				StartPos: startPos,
				EndPos:   asName.End(),
			}
		}
		return &model.MatchAs{
			Name:     name,
			StartPos: startPos,
			EndPos:   name.End(),
		}
	}

	// Literal patterns
	expr := p.parsePrimaryExpr()
	return &model.MatchValue{
		Value:    expr,
		StartPos: startPos,
		EndPos:   expr.End(),
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

func (p *Parser) parseBlock() []model.Stmt {
	var stmts []model.Stmt

	if !p.match(model.TK_Indent) {
		return stmts
	}

	for !p.check(model.TK_Dedent) && !p.isAtEnd() {
		if p.check(model.TK_Newline) || p.check(model.TK_Comment) {
			p.advance()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	p.match(model.TK_Dedent)
	return stmts
}

// Expression parsing using Pratt parser

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
		return &model.YieldFrom{
			Value:    value,
			StartPos: startPos,
		}
	}

	var value model.Expr
	endPos := startPos
	if !p.check(model.TK_Newline) && !p.check(model.TK_RParen) && !p.check(model.TK_EOF) {
		value = p.parseExpression()
		endPos = value.End()
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

func (p *Parser) parseIdentifier() *model.Identifier {
	tok := p.expect(model.TK_Identifier)
	return &model.Identifier{
		Name:     tok.Literal,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
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
	return &model.Starred{
		Value:    value,
		StartPos: startPos,
	}
}

func (p *Parser) parseCallExpr(fn model.Expr) model.Expr {
	p.expect(model.TK_LParen)

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
			args = append(args, p.parseExpression())
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
	p.expect(model.TK_Else)
	orElse := p.parsePrecedence(precTernary)

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

		ops = append(ops, op)
		comparators = append(comparators, p.parsePrecedence(precBitOr))
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
