package ast

// Interfaces ------------------------------------

// Node is the base interface implemented by all AST nodes.
type Node interface {
	node()
}

// Stmt is any top-level block node in the source file.
type Stmt interface {
	Node
	stmtNode()
}

// Expr is any expression node.
type Expr interface {
	Node
	exprNode()
}

// SourceFile represents a complete parsed file.
type SourceFile struct {
	Stmts []Stmt
}

// Statements ----------------------------------------

type (
	Text struct {
		Value []byte
	}

	Comment struct {
		Value []byte
	}

	RenderStmt struct {
		Expr Expr
	}

	IfStmt struct {
		Condition   Expr
		Consequence []Stmt
		// ElseIfClause or ElseClause
		Alternative any
	}

	SwitchStmt struct {
		Value Expr
		Cases []*CaseClause
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
		Condition   Expr
		Consequence []Stmt
		// ElseIfClause or ElseClause
		Alternative any
	}

	// ElseClause represents a final `else` branch inside an [IfStmt].
	ElseClause struct {
		Consequence []Stmt
	}

	// CaseClause represents a single `case` branch inside a [SwitchStmt].
	CaseClause struct {
		Value Expr
		Body  []Stmt
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
		Kind  BasicLitKind
		Value []byte
	}

	Ident struct {
		Value []byte
	}

	UnaryExpr struct {
		Op   UnOpKind
		Expr Expr
	}

	BinaryExpr struct {
		Left  Expr
		Op    BinOpKind
		Right Expr
	}

	ParenExpr struct {
		Expr Expr
	}

	CallExpr struct {
		Func *Ident
		Args []Expr
	}

	PipeExpr struct {
		Arg  Expr
		Func *Ident
	}
)

func (*BasicLit) exprNode()   {}
func (*Ident) exprNode()      {}
func (*UnaryExpr) exprNode()  {}
func (*BinaryExpr) exprNode() {}
func (*ParenExpr) exprNode()  {}
func (*CallExpr) exprNode()   {}
func (*PipeExpr) exprNode()   {}

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
