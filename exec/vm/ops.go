package vm

import (
	"fmt"
	"math"

	"github.com/anhcraft/rice/exec/ast/opr"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
)

func binaryOp(op Opcode, left, right types.Value) (types.Value, error) {
	var isLeftPrimitive, isRightPrimitive bool
	if _, ok := left.(values.Primitive); ok {
		isLeftPrimitive = true
	}
	if _, ok := right.(values.Primitive); ok {
		isRightPrimitive = true
	}

	if left == nil || right == nil || !isLeftPrimitive || !isRightPrimitive {
		if op == OpEqual {
			return values.Bool(left == right), nil
		}
		return nil, fmt.Errorf("cannot eval %T %s %T", left, opSymbol(op), right)
	}

	_, leftWasStr := left.(values.String)
	_, rightWasStr := right.(values.String)

	cl, cr, err := values.ConvertPrimitiveImplicitly(left.(values.Primitive), right.(values.Primitive))
	if err != nil {
		return nil, fmt.Errorf("failed implicit type conversion of %T and %T: %w", left, right, err)
	}
	left, right = cl, cr

	if op == OpEqual {
		return values.Bool(left == right), nil
	}

	switch left.(type) {
	case values.Int:
		a := left.(values.Int)
		b := right.(values.Int)
		switch op {
		case OpAdd:
			return a + b, nil
		case OpSubtract:
			return a - b, nil
		case OpMultiply:
			return a * b, nil
		case OpDivide:
			if b == 0 {
				return nil, fmt.Errorf("integer division by zero")
			}
			return a / b, nil
		case OpModulo:
			if b == 0 {
				return nil, fmt.Errorf("integer remainder by zero")
			}
			return a % b, nil
		case OpLess:
			return values.Bool(a < b), nil
		case OpMore:
			return values.Bool(a > b), nil
		case OpLessOrEqual:
			return values.Bool(a <= b), nil
		case OpMoreOrEqual:
			return values.Bool(a >= b), nil
		case OpAnd, OpOr:
			return nil, fmt.Errorf("cannot eval %T %s %T", left, opSymbol(op), right)
		}
	case values.Float:
		a := left.(values.Float)
		b := right.(values.Float)
		switch op {
		case OpAdd:
			return a + b, nil
		case OpSubtract:
			return a - b, nil
		case OpMultiply:
			return a * b, nil
		case OpDivide:
			return a / b, nil
		case OpModulo:
			return values.Float(math.Mod(float64(a), float64(b))), nil
		case OpLess:
			return values.Bool(a < b), nil
		case OpMore:
			return values.Bool(a > b), nil
		case OpLessOrEqual:
			return values.Bool(a <= b), nil
		case OpMoreOrEqual:
			return values.Bool(a >= b), nil
		case OpAnd, OpOr:
			return nil, fmt.Errorf("cannot eval %T %s %T", left, opSymbol(op), right)
		}
	case values.Bool:
		a := left.(values.Bool)
		b := right.(values.Bool)
		switch op {
		case OpAnd:
			return a && b, nil
		case OpOr:
			return a || b, nil
		}
	case values.String:
		if !leftWasStr || !rightWasStr {
			switch op {
			case OpLess, OpMore, OpLessOrEqual, OpMoreOrEqual:
				return nil, fmt.Errorf("cannot eval %T %s %T", left, opSymbol(op), right)
			}
		}
		a := left.(values.String)
		b := right.(values.String)
		switch op {
		case OpAdd:
			return a + b, nil
		case OpLess:
			return values.Bool(a < b), nil
		case OpMore:
			return values.Bool(a > b), nil
		case OpLessOrEqual:
			return values.Bool(a <= b), nil
		case OpMoreOrEqual:
			return values.Bool(a >= b), nil
		}
	}
	return nil, fmt.Errorf("cannot eval %T %s %T", left, opSymbol(op), right)
}

func unaryNegate(v types.Value) (types.Value, error) {
	switch n := v.(type) {
	case values.Int:
		return -n, nil
	case values.Float:
		return -n, nil
	default:
		return nil, fmt.Errorf("cannot match operator %q to operand of type %T", opr.Neg, v)
	}
}

func unaryNot(v types.Value) (types.Value, error) {
	if b, ok := v.(values.Bool); ok {
		return !b, nil
	}
	return nil, fmt.Errorf("cannot match operator %q to operand of type %T", opr.Inv, v)
}

func increment(v types.Value, delta values.Int) (types.Value, error) {
	switch n := v.(type) {
	case values.Int:
		return n + delta, nil
	case values.Float:
		return n + values.Float(delta), nil
	default:
		return nil, fmt.Errorf("value of type %T is not numeric", v)
	}
}

func opSymbol(op Opcode) string {
	switch op {
	case OpAdd:
		return "+"
	case OpSubtract:
		return "-"
	case OpMultiply:
		return "*"
	case OpDivide:
		return "/"
	case OpModulo:
		return "%"
	case OpEqual:
		return "=="
	case OpLess:
		return "<"
	case OpMore:
		return ">"
	case OpLessOrEqual:
		return "<="
	case OpMoreOrEqual:
		return ">="
	case OpAnd:
		return "&&"
	case OpOr:
		return "||"
	default:
		return op.String()
	}
}
