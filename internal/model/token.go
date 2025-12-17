package model

type TokenKind int

const (
	TK_Identifier TokenKind = iota
	TK_NumberLit
)

type Token struct {
	Kind TokenKind
}
