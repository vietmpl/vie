package ast

type Pos struct {
	Line      uint // line position in a document (zero-based)
	Character uint // character offset on a line in a document (zero-based)
}

// Interfaces ------------------------------------

// Node is the base interface implemented by all AST nodes.
type Node interface {
	node()
	// TODO: implement Pos().
}

// Stmt is any top-level block node in the source file.
type Stmt interface {
	Node
	stmtNode()
}

// Expr is any expression node.
type Expr interface {
	Node
	Pos() Pos // position of first character belonging to the expression
	exprNode()
	// TODO: add End().
}

// File represents a complete parsed file.
type File struct {
	Stmts []Stmt
}

// Statements ----------------------------------------

type (
	Text struct {
		Value string
	}

	Comment struct {
		Content string
	}

	RenderStmt struct {
		X Expr // expression
	}

	IfStmt struct {
		Cond    Expr
		Cons    []Stmt
		ElseIfs []ElseIfClause
		Else    *ElseClause
	}

	SwitchStmt struct {
		Value Expr
		Cases []CaseClause
	}
)

func (*Text) stmtNode()       {}
func (*RenderStmt) stmtNode() {}
func (*IfStmt) stmtNode()     {}
func (*SwitchStmt) stmtNode() {}

// Clauses Part of a larger statement but does not implement [Stmt] itself.
type (
	// ElseIfClause represents an `else if` branch inside an [IfStmt].
	ElseIfClause struct {
		Cond Expr
		Cons []Stmt
	}

	// ElseClause represents a final `else` branch inside an [IfStmt].
	ElseClause struct {
		Cons []Stmt
	}

	// CaseClause represents a single `case` branch inside a [SwitchStmt].
	CaseClause struct {
		List []Expr
		Body []Stmt
	}
)

// Expressions -----------------------------------

type BasicLitKind uint8

const (
	KindBool BasicLitKind = iota
	KindString
)

type (
	BasicLit struct {
		ValuePos Pos
		Kind     BasicLitKind
		Value    string
	}

	Ident struct {
		NamePos Pos    // identifier position
		Name    string // identifier name
	}

	UnaryExpr struct {
		OpPos Pos      // position of Op
		Op    UnOpKind // operator
		X     Expr     // operand
	}

	BinaryExpr struct {
		X  Expr      // left operand
		Op BinOpKind // operator
		Y  Expr      // right operand
	}

	ParenExpr struct {
		Lparen Pos  // position of "("
		X      Expr // parenthesized expression
	}

	CallExpr struct {
		Func Ident  // function name
		Args []Expr // function arguments
	}

	PipeExpr struct {
		Arg  Expr
		Func Ident
	}
)

func (*BasicLit) exprNode()   {}
func (*Ident) exprNode()      {}
func (*UnaryExpr) exprNode()  {}
func (*BinaryExpr) exprNode() {}
func (*ParenExpr) exprNode()  {}
func (*CallExpr) exprNode()   {}
func (*PipeExpr) exprNode()   {}

func (s *BasicLit) Pos() Pos   { return s.ValuePos }
func (s *Ident) Pos() Pos      { return s.NamePos }
func (s *UnaryExpr) Pos() Pos  { return s.OpPos }
func (s *BinaryExpr) Pos() Pos { return s.X.Pos() }
func (s *ParenExpr) Pos() Pos  { return s.Lparen }
func (s *CallExpr) Pos() Pos   { return s.Func.Pos() }
func (s *PipeExpr) Pos() Pos   { return s.Arg.Pos() }

func (*Text) node()       {}
func (*RenderStmt) node() {}
func (*IfStmt) node()     {}
func (*SwitchStmt) node() {}
func (*BasicLit) node()   {}
func (*Ident) node()      {}
func (*UnaryExpr) node()  {}
func (*BinaryExpr) node() {}
func (*ParenExpr) node()  {}
func (*CallExpr) node()   {}
func (*PipeExpr) node()   {}
