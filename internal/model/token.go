package model

import "fmt"

// TokenKind represents the type of a Python token.
type TokenKind int

const (
	// Special tokens
	TK_Illegal TokenKind = iota
	TK_Newline
	TK_Indent
	TK_Dedent
	TK_Comment

	// Literals
	TK_Identifier
	TK_IntLit
	TK_FloatLit
	TK_ImaginaryLit
	TK_StringLit
	TK_BytesLit
	TK_FStringLit
	TK_FStringStart
	TK_FStringMiddle
	TK_FStringEnd

	// Keywords
	TK_False
	TK_None
	TK_True
	TK_And
	TK_As
	TK_Assert
	TK_Async
	TK_Await
	TK_Break
	TK_Case
	TK_Class
	TK_Continue
	TK_Def
	TK_Del
	TK_Elif
	TK_Else
	TK_Except
	TK_Finally
	TK_For
	TK_From
	TK_Global
	TK_If
	TK_Import
	TK_In
	TK_Is
	TK_Lambda
	TK_Match
	TK_Nonlocal
	TK_Not
	TK_Or
	TK_Pass
	TK_Raise
	TK_Return
	TK_Try
	TK_Type
	TK_While
	TK_With
	TK_Yield

	// Operators
	TK_Plus         // +
	TK_Minus        // -
	TK_Star         // *
	TK_DoubleStar   // **
	TK_Slash        // /
	TK_DoubleSlash  // //
	TK_Percent      // %
	TK_At           // @
	TK_LShift       // <<
	TK_RShift       // >>
	TK_Ampersand    // &
	TK_Pipe         // |
	TK_Caret        // ^
	TK_Tilde        // ~
	TK_Walrus       // :=
	TK_Less         // <
	TK_Greater      // >
	TK_LessEqual    // <=
	TK_GreaterEqual // >=
	TK_Equal        // ==
	TK_NotEqual     // !=
	TK_Arrow        // ->
	TK_NotIn        // not in (compound)
	TK_IsNot        // is not (compound)

	// Delimiters
	TK_LParen            // (
	TK_RParen            // )
	TK_LBracket          // [
	TK_RBracket          // ]
	TK_LBrace            // {
	TK_RBrace            // }
	TK_Comma             // ,
	TK_Colon             // :
	TK_Semicolon         // ;
	TK_Dot               // .
	TK_Ellipsis          // ...
	TK_Assign            // =
	TK_PlusAssign        // +=
	TK_MinusAssign       // -=
	TK_StarAssign        // *=
	TK_SlashAssign       // /=
	TK_DoubleSlashAssign // //=
	TK_PercentAssign     // %=
	TK_AtAssign          // @=
	TK_AmpersandAssign   // &=
	TK_PipeAssign        // |=
	TK_CaretAssign       // ^=
	TK_RShiftAssign      // >>=
	TK_LShiftAssign      // <<=
	TK_DoubleStarAssign  // **=

	TK_EOF
)

