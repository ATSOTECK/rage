package model

// Node is the base interface for all AST nodes.
type Node interface {
	Pos() Position
	End() Position
}

// Expr is the interface for all expression nodes.
type Expr interface {
	Node
	exprNode()
}

// Stmt is the interface for all statement nodes.
type Stmt interface {
	Node
	stmtNode()
}

// Module represents a complete Python module (file).
type Module struct {
	Body     []Stmt
	StartPos Position
	EndPos   Position
}

func (m *Module) Pos() Position { return m.StartPos }
func (m *Module) End() Position { return m.EndPos }

// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------

// Identifier represents a name/identifier.
type Identifier struct {
	Name     string
	StartPos Position
	EndPos   Position
}

func (i *Identifier) Pos() Position { return i.StartPos }
func (i *Identifier) End() Position { return i.EndPos }
func (i *Identifier) exprNode()     {}

// IntLit represents an integer literal.
type IntLit struct {
	Value    string
	StartPos Position
	EndPos   Position
}

func (i *IntLit) Pos() Position { return i.StartPos }
func (i *IntLit) End() Position { return i.EndPos }
func (i *IntLit) exprNode()     {}

// FloatLit represents a floating-point literal.
type FloatLit struct {
	Value    string
	StartPos Position
	EndPos   Position
}

func (f *FloatLit) Pos() Position { return f.StartPos }
func (f *FloatLit) End() Position { return f.EndPos }
func (f *FloatLit) exprNode()     {}

// ImaginaryLit represents an imaginary number literal.
type ImaginaryLit struct {
	Value    string
	StartPos Position
	EndPos   Position
}

func (i *ImaginaryLit) Pos() Position { return i.StartPos }
func (i *ImaginaryLit) End() Position { return i.EndPos }
func (i *ImaginaryLit) exprNode()     {}

// StringLit represents a string literal.
type StringLit struct {
	Value    string
	StartPos Position
	EndPos   Position
}

func (s *StringLit) Pos() Position { return s.StartPos }
func (s *StringLit) End() Position { return s.EndPos }
func (s *StringLit) exprNode()     {}

// BytesLit represents a bytes literal.
type BytesLit struct {
	Value    string
	StartPos Position
	EndPos   Position
}

func (b *BytesLit) Pos() Position { return b.StartPos }
func (b *BytesLit) End() Position { return b.EndPos }
func (b *BytesLit) exprNode()     {}

// BoolLit represents True or False.
type BoolLit struct {
	Value    bool
	StartPos Position
	EndPos   Position
}

func (b *BoolLit) Pos() Position { return b.StartPos }
func (b *BoolLit) End() Position { return b.EndPos }
func (b *BoolLit) exprNode()     {}

// NoneLit represents None.
type NoneLit struct {
	StartPos Position
	EndPos   Position
}

func (n *NoneLit) Pos() Position { return n.StartPos }
func (n *NoneLit) End() Position { return n.EndPos }
func (n *NoneLit) exprNode()     {}

// Ellipsis represents the ... literal.
type Ellipsis struct {
	StartPos Position
	EndPos   Position
}

func (e *Ellipsis) Pos() Position { return e.StartPos }
func (e *Ellipsis) End() Position { return e.EndPos }
func (e *Ellipsis) exprNode()     {}

// UnaryOp represents a unary operation.
type UnaryOp struct {
	Op       TokenKind
	Operand  Expr
	StartPos Position
}

func (u *UnaryOp) Pos() Position { return u.StartPos }
func (u *UnaryOp) End() Position { return u.Operand.End() }
func (u *UnaryOp) exprNode()     {}

// BinaryOp represents a binary operation.
type BinaryOp struct {
	Left  Expr
	Op    TokenKind
	Right Expr
}

func (b *BinaryOp) Pos() Position { return b.Left.Pos() }
func (b *BinaryOp) End() Position { return b.Right.End() }
func (b *BinaryOp) exprNode()     {}

// BoolOp represents 'and' / 'or' with multiple values.
type BoolOp struct {
	Op     TokenKind // TK_And or TK_Or
	Values []Expr
}

