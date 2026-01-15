package render

import (
	"fmt"
	"io"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/builtin"
	"github.com/vietmpl/vie/token"
	"github.com/vietmpl/vie/value"
)

type renderer struct {
	context map[string]value.Value
	w       io.Writer
}

func (r *renderer) renderBlocks(b []ast.Block) error {
	for _, block := range b {
		if err := r.renderBlock(block); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) renderBlock(b ast.Block) error {
	switch block := b.(type) {
	case *ast.TextBlock:
		_, _ = io.WriteString(r.w, block.Content)
		return nil

	case *ast.CommentBlock:
		// Comments do not produce output.
		return nil

	case *ast.DisplayBlock:
		displayValue, err := r.evalExpr(block.Value)
		if err != nil {
			return err
		}
		stringValue, err := expectValueType[value.String](displayValue)
		if err != nil {
			return err
		}
		io.WriteString(r.w, string(stringValue))
		return nil

	case *ast.IfBlock:
		for _, branch := range block.Branches {
			conditionValue, err := r.evalExpr(branch.Condition)
			if err != nil {
				return err
			}
			condition, err := expectValueType[value.Bool](conditionValue)
			if err != nil {
				return err
			}
			if condition {
				if err := r.renderBlocks(branch.Consequence); err != nil {
					return err
				}
				break
			}
		}
		if block.Alternative != nil {
			if err := r.renderBlocks(*block.Alternative); err != nil {
				return err
			}
		}
		return nil

	default:
		panic(fmt.Sprintf("unexpected ast.Block: %T", block))
	}
}

func (r renderer) evalExpr(e ast.Expr) (value.Value, error) {
	switch expr := e.(type) {
	case *ast.BasicLiteral:
		return value.FromBasicLit(expr), nil

	case *ast.Identifier:
		return r.context[expr.Value], nil

	case *ast.BinaryExpr:
		lOperand, err := r.evalExpr(expr.LOperand)
		if err != nil {
			return nil, err
		}
		rOperand, err := r.evalExpr(expr.ROperand)
		if err != nil {
			return nil, err
		}
		switch expr.Operator {
		case token.TILDE:
			lOperandValue, err := expectValueType[value.String](lOperand)
			if err != nil {
				return nil, err
			}
			rOperandValue, err := expectValueType[value.String](rOperand)
			if err != nil {
				return nil, err
			}
			return lOperandValue.Concat(rOperandValue), nil

		case token.KEYWORD_OR:
			lOperandValue, err := expectValueType[value.Bool](lOperand)
			if err != nil {
				return nil, err
			}
			rOperandValue, err := expectValueType[value.Bool](rOperand)
			if err != nil {
				return nil, err
			}
			return lOperandValue.Or(rOperandValue), nil

		case token.KEYWORD_AND:
			lOperandValue, err := expectValueType[value.Bool](lOperand)
			if err != nil {
				return nil, err
			}
			rOperandValue, err := expectValueType[value.Bool](rOperand)
			if err != nil {
				return nil, err
			}
			return lOperandValue.And(rOperandValue), nil

		case token.EQUAL_EQUAL:
			if lOperand == nil || rOperand == nil {
				return nil, nil
			}

			var err error
			switch lOperand.(type) {
			case value.Bool:
				_, err = expectValueType[value.Bool](rOperand)
			case value.String:
				_, err = expectValueType[value.String](rOperand)
			default:
				panic(fmt.Sprintf("unexpected value.Value: %T", lOperand))
			}
			if err != nil {
				return nil, err
			}
			return value.Bool(lOperand == rOperand), nil

		case token.BANG_EQUAL:
			if lOperand == nil || rOperand == nil {
				return nil, nil
			}

			var err error
			switch lOperand.(type) {
			case value.Bool:
				_, err = expectValueType[value.Bool](rOperand)
			case value.String:
				_, err = expectValueType[value.String](rOperand)
			default:
				panic(fmt.Sprintf("unexpected value.Value: %T", lOperand))
			}
			if err != nil {
				return nil, err
			}
			return value.Bool(lOperand != rOperand), nil

		default:
			panic(fmt.Sprintf("unexpected binary operator: %s", expr.Operator))
		}

	case *ast.UnaryExpr:
		operand, err := r.evalExpr(expr.Operand)
		if err != nil {
			return nil, err
		}

		switch expr.Operator {
		case token.BANG:
			operandBool, err := expectValueType[value.Bool](operand)
			if err != nil {
				return nil, err
			}
			return !operandBool, nil

		default:
			panic(fmt.Sprintf("unexpected unary operator: %s", expr.Operator))
		}

	case *ast.ParenExpr:
		return r.evalExpr(expr.Value)

	case *ast.CallExpr:
		argumentValues, err := r.evalExprList(expr.Arguments)
		if err != nil {
			return nil, err
		}
		return r.evalCall(expr.Function, argumentValues)

	case *ast.PipeExpr:
		argument, err := r.evalExpr(expr.Argument)
		if err != nil {
			return nil, err
		}
		return r.evalCall(expr.Function, []value.Value{argument})

	default:
		panic(fmt.Sprintf("unexpected ast.Expr: %T", e))
	}
}

func (r renderer) evalExprList(exprList []ast.Expr) ([]value.Value, error) {
	values := make([]value.Value, 0, len(exprList))
	for _, expr := range exprList {
		v, err := r.evalExpr(expr)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

func (r renderer) evalCall(functionIdentifier ast.Identifier, argumentValues []value.Value) (value.Value, error) {
	function, err := builtin.LookupFunction(functionIdentifier)
	if err != nil {
		return nil, err
	}

	if len(function.ArgTypes) != len(argumentValues) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d",
			function.Name, len(function.ArgTypes), len(argumentValues))
	}

	return function.Call(argumentValues)
}

func expectValueType[T value.Value](val value.Value) (valT T, err error) {
	if val == nil {
		return
	}
	valT, ok := val.(T)
	if !ok {
		err = fmt.Errorf("expected %s, found %s", valT.Type(), val.Type())
		return
	}
	return
}
