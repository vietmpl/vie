package ast

// Interfaces ------------------------------------

// Node is the base interface implemented by all AST nodes.
type Node interface {
	node()
}

// Block is any top-level block node in the source file.
type Block interface {
	Node
	blockNode()
}

// Expr is any expression node.
type Expr interface {
	Node
	exprNode()
}

// SourceFile represents a complete parsed file.
type SourceFile struct {
	Blocks []Block
}

// Blocks ----------------------------------------

type (
	TextBlock struct {
		Value []byte
	}

	RenderBlock struct {
		Expr Expr
	}

	IfBlock struct {
		Condition   Expr
		Consequence []Block
		ElseIfs     []ElseIfClause
		Else        *ElseClause // may be nil
	}

	SwitchBlock struct {
		Value Expr
		Cases []CaseClause
	}
)

func (*TextBlock) blockNode()   {}
func (*RenderBlock) blockNode() {}
func (*IfBlock) blockNode()     {}
func (*SwitchBlock) blockNode() {}

// Clauses Part of a larger block but does not implement [Block] itself.
type (
	// ElseIfClause represents an `else if` branch inside an [IfBlock].
	ElseIfClause struct {
		Condition   Expr
		Consequence []Block
	}

	// ElseClause represents a final `else` branch inside an [IfBlock].
	ElseClause struct {
		Consequence []Block
	}

	// CaseClause represents a single `case` branch inside a [SwitchBlock].
	CaseClause struct {
		Value Expr
		Body  []Block
	}
)

// Expressions -----------------------------------

// BasicLitKind enumerates the possible types of a BasicLit.
type BasicLitKind int

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
		Op   []byte
		Expr Expr
	}

	BinaryExpr struct {
		Left  Expr
		Op    []byte
		Right Expr
	}

	ParenExpr struct {
		Expr Expr
	}

	CallExpr struct {
		Func Expr
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

func (*TextBlock) node()   {}
func (*RenderBlock) node() {}
func (*IfBlock) node()     {}
func (*SwitchBlock) node() {}
func (*BasicLit) node()    {}
func (*Ident) node()       {}
func (*UnaryExpr) node()   {}
func (*BinaryExpr) node()  {}
func (*ParenExpr) node()   {}
func (*CallExpr) node()    {}
func (*PipeExpr) node()    {}