func (b *BoolOp) Pos() Position { return b.Values[0].Pos() }
func (b *BoolOp) End() Position { return b.Values[len(b.Values)-1].End() }
func (b *BoolOp) exprNode()     {}

// Compare represents a comparison (can be chained: a < b < c).
type Compare struct {
	Left        Expr
	Ops         []TokenKind
	Comparators []Expr
}

func (c *Compare) Pos() Position { return c.Left.Pos() }
func (c *Compare) End() Position { return c.Comparators[len(c.Comparators)-1].End() }
func (c *Compare) exprNode()     {}

// Call represents a function call.
type Call struct {
	Func     Expr
	Args     []Expr
	Keywords []*Keyword
	EndPos   Position
}

func (c *Call) Pos() Position { return c.Func.Pos() }
func (c *Call) End() Position { return c.EndPos }
func (c *Call) exprNode()     {}

// Keyword represents a keyword argument in a call.
type Keyword struct {
	Arg      *Identifier // nil for **kwargs
	Value    Expr
	StartPos Position
}

func (k *Keyword) Pos() Position { return k.StartPos }
func (k *Keyword) End() Position { return k.Value.End() }

// Attribute represents attribute access (obj.attr).
type Attribute struct {
	Value Expr
	Attr  *Identifier
}

func (a *Attribute) Pos() Position { return a.Value.Pos() }
func (a *Attribute) End() Position { return a.Attr.End() }
func (a *Attribute) exprNode()     {}

// Subscript represents subscription (obj[index]).
type Subscript struct {
	Value  Expr
	Slice  Expr
	EndPos Position
}

func (s *Subscript) Pos() Position { return s.Value.Pos() }
func (s *Subscript) End() Position { return s.EndPos }
func (s *Subscript) exprNode()     {}

// Slice represents a slice (lower:upper:step).
type Slice struct {
	Lower    Expr // can be nil
	Upper    Expr // can be nil
	Step     Expr // can be nil
	StartPos Position
	EndPos   Position
}

func (s *Slice) Pos() Position { return s.StartPos }
func (s *Slice) End() Position { return s.EndPos }
func (s *Slice) exprNode()     {}

// List represents a list literal [a, b, c].
type List struct {
	Elts     []Expr
	StartPos Position
	EndPos   Position
}

func (l *List) Pos() Position { return l.StartPos }
func (l *List) End() Position { return l.EndPos }
func (l *List) exprNode()     {}

// Tuple represents a tuple (a, b, c).
type Tuple struct {
	Elts     []Expr
	StartPos Position
	EndPos   Position
}

func (t *Tuple) Pos() Position { return t.StartPos }
func (t *Tuple) End() Position { return t.EndPos }
func (t *Tuple) exprNode()     {}

// Dict represents a dictionary {k: v, ...}.
type Dict struct {
	Keys     []Expr // nil key means **unpacking
	Values   []Expr
	StartPos Position
	EndPos   Position
}

func (d *Dict) Pos() Position { return d.StartPos }
func (d *Dict) End() Position { return d.EndPos }
func (d *Dict) exprNode()     {}

// Set represents a set literal {a, b, c}.
type Set struct {
	Elts     []Expr
	StartPos Position
	EndPos   Position
}

func (s *Set) Pos() Position { return s.StartPos }
func (s *Set) End() Position { return s.EndPos }
func (s *Set) exprNode()     {}

// ListComp represents a list comprehension [x for x in xs].
type ListComp struct {
	Elt        Expr
	Generators []*Comprehension
	StartPos   Position
	EndPos     Position
}

func (l *ListComp) Pos() Position { return l.StartPos }
func (l *ListComp) End() Position { return l.EndPos }
func (l *ListComp) exprNode()     {}

// SetComp represents a set comprehension {x for x in xs}.
type SetComp struct {
	Elt        Expr
	Generators []*Comprehension
	StartPos   Position
	EndPos     Position
}

func (s *SetComp) Pos() Position { return s.StartPos }
func (s *SetComp) End() Position { return s.EndPos }
func (s *SetComp) exprNode()     {}