var tokenNames = map[TokenKind]string{
	TK_Illegal: "ILLEGAL",
	TK_EOF:     "EOF",
	TK_Newline: "NEWLINE",
	TK_Indent:  "INDENT",
	TK_Dedent:  "DEDENT",
	TK_Comment: "COMMENT",

	TK_Identifier:    "IDENTIFIER",
	TK_IntLit:        "INT",
	TK_FloatLit:      "FLOAT",
	TK_ImaginaryLit:  "IMAGINARY",
	TK_StringLit:     "STRING",
	TK_BytesLit:      "BYTES",
	TK_FStringLit:    "FSTRING",
	TK_FStringStart:  "FSTRING_START",
	TK_FStringMiddle: "FSTRING_MIDDLE",
	TK_FStringEnd:    "FSTRING_END",

	TK_False:    "False",
	TK_None:     "None",
	TK_True:     "True",
	TK_And:      "and",
	TK_As:       "as",
	TK_Assert:   "assert",
	TK_Async:    "async",
	TK_Await:    "await",
	TK_Break:    "break",
	TK_Case:     "case",
	TK_Class:    "class",
	TK_Continue: "continue",
	TK_Def:      "def",
	TK_Del:      "del",
	TK_Elif:     "elif",
	TK_Else:     "else",
	TK_Except:   "except",
	TK_Finally:  "finally",
	TK_For:      "for",
	TK_From:     "from",
	TK_Global:   "global",
	TK_If:       "if",
	TK_Import:   "import",
	TK_In:       "in",
	TK_Is:       "is",
	TK_Lambda:   "lambda",
	TK_Match:    "match",
	TK_Nonlocal: "nonlocal",
	TK_Not:      "not",
	TK_Or:       "or",
	TK_Pass:     "pass",
	TK_Raise:    "raise",
	TK_Return:   "return",
	TK_Try:      "try",
	TK_Type:     "type",
	TK_While:    "while",
	TK_With:     "with",
	TK_Yield:    "yield",

	TK_Plus:         "+",
	TK_Minus:        "-",
	TK_Star:         "*",
	TK_DoubleStar:   "**",
	TK_Slash:        "/",
	TK_DoubleSlash:  "//",
	TK_Percent:      "%",
	TK_At:           "@",
	TK_LShift:       "<<",
	TK_RShift:       ">>",
	TK_Ampersand:    "&",
	TK_Pipe:         "|",
	TK_Caret:        "^",
	TK_Tilde:        "~",
	TK_Walrus:       ":=",
	TK_Less:         "<",
	TK_Greater:      ">",
	TK_LessEqual:    "<=",
	TK_GreaterEqual: ">=",
	TK_Equal:        "==",
	TK_NotEqual:     "!=",
	TK_Arrow:        "->",

	TK_LParen:            "(",
	TK_RParen:            ")",
	TK_LBracket:          "[",
	TK_RBracket:          "]",
	TK_LBrace:            "{",
	TK_RBrace:            "}",
	TK_Comma:             ",",
	TK_Colon:             ":",
	TK_Semicolon:         ";",
	TK_Dot:               ".",
	TK_Ellipsis:          "...",
	TK_Assign:            "=",
	TK_PlusAssign:        "+=",
	TK_MinusAssign:       "-=",
	TK_StarAssign:        "*=",
	TK_SlashAssign:       "/=",
	TK_DoubleSlashAssign: "//=",
	TK_PercentAssign:     "%=",
	TK_AtAssign:          "@=",
	TK_AmpersandAssign:   "&=",
	TK_PipeAssign:        "|=",
	TK_CaretAssign:       "^=",
	TK_RShiftAssign:      ">>=",
	TK_LShiftAssign:      "<<=",
	TK_DoubleStarAssign:  "**=",
}

func (tk TokenKind) String() string {
	if name, ok := tokenNames[tk]; ok {
		return name
	}
	return fmt.Sprintf("TokenKind(%d)", tk)
}

// Keywords maps Python keywords to their token kinds.
var Keywords = map[string]TokenKind{
	"False":    TK_False,
	"None":     TK_None,
	"True":     TK_True,
	"and":      TK_And,
	"as":       TK_As,
	"assert":   TK_Assert,
	"async":    TK_Async,
	"await":    TK_Await,
	"break":    TK_Break,
	"case":     TK_Case,
	"class":    TK_Class,
	"continue": TK_Continue,
	"def":      TK_Def,
	"del":      TK_Del,
	"elif":     TK_Elif,
	"else":     TK_Else,
	"except":   TK_Except,
	"finally":  TK_Finally,
	"for":      TK_For,
	"from":     TK_From,
	"global":   TK_Global,
	"if":       TK_If,
	"import":   TK_Import,
	"in":       TK_In,
	"is":       TK_Is,
	"lambda":   TK_Lambda,
	"match":    TK_Match,
	"nonlocal": TK_Nonlocal,
	"not":      TK_Not,
	"or":       TK_Or,
	"pass":     TK_Pass,
	"raise":    TK_Raise,
	"return":   TK_Return,
	"try":      TK_Try,
	"type":     TK_Type,
	"while":    TK_While,
	"with":     TK_With,
	"yield":    TK_Yield,
}

// Position represents a position in the source code.
type Position struct {
	Filename string
	Line     int
	Column   int
	Offset   int
}

func (p Position) String() string {
	if p.Filename != "" {
		return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Token represents a lexical token.
type Token struct {
	Kind    TokenKind
	Literal string
	Pos     Position
	EndPos  Position
}

func (t Token) String() string {
	if t.Literal != "" {
		return fmt.Sprintf("%s(%q) at %s", t.Kind, t.Literal, t.Pos)
	}
	return fmt.Sprintf("%s at %s", t.Kind, t.Pos)
}

// LookupIdent checks if an identifier is a keyword.
func LookupIdent(ident string) TokenKind {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return TK_Identifier
}
