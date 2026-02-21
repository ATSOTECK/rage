package compiler

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/model"
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

func (p *Parser) parseIdentifier() *model.Identifier {
	tok := p.expect(model.TK_Identifier)
	return &model.Identifier{
		Name:     tok.Literal,
		StartPos: tok.Pos,
		EndPos:   tok.EndPos,
	}
}