// DictComp represents a dict comprehension {k: v for k, v in items}.
type DictComp struct {
	Key        Expr
	Value      Expr
	Generators []*Comprehension
	StartPos   Position
	EndPos     Position
}

func (d *DictComp) Pos() Position { return d.StartPos }
func (d *DictComp) End() Position { return d.EndPos }
func (d *DictComp) exprNode()     {}

// GeneratorExpr represents a generator expression (x for x in xs).
type GeneratorExpr struct {
	Elt        Expr
	Generators []*Comprehension
	StartPos   Position
	EndPos     Position
}

func (g *GeneratorExpr) Pos() Position { return g.StartPos }
func (g *GeneratorExpr) End() Position { return g.EndPos }
func (g *GeneratorExpr) exprNode()     {}

// Comprehension represents a single for clause in a comprehension.
type Comprehension struct {
	Target  Expr
	Iter    Expr
	Ifs     []Expr
	IsAsync bool
}

// IfExpr represents a conditional expression (a if cond else b).
type IfExpr struct {
	Test   Expr
	Body   Expr
	OrElse Expr
}

func (i *IfExpr) Pos() Position { return i.Body.Pos() }
func (i *IfExpr) End() Position { return i.OrElse.End() }
func (i *IfExpr) exprNode()     {}

// Lambda represents a lambda expression.
type Lambda struct {
	Args     *Arguments
	Body     Expr
	StartPos Position
}

func (l *Lambda) Pos() Position { return l.StartPos }
func (l *Lambda) End() Position { return l.Body.End() }
func (l *Lambda) exprNode()     {}

// Yield represents a yield expression.
type Yield struct {
	Value    Expr // can be nil
	StartPos Position
	EndPos   Position
}

func (y *Yield) Pos() Position { return y.StartPos }
func (y *Yield) End() Position { return y.EndPos }
func (y *Yield) exprNode()     {}

// YieldFrom represents a yield from expression.
type YieldFrom struct {
	Value    Expr
	StartPos Position
}

func (y *YieldFrom) Pos() Position { return y.StartPos }
func (y *YieldFrom) End() Position { return y.Value.End() }
func (y *YieldFrom) exprNode()     {}

// Await represents an await expression.
type Await struct {
	Value    Expr
	StartPos Position
}

func (a *Await) Pos() Position { return a.StartPos }
func (a *Await) End() Position { return a.Value.End() }
func (a *Await) exprNode()     {}

// Starred represents a starred expression (*x).
type Starred struct {
	Value    Expr
	StartPos Position
}

func (s *Starred) Pos() Position { return s.StartPos }
func (s *Starred) End() Position { return s.Value.End() }
func (s *Starred) exprNode()     {}

// NamedExpr represents a named expression (walrus operator x := value).
type NamedExpr struct {
	Target *Identifier
	Value  Expr
}

func (n *NamedExpr) Pos() Position { return n.Target.Pos() }
func (n *NamedExpr) End() Position { return n.Value.End() }
func (n *NamedExpr) exprNode()     {}

// ----------------------------------------------------------------------------
// Statements
// ----------------------------------------------------------------------------

// ExprStmt represents an expression statement.
type ExprStmt struct {
	Value Expr
}

func (e *ExprStmt) Pos() Position { return e.Value.Pos() }
func (e *ExprStmt) End() Position { return e.Value.End() }
func (e *ExprStmt) stmtNode()     {}

// Assign represents an assignment statement.
type Assign struct {
	Targets []Expr
	Value   Expr
}

func (a *Assign) Pos() Position { return a.Targets[0].Pos() }
func (a *Assign) End() Position { return a.Value.End() }
func (a *Assign) stmtNode()     {}

// AnnAssign represents an annotated assignment.
type AnnAssign struct {
	Target     Expr
	Annotation Expr
	Value      Expr // can be nil
	Simple     bool
	StartPos   Position
}

