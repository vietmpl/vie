package ast

type Location struct {
	Line   uint
	Column uint
}

type Template struct {
	Blocks []Block
}

// Interfaces ------------------------------------

// Node is the base interface implemented by all AST nodes.
type Node interface {
	node()
	// TODO(skewb1k): add Start().
}

type Block interface {
	Node
	blockNode()
}

type Expr interface {
	Node
	Start() Location
	exprNode()
	// TODO(skewb1k): add End().
}

// Blocks ----------------------------------------

type (
	TextBlock struct {
		Content string
	}

	CommentBlock struct {
		Content string
	}

	DisplayBlock struct {
		Value Expr
	}

	IfBlock struct {
		Condition   Expr
		Consequence []Block
		ElseIfs     []ElseIfClause
		Else        *ElseClause
	}
)

func (*TextBlock) blockNode()    {}
func (*CommentBlock) blockNode() {}
func (*DisplayBlock) blockNode() {}
func (*IfBlock) blockNode()      {}

// Clauses are part of a larger block but does not implement [Block] itself.
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
)

// Expressions -----------------------------------

type BasicLitKind uint8

const (
	KindBool BasicLitKind = iota
	KindString
)

type (
	BasicLiteral struct {
		Start_ Location
		Kind   BasicLitKind
		Value  string
	}

	Identifier struct {
		Start_ Location
		Value  string
	}

	UnaryExpr struct {
		OperatorLocation Location
		Operator         UnaryOperator
		Operand          Expr
	}

	BinaryExpr struct {
		LOperand Expr
		Operator BinaryOperator
		ROperand Expr
	}

	ParenExpr struct {
		LparenLocation Location
		Value          Expr
	}

	CallExpr struct {
		Function  Identifier
		Arguments []Expr
	}

	PipeExpr struct {
		Argument Expr
		Function Identifier
	}
)

func (*BasicLiteral) exprNode() {}
func (*Identifier) exprNode()   {}
func (*UnaryExpr) exprNode()    {}
func (*BinaryExpr) exprNode()   {}
func (*ParenExpr) exprNode()    {}
func (*CallExpr) exprNode()     {}
func (*PipeExpr) exprNode()     {}

func (x *BasicLiteral) Start() Location { return x.Start_ }
func (x *Identifier) Start() Location   { return x.Start_ }
func (x *UnaryExpr) Start() Location    { return x.OperatorLocation }
func (x *BinaryExpr) Start() Location   { return x.LOperand.Start() }
func (x *ParenExpr) Start() Location    { return x.LparenLocation }
func (x *CallExpr) Start() Location     { return x.Function.Start() }
func (x *PipeExpr) Start() Location     { return x.Argument.Start() }

func (*TextBlock) node()    {}
func (*CommentBlock) node() {}
func (*DisplayBlock) node() {}
func (*IfBlock) node()      {}
func (*BasicLiteral) node() {}
func (*Identifier) node()   {}
func (*UnaryExpr) node()    {}
func (*BinaryExpr) node()   {}
func (*ParenExpr) node()    {}
func (*CallExpr) node()     {}
func (*PipeExpr) node()     {}
