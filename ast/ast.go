package ast

type Pos struct {
	// TODO(skewb1k): storing the path in every Pos is wasteful.
	// Move path resolution to a higher layer to avoid per-position duplication.
	Path      string
	Line      uint // line position in a document (zero-based)
	Character uint // character offset on a line in a document (zero-based)
}

// Interfaces ------------------------------------

// Node is the base interface implemented by all AST nodes.
type Node interface {
	node()
	// TODO(skewb1k): implement Pos().
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
	// TODO(skewb1k): add End().
}

// File represents a complete parsed file.
type File struct {
	Stmts []Stmt
}

// Statements ----------------------------------------

type (
	// A BadStmt node is a placeholder for statements containing
	// syntax errors for which no correct statement nodes can be
	// created.
	BadStmt struct {
		From, To Pos // position range of bad statement
	}

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

func (*BadStmt) stmtNode()    {}
func (*Text) stmtNode()       {}
func (*Comment) stmtNode()    {}
func (*RenderStmt) stmtNode() {}
func (*IfStmt) stmtNode()     {}
func (*SwitchStmt) stmtNode() {}

// Clauses are part of a larger statement but does not implement [Stmt] itself.
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
	// A BadExpr node is a placeholder for an expression containing
	// syntax errors for which a correct expression node cannot be
	// created.
	BadExpr struct {
		From, To Pos // position range of bad expression
	}

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

func (*BadExpr) exprNode()    {}
func (*BasicLit) exprNode()   {}
func (*Ident) exprNode()      {}
func (*UnaryExpr) exprNode()  {}
func (*BinaryExpr) exprNode() {}
func (*ParenExpr) exprNode()  {}
func (*CallExpr) exprNode()   {}
func (*PipeExpr) exprNode()   {}

func (x *BadExpr) Pos() Pos    { return x.From }
func (x *BasicLit) Pos() Pos   { return x.ValuePos }
func (x *Ident) Pos() Pos      { return x.NamePos }
func (x *UnaryExpr) Pos() Pos  { return x.OpPos }
func (x *BinaryExpr) Pos() Pos { return x.X.Pos() }
func (x *ParenExpr) Pos() Pos  { return x.Lparen }
func (x *CallExpr) Pos() Pos   { return x.Func.Pos() }
func (x *PipeExpr) Pos() Pos   { return x.Arg.Pos() }

func (*BadExpr) node()    {}
func (*BadStmt) node()    {}
func (*Text) node()       {}
func (*Comment) node()    {}
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