func (a *AnnAssign) Pos() Position { return a.StartPos }
func (a *AnnAssign) End() Position {
	if a.Value != nil {
		return a.Value.End()
	}
	return a.Annotation.End()
}
func (a *AnnAssign) stmtNode() {}

// AugAssign represents an augmented assignment (+=, -=, etc.).
type AugAssign struct {
	Target Expr
	Op     TokenKind
	Value  Expr
}

func (a *AugAssign) Pos() Position { return a.Target.Pos() }
func (a *AugAssign) End() Position { return a.Value.End() }
func (a *AugAssign) stmtNode()     {}

// Pass represents a pass statement.
type Pass struct {
	StartPos Position
	EndPos   Position
}

func (p *Pass) Pos() Position { return p.StartPos }
func (p *Pass) End() Position { return p.EndPos }
func (p *Pass) stmtNode()     {}

// Break represents a break statement.
type Break struct {
	StartPos Position
	EndPos   Position
}

func (b *Break) Pos() Position { return b.StartPos }
func (b *Break) End() Position { return b.EndPos }
func (b *Break) stmtNode()     {}

// Continue represents a continue statement.
type Continue struct {
	StartPos Position
	EndPos   Position
}

func (c *Continue) Pos() Position { return c.StartPos }
func (c *Continue) End() Position { return c.EndPos }
func (c *Continue) stmtNode()     {}

// Return represents a return statement.
type Return struct {
	Value    Expr // can be nil
	StartPos Position
	EndPos   Position
}

func (r *Return) Pos() Position { return r.StartPos }
func (r *Return) End() Position { return r.EndPos }
func (r *Return) stmtNode()     {}

// Raise represents a raise statement.
type Raise struct {
	Exc      Expr // can be nil
	Cause    Expr // can be nil
	StartPos Position
	EndPos   Position
}

func (r *Raise) Pos() Position { return r.StartPos }
func (r *Raise) End() Position { return r.EndPos }
func (r *Raise) stmtNode()     {}

// Delete represents a del statement.
type Delete struct {
	Targets  []Expr
	StartPos Position
}

func (d *Delete) Pos() Position { return d.StartPos }
func (d *Delete) End() Position { return d.Targets[len(d.Targets)-1].End() }
func (d *Delete) stmtNode()     {}

// Assert represents an assert statement.
type Assert struct {
	Test     Expr
	Msg      Expr // can be nil
	StartPos Position
	EndPos   Position
}

func (a *Assert) Pos() Position { return a.StartPos }
func (a *Assert) End() Position { return a.EndPos }
func (a *Assert) stmtNode()     {}

// Global represents a global statement.
type Global struct {
	Names    []*Identifier
	StartPos Position
	EndPos   Position
}

func (g *Global) Pos() Position { return g.StartPos }
func (g *Global) End() Position { return g.EndPos }
func (g *Global) stmtNode()     {}

// Nonlocal represents a nonlocal statement.
type Nonlocal struct {
	Names    []*Identifier
	StartPos Position
	EndPos   Position
}

func (n *Nonlocal) Pos() Position { return n.StartPos }
func (n *Nonlocal) End() Position { return n.EndPos }
func (n *Nonlocal) stmtNode()     {}

// If represents an if statement.
type If struct {
	Test     Expr
	Body     []Stmt
	OrElse   []Stmt // can include elif (nested If) or else body
	StartPos Position
	EndPos   Position
}

func (i *If) Pos() Position { return i.StartPos }
func (i *If) End() Position { return i.EndPos }
func (i *If) stmtNode()     {}

// While represents a while statement.
type While struct {
	Test     Expr
	Body     []Stmt
	OrElse   []Stmt
	StartPos Position
	EndPos   Position
}

func (w *While) Pos() Position { return w.StartPos }
func (w *While) End() Position { return w.EndPos }
func (w *While) stmtNode()     {}

// For represents a for statement.
type For struct {
	Target   Expr
	Iter     Expr
	Body     []Stmt
	OrElse   []Stmt
	IsAsync  bool
	StartPos Position
	EndPos   Position
}

func (f *For) Pos() Position { return f.StartPos }
func (f *For) End() Position { return f.EndPos }
func (f *For) stmtNode()     {}

// With represents a with statement.
type With struct {
	Items    []*WithItem
	Body     []Stmt
	IsAsync  bool
	StartPos Position
	EndPos   Position
}

func (w *With) Pos() Position { return w.StartPos }
func (w *With) End() Position { return w.EndPos }
func (w *With) stmtNode()     {}

// WithItem represents a single item in a with statement.
type WithItem struct {
	ContextExpr Expr
	OptionalVar Expr // can be nil
}

// Try represents a try statement.
type Try struct {
	Body      []Stmt
	Handlers  []*ExceptHandler
	OrElse    []Stmt
	FinalBody []Stmt
	StartPos  Position
	EndPos    Position
}

func (t *Try) Pos() Position { return t.StartPos }
func (t *Try) End() Position { return t.EndPos }
func (t *Try) stmtNode()     {}

// ExceptHandler represents an except clause.
type ExceptHandler struct {
	Type     Expr        // can be nil for bare except
	Name     *Identifier // can be nil
	Body     []Stmt
	StartPos Position
	EndPos   Position
}

func (e *ExceptHandler) Pos() Position { return e.StartPos }
func (e *ExceptHandler) End() Position { return e.EndPos }

// FunctionDef represents a function definition.
type FunctionDef struct {
	Name       *Identifier
	Args       *Arguments
	Body       []Stmt
	Decorators []Expr
	Returns    Expr // return type annotation, can be nil
	IsAsync    bool
	StartPos   Position
	EndPos     Position
}

func (f *FunctionDef) Pos() Position { return f.StartPos }
func (f *FunctionDef) End() Position { return f.EndPos }
func (f *FunctionDef) stmtNode()     {}

// Arguments represents function arguments.
type Arguments struct {
	PosOnlyArgs []*Arg
	Args        []*Arg
	VarArg      *Arg // *args
	KwOnlyArgs  []*Arg
	KwDefaults  []Expr // defaults for kwonly args
	KwArg       *Arg   // **kwargs
	Defaults    []Expr // defaults for positional args
}

// Arg represents a single argument.
type Arg struct {
	Arg        *Identifier
	Annotation Expr // can be nil
	StartPos   Position
	EndPos     Position
}

func (a *Arg) Pos() Position { return a.StartPos }
func (a *Arg) End() Position { return a.EndPos }

// ClassDef represents a class definition.
type ClassDef struct {
	Name       *Identifier
	Bases      []Expr
	Keywords   []*Keyword
	Body       []Stmt
	Decorators []Expr
	StartPos   Position
	EndPos     Position
}

func (c *ClassDef) Pos() Position { return c.StartPos }
func (c *ClassDef) End() Position { return c.EndPos }
func (c *ClassDef) stmtNode()     {}

// Import represents an import statement.
type Import struct {
	Names    []*Alias
	StartPos Position
	EndPos   Position
}

func (i *Import) Pos() Position { return i.StartPos }
func (i *Import) End() Position { return i.EndPos }
func (i *Import) stmtNode()     {}

// ImportFrom represents a from ... import statement.
type ImportFrom struct {
	Module   *Identifier // can be nil for relative imports
	Names    []*Alias
	Level    int // number of dots for relative import
	StartPos Position
	EndPos   Position
}

func (i *ImportFrom) Pos() Position { return i.StartPos }
func (i *ImportFrom) End() Position { return i.EndPos }
func (i *ImportFrom) stmtNode()     {}

// Alias represents an import alias.
type Alias struct {
	Name     *Identifier
	AsName   *Identifier // can be nil
	StartPos Position
	EndPos   Position
}

func (a *Alias) Pos() Position { return a.StartPos }
func (a *Alias) End() Position { return a.EndPos }

// Match represents a match statement.
type Match struct {
	Subject  Expr
	Cases    []*MatchCase
	StartPos Position
	EndPos   Position
}

func (m *Match) Pos() Position { return m.StartPos }
func (m *Match) End() Position { return m.EndPos }
func (m *Match) stmtNode()     {}

// MatchCase represents a case in a match statement.
type MatchCase struct {
	Pattern  Pattern
	Guard    Expr // can be nil
	Body     []Stmt
	StartPos Position
	EndPos   Position
}

func (m *MatchCase) Pos() Position { return m.StartPos }
func (m *MatchCase) End() Position { return m.EndPos }

// Pattern is the interface for match patterns.
type Pattern interface {
	Node
	patternNode()
}

// MatchValue represents a literal or constant pattern.
type MatchValue struct {
	Value    Expr
	StartPos Position
	EndPos   Position
}

func (m *MatchValue) Pos() Position { return m.StartPos }
func (m *MatchValue) End() Position { return m.EndPos }
func (m *MatchValue) patternNode()  {}

// MatchSingleton represents True, False, or None in patterns.
type MatchSingleton struct {
	Value    Expr
	StartPos Position
	EndPos   Position
}

func (m *MatchSingleton) Pos() Position { return m.StartPos }
func (m *MatchSingleton) End() Position { return m.EndPos }
func (m *MatchSingleton) patternNode()  {}

// MatchSequence represents a sequence pattern [a, b, c].
type MatchSequence struct {
	Patterns []Pattern
	StartPos Position
	EndPos   Position
}

func (m *MatchSequence) Pos() Position { return m.StartPos }
func (m *MatchSequence) End() Position { return m.EndPos }
func (m *MatchSequence) patternNode()  {}

// MatchMapping represents a mapping pattern {k: v, ...}.
type MatchMapping struct {
	Keys     []Expr
	Patterns []Pattern
	Rest     *Identifier // **rest, can be nil
	StartPos Position
	EndPos   Position
}

func (m *MatchMapping) Pos() Position { return m.StartPos }
func (m *MatchMapping) End() Position { return m.EndPos }
func (m *MatchMapping) patternNode()  {}

// MatchClass represents a class pattern Cls(a, b=c).
type MatchClass struct {
	Cls         Expr
	Patterns    []Pattern
	KwdAttrs    []*Identifier
	KwdPatterns []Pattern
	StartPos    Position
	EndPos      Position
}

func (m *MatchClass) Pos() Position { return m.StartPos }
func (m *MatchClass) End() Position { return m.EndPos }
func (m *MatchClass) patternNode()  {}

// MatchStar represents a starred pattern *rest.
type MatchStar struct {
	Name     *Identifier // can be nil for _
	StartPos Position
	EndPos   Position
}

func (m *MatchStar) Pos() Position { return m.StartPos }
func (m *MatchStar) End() Position { return m.EndPos }
func (m *MatchStar) patternNode()  {}

// MatchAs represents an as pattern or wildcard.
type MatchAs struct {
	Pattern  Pattern     // nil for wildcard _
	Name     *Identifier // nil for wildcard _
	StartPos Position
	EndPos   Position
}

func (m *MatchAs) Pos() Position { return m.StartPos }
func (m *MatchAs) End() Position { return m.EndPos }
func (m *MatchAs) patternNode()  {}

// MatchOr represents an or pattern a | b | c.
type MatchOr struct {
	Patterns []Pattern
	StartPos Position
	EndPos   Position
}

func (m *MatchOr) Pos() Position { return m.StartPos }
func (m *MatchOr) End() Position { return m.EndPos }
func (m *MatchOr) patternNode()  {}

// TypeAlias represents a type alias statement (type X = ...).
type TypeAlias struct {
	Name       *Identifier
	TypeParams []*TypeParam
	Value      Expr
	StartPos   Position
	EndPos     Position
}

func (t *TypeAlias) Pos() Position { return t.StartPos }
func (t *TypeAlias) End() Position { return t.EndPos }
func (t *TypeAlias) stmtNode()     {}

// TypeParam represents a type parameter.
type TypeParam struct {
	Name     *Identifier
	Bound    Expr // can be nil
	StartPos Position
	EndPos   Position
}

func (t *TypeParam) Pos() Position { return t.StartPos }
func (t *TypeParam) End() Position { return t.EndPos }
